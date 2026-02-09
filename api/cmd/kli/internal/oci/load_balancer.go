package oci

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/oracle/oci-go-sdk/v65/core"
	"github.com/oracle/oci-go-sdk/v65/networkloadbalancer"
)

// NLBName returns the NLB name for an installation
func NLBName(installationKey string) string {
	return fmt.Sprintf("kl-%s-nlb", shortKey(installationKey))
}

// ReservedIPName returns the reserved IP name for an installation
func ReservedIPName(installationKey string) string {
	return fmt.Sprintf("kl-%s-ip", shortKey(installationKey))
}

// BackendSetName is the name used for the NLB backend set
const BackendSetName = "kl-backend"

// ReservePublicIP reserves a public IP for the NLB
func ReservePublicIP(ctx context.Context, cfg *OCIConfig, installationKey string) (string, error) {
	vnClient, err := core.NewVirtualNetworkClientWithConfigurationProvider(cfg.ConfigProvider)
	if err != nil {
		return "", fmt.Errorf("failed to create virtual network client: %w", err)
	}

	ipName := ReservedIPName(installationKey)

	// Check if IP already exists
	existingID, existingIP, err := findReservedIPByName(ctx, vnClient, cfg.CompartmentOCID, ipName)
	if err == nil && existingID != "" {
		return existingIP, nil
	}

	tags := freeformTags(installationKey)
	lifetime := core.CreatePublicIpDetailsLifetimeReserved
	resp, err := vnClient.CreatePublicIp(ctx, core.CreatePublicIpRequest{
		CreatePublicIpDetails: core.CreatePublicIpDetails{
			CompartmentId: &cfg.CompartmentOCID,
			DisplayName:   &ipName,
			Lifetime:      lifetime,
			FreeformTags:  tags,
		},
	})
	if err != nil {
		if strings.Contains(err.Error(), "already exists") || strings.Contains(err.Error(), "Conflict") {
			existingID, existingIP, findErr := findReservedIPByName(ctx, vnClient, cfg.CompartmentOCID, ipName)
			if findErr == nil && existingID != "" {
				return existingIP, nil
			}
		}
		return "", fmt.Errorf("failed to reserve public IP: %w", err)
	}

	if resp.IpAddress != nil {
		return *resp.IpAddress, nil
	}

	// Wait for IP to be assigned
	if resp.Id != nil {
		return waitForPublicIP(ctx, vnClient, *resp.Id)
	}

	return "", fmt.Errorf("failed to get reserved IP address")
}

// CreateNetworkLoadBalancer creates an OCI Network Load Balancer
func CreateNetworkLoadBalancer(ctx context.Context, cfg *OCIConfig, subnetID, instanceID, instanceIP, installationKey string) (string, string, error) {
	nlbClient, err := networkloadbalancer.NewNetworkLoadBalancerClientWithConfigurationProvider(cfg.ConfigProvider)
	if err != nil {
		return "", "", fmt.Errorf("failed to create NLB client: %w", err)
	}

	nlbName := NLBName(installationKey)

	// Check if NLB already exists
	existingID, existingIP, findErr := findNLBByName(ctx, nlbClient, cfg.CompartmentOCID, nlbName)
	if findErr == nil && existingID != "" {
		return existingID, existingIP, nil
	}

	tags := freeformTags(installationKey)

	// Health check config (TCP for Layer 4 NLB)
	healthCheckPort := 80
	healthCheckInterval := 5000 // 5 seconds in ms
	healthCheckTimeout := 3000  // 3 seconds in ms
	healthCheckRetries := 2

	// Backend set with the instance
	backendPort := 80
	backendSets := map[string]networkloadbalancer.BackendSetDetails{
		BackendSetName: {
			Policy: networkloadbalancer.NetworkLoadBalancingPolicyFiveTuple,
			HealthChecker: &networkloadbalancer.HealthChecker{
				Protocol:         networkloadbalancer.HealthCheckProtocolsTcp,
				Port:             &healthCheckPort,
				IntervalInMillis: &healthCheckInterval,
				TimeoutInMillis:  &healthCheckTimeout,
				Retries:          &healthCheckRetries,
			},
			Backends: []networkloadbalancer.Backend{
				{
					IpAddress: &instanceIP,
					Port:      &backendPort,
				},
			},
		},
	}

	// Listener on port 80
	listenerPort := 80
	listenerName := "kl-http"
	listeners := map[string]networkloadbalancer.ListenerDetails{
		listenerName: {
			Name:                  &listenerName,
			DefaultBackendSetName: strPtr(BackendSetName),
			Port:                  &listenerPort,
			Protocol:              networkloadbalancer.ListenerProtocolsTcp,
		},
	}

	isPrivate := false
	resp, err := nlbClient.CreateNetworkLoadBalancer(ctx, networkloadbalancer.CreateNetworkLoadBalancerRequest{
		CreateNetworkLoadBalancerDetails: networkloadbalancer.CreateNetworkLoadBalancerDetails{
			CompartmentId: &cfg.CompartmentOCID,
			DisplayName:   &nlbName,
			SubnetId:      &subnetID,
			IsPrivate:     &isPrivate,
			BackendSets:   backendSets,
			Listeners:     listeners,
			FreeformTags:  tags,
		},
	})
	if err != nil {
		if strings.Contains(err.Error(), "already exists") || strings.Contains(err.Error(), "Conflict") {
			existingID, existingIP, findErr := findNLBByName(ctx, nlbClient, cfg.CompartmentOCID, nlbName)
			if findErr == nil && existingID != "" {
				return existingID, existingIP, nil
			}
		}
		return "", "", fmt.Errorf("failed to create NLB: %w", err)
	}

	nlbID := ""
	if resp.Id != nil {
		nlbID = *resp.Id
	}

	// Wait for NLB to become active
	ip, err := WaitForNLBActive(ctx, cfg, nlbID)
	if err != nil {
		return nlbID, "", fmt.Errorf("NLB created but failed waiting for active state: %w", err)
	}

	return nlbID, ip, nil
}

// WaitForNLBActive waits for the NLB to become active and returns its IP
func WaitForNLBActive(ctx context.Context, cfg *OCIConfig, nlbID string) (string, error) {
	nlbClient, err := networkloadbalancer.NewNetworkLoadBalancerClientWithConfigurationProvider(cfg.ConfigProvider)
	if err != nil {
		return "", fmt.Errorf("failed to create NLB client: %w", err)
	}

	deadline := time.Now().Add(10 * time.Minute)
	for time.Now().Before(deadline) {
		resp, err := nlbClient.GetNetworkLoadBalancer(ctx, networkloadbalancer.GetNetworkLoadBalancerRequest{
			NetworkLoadBalancerId: &nlbID,
		})
		if err != nil {
			time.Sleep(10 * time.Second)
			continue
		}

		if resp.LifecycleState == networkloadbalancer.LifecycleStateActive {
			// Get the public IP
			if len(resp.IpAddresses) > 0 {
				for _, ipAddr := range resp.IpAddresses {
					if ipAddr.IpAddress != nil && (ipAddr.IsPublic == nil || *ipAddr.IsPublic) {
						return *ipAddr.IpAddress, nil
					}
				}
			}
			return "", nil
		}

		if resp.LifecycleState == networkloadbalancer.LifecycleStateFailed {
			return "", fmt.Errorf("NLB entered FAILED state")
		}

		select {
		case <-ctx.Done():
			return "", ctx.Err()
		case <-time.After(10 * time.Second):
			// Continue polling
		}
	}

	return "", fmt.Errorf("NLB did not become active within timeout")
}

// GetNLBIP returns the public IP of the NLB
func GetNLBIP(ctx context.Context, cfg *OCIConfig, installationKey string) (string, error) {
	nlbClient, err := networkloadbalancer.NewNetworkLoadBalancerClientWithConfigurationProvider(cfg.ConfigProvider)
	if err != nil {
		return "", fmt.Errorf("failed to create NLB client: %w", err)
	}

	nlbName := NLBName(installationKey)
	nlbID, ip, err := findNLBByName(ctx, nlbClient, cfg.CompartmentOCID, nlbName)
	if err != nil {
		return "", err
	}
	if ip != "" {
		return ip, nil
	}

	// Get NLB details
	resp, err := nlbClient.GetNetworkLoadBalancer(ctx, networkloadbalancer.GetNetworkLoadBalancerRequest{
		NetworkLoadBalancerId: &nlbID,
	})
	if err != nil {
		return "", fmt.Errorf("failed to get NLB: %w", err)
	}

	for _, ipAddr := range resp.IpAddresses {
		if ipAddr.IpAddress != nil && (ipAddr.IsPublic == nil || *ipAddr.IsPublic) {
			return *ipAddr.IpAddress, nil
		}
	}

	return "", fmt.Errorf("no public IP found for NLB")
}

// DeleteNetworkLoadBalancer deletes the NLB
func DeleteNetworkLoadBalancer(ctx context.Context, cfg *OCIConfig, installationKey string) error {
	nlbClient, err := networkloadbalancer.NewNetworkLoadBalancerClientWithConfigurationProvider(cfg.ConfigProvider)
	if err != nil {
		return fmt.Errorf("failed to create NLB client: %w", err)
	}

	nlbName := NLBName(installationKey)
	nlbID, _, err := findNLBByName(ctx, nlbClient, cfg.CompartmentOCID, nlbName)
	if err != nil || nlbID == "" {
		return nil // Already gone
	}

	_, err = nlbClient.DeleteNetworkLoadBalancer(ctx, networkloadbalancer.DeleteNetworkLoadBalancerRequest{
		NetworkLoadBalancerId: &nlbID,
	})
	if err != nil {
		if strings.Contains(err.Error(), "404") || strings.Contains(err.Error(), "not found") {
			return nil
		}
		return fmt.Errorf("failed to delete NLB: %w", err)
	}

	// Wait for NLB deletion
	deadline := time.Now().Add(5 * time.Minute)
	for time.Now().Before(deadline) {
		_, err := nlbClient.GetNetworkLoadBalancer(ctx, networkloadbalancer.GetNetworkLoadBalancerRequest{
			NetworkLoadBalancerId: &nlbID,
		})
		if err != nil {
			// NLB is gone
			return nil
		}
		time.Sleep(5 * time.Second)
	}

	return nil
}

// DeleteReservedPublicIP deletes the reserved public IP
func DeleteReservedPublicIP(ctx context.Context, cfg *OCIConfig, installationKey string) error {
	vnClient, err := core.NewVirtualNetworkClientWithConfigurationProvider(cfg.ConfigProvider)
	if err != nil {
		return fmt.Errorf("failed to create virtual network client: %w", err)
	}

	ipName := ReservedIPName(installationKey)
	ipID, _, err := findReservedIPByName(ctx, vnClient, cfg.CompartmentOCID, ipName)
	if err != nil || ipID == "" {
		return nil // Already gone
	}

	_, err = vnClient.DeletePublicIp(ctx, core.DeletePublicIpRequest{
		PublicIpId: &ipID,
	})
	if err != nil {
		if strings.Contains(err.Error(), "404") || strings.Contains(err.Error(), "not found") {
			return nil
		}
		return fmt.Errorf("failed to delete reserved IP: %w", err)
	}

	return nil
}

// findNLBByName finds an NLB by display name
func findNLBByName(ctx context.Context, nlbClient networkloadbalancer.NetworkLoadBalancerClient, compartmentID, name string) (string, string, error) {
	resp, err := nlbClient.ListNetworkLoadBalancers(ctx, networkloadbalancer.ListNetworkLoadBalancersRequest{
		CompartmentId: &compartmentID,
		DisplayName:   &name,
	})
	if err != nil {
		return "", "", err
	}

	for _, nlb := range resp.Items {
		if nlb.Id == nil {
			continue
		}
		if nlb.LifecycleState == networkloadbalancer.LifecycleStateActive ||
			nlb.LifecycleState == networkloadbalancer.LifecycleStateCreating {
			ip := ""
			for _, ipAddr := range nlb.IpAddresses {
				if ipAddr.IpAddress != nil && (ipAddr.IsPublic == nil || *ipAddr.IsPublic) {
					ip = *ipAddr.IpAddress
					break
				}
			}
			return *nlb.Id, ip, nil
		}
	}

	return "", "", fmt.Errorf("NLB %s not found", name)
}

// findReservedIPByName finds a reserved public IP by display name
func findReservedIPByName(ctx context.Context, vnClient core.VirtualNetworkClient, compartmentID, name string) (string, string, error) {
	lifetime := core.ListPublicIpsLifetimeReserved
	scope := core.ListPublicIpsScopeRegion
	resp, err := vnClient.ListPublicIps(ctx, core.ListPublicIpsRequest{
		CompartmentId: &compartmentID,
		Lifetime:      lifetime,
		Scope:         scope,
	})
	if err != nil {
		return "", "", err
	}

	for _, ip := range resp.Items {
		if ip.DisplayName != nil && *ip.DisplayName == name {
			if ip.LifecycleState == core.PublicIpLifecycleStateAvailable || ip.LifecycleState == core.PublicIpLifecycleStateAssigned {
				ipAddr := ""
				if ip.IpAddress != nil {
					ipAddr = *ip.IpAddress
				}
				return *ip.Id, ipAddr, nil
			}
		}
	}

	return "", "", fmt.Errorf("reserved IP %s not found", name)
}

// waitForPublicIP waits for a public IP to be assigned
func waitForPublicIP(ctx context.Context, vnClient core.VirtualNetworkClient, ipID string) (string, error) {
	deadline := time.Now().Add(2 * time.Minute)
	for time.Now().Before(deadline) {
		resp, err := vnClient.GetPublicIp(ctx, core.GetPublicIpRequest{
			PublicIpId: &ipID,
		})
		if err != nil {
			time.Sleep(2 * time.Second)
			continue
		}

		if resp.IpAddress != nil && *resp.IpAddress != "" {
			return *resp.IpAddress, nil
		}

		time.Sleep(2 * time.Second)
	}

	return "", fmt.Errorf("timeout waiting for public IP assignment")
}
