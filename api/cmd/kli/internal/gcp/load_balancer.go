package gcp

import (
	"context"
	"fmt"
	"strings"
	"time"

	compute "cloud.google.com/go/compute/apiv1"
	"cloud.google.com/go/compute/apiv1/computepb"
)

// LBInfo contains information about the created load balancer
type LBInfo struct {
	IP             string
	ForwardingRule string
}

// shortKey returns the first 8 characters of an installation key for resource naming
func shortKey(installationKey string) string {
	if len(installationKey) > 8 {
		return installationKey[:8]
	}
	return installationKey
}

// ReserveExternalIP reserves a global static IP address for the load balancer
func ReserveExternalIP(ctx context.Context, cfg *GCPConfig, installationKey string) (string, error) {
	addressesClient, err := compute.NewGlobalAddressesRESTClient(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to create addresses client: %w", err)
	}
	defer addressesClient.Close()

	addressName := fmt.Sprintf("kl-%s-ip", shortKey(installationKey))

	// Check if address already exists
	existing, err := addressesClient.Get(ctx, &computepb.GetGlobalAddressRequest{
		Project: cfg.Project,
		Address: addressName,
	})
	if err == nil && existing.Address != nil {
		return *existing.Address, nil
	}

	address := &computepb.Address{
		Name:        ptrString(addressName),
		Description: ptrString("Kloudlite load balancer IP"),
		IpVersion:   ptrString("IPV4"),
	}

	op, err := addressesClient.Insert(ctx, &computepb.InsertGlobalAddressRequest{
		Project:         cfg.Project,
		AddressResource: address,
	})
	if err != nil {
		if strings.Contains(err.Error(), "already exists") {
			// Get the existing address
			existing, getErr := addressesClient.Get(ctx, &computepb.GetGlobalAddressRequest{
				Project: cfg.Project,
				Address: addressName,
			})
			if getErr == nil && existing.Address != nil {
				return *existing.Address, nil
			}
		}
		return "", fmt.Errorf("failed to reserve IP address: %w", err)
	}

	if err := op.Wait(ctx); err != nil {
		return "", fmt.Errorf("failed waiting for IP reservation: %w", err)
	}

	// Get the reserved address
	result, err := addressesClient.Get(ctx, &computepb.GetGlobalAddressRequest{
		Project: cfg.Project,
		Address: addressName,
	})
	if err != nil {
		return "", fmt.Errorf("failed to get reserved IP: %w", err)
	}

	return *result.Address, nil
}

// CreateHealthCheck creates an HTTP health check for the backend service
func CreateHealthCheck(ctx context.Context, cfg *GCPConfig, installationKey string) (string, error) {
	healthChecksClient, err := compute.NewHealthChecksRESTClient(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to create health checks client: %w", err)
	}
	defer healthChecksClient.Close()

	hcName := fmt.Sprintf("kl-%s-hc", shortKey(installationKey))

	// Check if health check already exists
	existing, err := healthChecksClient.Get(ctx, &computepb.GetHealthCheckRequest{
		Project:     cfg.Project,
		HealthCheck: hcName,
	})
	if err == nil && existing.SelfLink != nil {
		return *existing.SelfLink, nil
	}

	healthCheck := &computepb.HealthCheck{
		Name:               ptrString(hcName),
		Description:        ptrString("Kloudlite health check"),
		CheckIntervalSec:   ptrInt32(30),
		TimeoutSec:         ptrInt32(5),
		HealthyThreshold:   ptrInt32(2),
		UnhealthyThreshold: ptrInt32(3),
		Type:               ptrString("HTTP"),
		HttpHealthCheck: &computepb.HTTPHealthCheck{
			Port:        ptrInt32(80),
			RequestPath: ptrString("/"),
		},
	}

	op, err := healthChecksClient.Insert(ctx, &computepb.InsertHealthCheckRequest{
		Project:             cfg.Project,
		HealthCheckResource: healthCheck,
	})
	if err != nil {
		if strings.Contains(err.Error(), "already exists") {
			existing, getErr := healthChecksClient.Get(ctx, &computepb.GetHealthCheckRequest{
				Project:     cfg.Project,
				HealthCheck: hcName,
			})
			if getErr == nil && existing.SelfLink != nil {
				return *existing.SelfLink, nil
			}
		}
		return "", fmt.Errorf("failed to create health check: %w", err)
	}

	if err := op.Wait(ctx); err != nil {
		return "", fmt.Errorf("failed waiting for health check creation: %w", err)
	}

	// Get the created health check
	result, err := healthChecksClient.Get(ctx, &computepb.GetHealthCheckRequest{
		Project:     cfg.Project,
		HealthCheck: hcName,
	})
	if err != nil {
		return "", fmt.Errorf("failed to get health check: %w", err)
	}

	return *result.SelfLink, nil
}

// CreateUnmanagedInstanceGroup creates an unmanaged instance group and adds the VM
func CreateUnmanagedInstanceGroup(ctx context.Context, cfg *GCPConfig, instanceName, installationKey string) (string, error) {
	instanceGroupsClient, err := compute.NewInstanceGroupsRESTClient(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to create instance groups client: %w", err)
	}
	defer instanceGroupsClient.Close()

	igName := fmt.Sprintf("kl-%s-ig", shortKey(installationKey))

	// Check if instance group already exists
	existing, err := instanceGroupsClient.Get(ctx, &computepb.GetInstanceGroupRequest{
		Project:       cfg.Project,
		Zone:          cfg.Zone,
		InstanceGroup: igName,
	})
	if err == nil && existing.SelfLink != nil {
		return *existing.SelfLink, nil
	}

	instanceGroup := &computepb.InstanceGroup{
		Name:        ptrString(igName),
		Description: ptrString("Kloudlite instance group"),
	}

	op, err := instanceGroupsClient.Insert(ctx, &computepb.InsertInstanceGroupRequest{
		Project:               cfg.Project,
		Zone:                  cfg.Zone,
		InstanceGroupResource: instanceGroup,
	})
	if err != nil {
		if strings.Contains(err.Error(), "already exists") {
			existing, getErr := instanceGroupsClient.Get(ctx, &computepb.GetInstanceGroupRequest{
				Project:       cfg.Project,
				Zone:          cfg.Zone,
				InstanceGroup: igName,
			})
			if getErr == nil && existing.SelfLink != nil {
				return *existing.SelfLink, nil
			}
		}
		return "", fmt.Errorf("failed to create instance group: %w", err)
	}

	if err := op.Wait(ctx); err != nil {
		return "", fmt.Errorf("failed waiting for instance group creation: %w", err)
	}

	// Add instance to the group
	instanceURL := fmt.Sprintf("projects/%s/zones/%s/instances/%s", cfg.Project, cfg.Zone, instanceName)
	addOp, err := instanceGroupsClient.AddInstances(ctx, &computepb.AddInstancesInstanceGroupRequest{
		Project:       cfg.Project,
		Zone:          cfg.Zone,
		InstanceGroup: igName,
		InstanceGroupsAddInstancesRequestResource: &computepb.InstanceGroupsAddInstancesRequest{
			Instances: []*computepb.InstanceReference{
				{Instance: ptrString(instanceURL)},
			},
		},
	})
	if err != nil {
		if !strings.Contains(err.Error(), "already a member") {
			return "", fmt.Errorf("failed to add instance to group: %w", err)
		}
	} else {
		if err := addOp.Wait(ctx); err != nil {
			return "", fmt.Errorf("failed waiting for instance addition: %w", err)
		}
	}

	// Get the created instance group
	result, err := instanceGroupsClient.Get(ctx, &computepb.GetInstanceGroupRequest{
		Project:       cfg.Project,
		Zone:          cfg.Zone,
		InstanceGroup: igName,
	})
	if err != nil {
		return "", fmt.Errorf("failed to get instance group: %w", err)
	}

	return *result.SelfLink, nil
}

// CreateBackendService creates a backend service with the health check and instance group
func CreateBackendService(ctx context.Context, cfg *GCPConfig, igURL, hcURL, installationKey string) (string, error) {
	backendServicesClient, err := compute.NewBackendServicesRESTClient(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to create backend services client: %w", err)
	}
	defer backendServicesClient.Close()

	bsName := fmt.Sprintf("kl-%s-backend", shortKey(installationKey))

	// Check if backend service already exists
	existing, err := backendServicesClient.Get(ctx, &computepb.GetBackendServiceRequest{
		Project:        cfg.Project,
		BackendService: bsName,
	})
	if err == nil && existing.SelfLink != nil {
		return *existing.SelfLink, nil
	}

	backendService := &computepb.BackendService{
		Name:                ptrString(bsName),
		Description:         ptrString("Kloudlite backend service"),
		Protocol:            ptrString("HTTP"),
		PortName:            ptrString("http"),
		TimeoutSec:          ptrInt32(30),
		LoadBalancingScheme: ptrString("EXTERNAL"),
		HealthChecks:        []string{hcURL},
		Backends: []*computepb.Backend{
			{
				Group:          ptrString(igURL),
				BalancingMode:  ptrString("UTILIZATION"),
				MaxUtilization: ptrFloat32(0.8),
			},
		},
	}

	op, err := backendServicesClient.Insert(ctx, &computepb.InsertBackendServiceRequest{
		Project:                cfg.Project,
		BackendServiceResource: backendService,
	})
	if err != nil {
		if strings.Contains(err.Error(), "already exists") {
			existing, getErr := backendServicesClient.Get(ctx, &computepb.GetBackendServiceRequest{
				Project:        cfg.Project,
				BackendService: bsName,
			})
			if getErr == nil && existing.SelfLink != nil {
				return *existing.SelfLink, nil
			}
		}
		return "", fmt.Errorf("failed to create backend service: %w", err)
	}

	if err := op.Wait(ctx); err != nil {
		return "", fmt.Errorf("failed waiting for backend service creation: %w", err)
	}

	// Get the created backend service
	result, err := backendServicesClient.Get(ctx, &computepb.GetBackendServiceRequest{
		Project:        cfg.Project,
		BackendService: bsName,
	})
	if err != nil {
		return "", fmt.Errorf("failed to get backend service: %w", err)
	}

	return *result.SelfLink, nil
}

// CreateURLMap creates a URL map pointing to the backend service
func CreateURLMap(ctx context.Context, cfg *GCPConfig, backendURL, installationKey string) (string, error) {
	urlMapsClient, err := compute.NewUrlMapsRESTClient(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to create URL maps client: %w", err)
	}
	defer urlMapsClient.Close()

	urlMapName := fmt.Sprintf("kl-%s-urlmap", shortKey(installationKey))

	// Check if URL map already exists
	existing, err := urlMapsClient.Get(ctx, &computepb.GetUrlMapRequest{
		Project: cfg.Project,
		UrlMap:  urlMapName,
	})
	if err == nil && existing.SelfLink != nil {
		return *existing.SelfLink, nil
	}

	urlMap := &computepb.UrlMap{
		Name:           ptrString(urlMapName),
		Description:    ptrString("Kloudlite URL map"),
		DefaultService: ptrString(backendURL),
	}

	op, err := urlMapsClient.Insert(ctx, &computepb.InsertUrlMapRequest{
		Project:        cfg.Project,
		UrlMapResource: urlMap,
	})
	if err != nil {
		if strings.Contains(err.Error(), "already exists") {
			existing, getErr := urlMapsClient.Get(ctx, &computepb.GetUrlMapRequest{
				Project: cfg.Project,
				UrlMap:  urlMapName,
			})
			if getErr == nil && existing.SelfLink != nil {
				return *existing.SelfLink, nil
			}
		}
		return "", fmt.Errorf("failed to create URL map: %w", err)
	}

	if err := op.Wait(ctx); err != nil {
		return "", fmt.Errorf("failed waiting for URL map creation: %w", err)
	}

	// Get the created URL map
	result, err := urlMapsClient.Get(ctx, &computepb.GetUrlMapRequest{
		Project: cfg.Project,
		UrlMap:  urlMapName,
	})
	if err != nil {
		return "", fmt.Errorf("failed to get URL map: %w", err)
	}

	return *result.SelfLink, nil
}

// CreateTargetHTTPProxy creates a target HTTP proxy
func CreateTargetHTTPProxy(ctx context.Context, cfg *GCPConfig, urlMapURL, installationKey string) (string, error) {
	targetHttpProxiesClient, err := compute.NewTargetHttpProxiesRESTClient(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to create target HTTP proxies client: %w", err)
	}
	defer targetHttpProxiesClient.Close()

	proxyName := fmt.Sprintf("kl-%s-proxy", shortKey(installationKey))

	// Check if proxy already exists
	existing, err := targetHttpProxiesClient.Get(ctx, &computepb.GetTargetHttpProxyRequest{
		Project:         cfg.Project,
		TargetHttpProxy: proxyName,
	})
	if err == nil && existing.SelfLink != nil {
		return *existing.SelfLink, nil
	}

	proxy := &computepb.TargetHttpProxy{
		Name:        ptrString(proxyName),
		Description: ptrString("Kloudlite HTTP proxy"),
		UrlMap:      ptrString(urlMapURL),
	}

	op, err := targetHttpProxiesClient.Insert(ctx, &computepb.InsertTargetHttpProxyRequest{
		Project:                 cfg.Project,
		TargetHttpProxyResource: proxy,
	})
	if err != nil {
		if strings.Contains(err.Error(), "already exists") {
			existing, getErr := targetHttpProxiesClient.Get(ctx, &computepb.GetTargetHttpProxyRequest{
				Project:         cfg.Project,
				TargetHttpProxy: proxyName,
			})
			if getErr == nil && existing.SelfLink != nil {
				return *existing.SelfLink, nil
			}
		}
		return "", fmt.Errorf("failed to create target HTTP proxy: %w", err)
	}

	if err := op.Wait(ctx); err != nil {
		return "", fmt.Errorf("failed waiting for target HTTP proxy creation: %w", err)
	}

	// Get the created proxy
	result, err := targetHttpProxiesClient.Get(ctx, &computepb.GetTargetHttpProxyRequest{
		Project:         cfg.Project,
		TargetHttpProxy: proxyName,
	})
	if err != nil {
		return "", fmt.Errorf("failed to get target HTTP proxy: %w", err)
	}

	return *result.SelfLink, nil
}

// CreateGlobalForwardingRule creates a global forwarding rule on port 80
func CreateGlobalForwardingRule(ctx context.Context, cfg *GCPConfig, ipAddress, proxyURL, installationKey string) error {
	forwardingRulesClient, err := compute.NewGlobalForwardingRulesRESTClient(ctx)
	if err != nil {
		return fmt.Errorf("failed to create forwarding rules client: %w", err)
	}
	defer forwardingRulesClient.Close()

	fwdName := fmt.Sprintf("kl-%s-fwd", shortKey(installationKey))

	// Check if forwarding rule already exists
	_, err = forwardingRulesClient.Get(ctx, &computepb.GetGlobalForwardingRuleRequest{
		Project:        cfg.Project,
		ForwardingRule: fwdName,
	})
	if err == nil {
		return nil // Already exists
	}

	forwardingRule := &computepb.ForwardingRule{
		Name:                ptrString(fwdName),
		Description:         ptrString("Kloudlite forwarding rule"),
		IPAddress:           ptrString(ipAddress),
		IPProtocol:          ptrString("TCP"),
		PortRange:           ptrString("80"),
		Target:              ptrString(proxyURL),
		LoadBalancingScheme: ptrString("EXTERNAL"),
	}

	op, err := forwardingRulesClient.Insert(ctx, &computepb.InsertGlobalForwardingRuleRequest{
		Project:                cfg.Project,
		ForwardingRuleResource: forwardingRule,
	})
	if err != nil {
		if strings.Contains(err.Error(), "already exists") {
			return nil
		}
		return fmt.Errorf("failed to create forwarding rule: %w", err)
	}

	if err := op.Wait(ctx); err != nil {
		return fmt.Errorf("failed waiting for forwarding rule creation: %w", err)
	}

	return nil
}

// DeleteLoadBalancer deletes all load balancer components in reverse order
func DeleteLoadBalancer(ctx context.Context, cfg *GCPConfig, installationKey string) error {
	key := shortKey(installationKey)

	// Delete in reverse dependency order
	// 1. Forwarding Rule
	if err := deleteForwardingRule(ctx, cfg, fmt.Sprintf("kl-%s-fwd", key)); err != nil {
		fmt.Printf("Warning: Failed to delete forwarding rule: %v\n", err)
	}

	// 2. Target HTTP Proxy
	if err := deleteTargetHTTPProxy(ctx, cfg, fmt.Sprintf("kl-%s-proxy", key)); err != nil {
		fmt.Printf("Warning: Failed to delete target HTTP proxy: %v\n", err)
	}

	// 3. URL Map
	if err := deleteURLMap(ctx, cfg, fmt.Sprintf("kl-%s-urlmap", key)); err != nil {
		fmt.Printf("Warning: Failed to delete URL map: %v\n", err)
	}

	// 4. Backend Service
	if err := deleteBackendService(ctx, cfg, fmt.Sprintf("kl-%s-backend", key)); err != nil {
		fmt.Printf("Warning: Failed to delete backend service: %v\n", err)
	}

	// 5. Instance Group
	if err := deleteInstanceGroup(ctx, cfg, fmt.Sprintf("kl-%s-ig", key)); err != nil {
		fmt.Printf("Warning: Failed to delete instance group: %v\n", err)
	}

	// 6. Health Check
	if err := deleteHealthCheck(ctx, cfg, fmt.Sprintf("kl-%s-hc", key)); err != nil {
		fmt.Printf("Warning: Failed to delete health check: %v\n", err)
	}

	// 7. Reserved IP
	if err := deleteGlobalAddress(ctx, cfg, fmt.Sprintf("kl-%s-ip", key)); err != nil {
		fmt.Printf("Warning: Failed to delete reserved IP: %v\n", err)
	}

	return nil
}

func deleteForwardingRule(ctx context.Context, cfg *GCPConfig, name string) error {
	client, err := compute.NewGlobalForwardingRulesRESTClient(ctx)
	if err != nil {
		return err
	}
	defer client.Close()

	op, err := client.Delete(ctx, &computepb.DeleteGlobalForwardingRuleRequest{
		Project:        cfg.Project,
		ForwardingRule: name,
	})
	if err != nil {
		if strings.Contains(err.Error(), "not found") || strings.Contains(err.Error(), "404") {
			return nil
		}
		return err
	}
	return op.Wait(ctx)
}

func deleteTargetHTTPProxy(ctx context.Context, cfg *GCPConfig, name string) error {
	client, err := compute.NewTargetHttpProxiesRESTClient(ctx)
	if err != nil {
		return err
	}
	defer client.Close()

	op, err := client.Delete(ctx, &computepb.DeleteTargetHttpProxyRequest{
		Project:         cfg.Project,
		TargetHttpProxy: name,
	})
	if err != nil {
		if strings.Contains(err.Error(), "not found") || strings.Contains(err.Error(), "404") {
			return nil
		}
		return err
	}
	return op.Wait(ctx)
}

func deleteURLMap(ctx context.Context, cfg *GCPConfig, name string) error {
	client, err := compute.NewUrlMapsRESTClient(ctx)
	if err != nil {
		return err
	}
	defer client.Close()

	op, err := client.Delete(ctx, &computepb.DeleteUrlMapRequest{
		Project: cfg.Project,
		UrlMap:  name,
	})
	if err != nil {
		if strings.Contains(err.Error(), "not found") || strings.Contains(err.Error(), "404") {
			return nil
		}
		return err
	}
	return op.Wait(ctx)
}

func deleteBackendService(ctx context.Context, cfg *GCPConfig, name string) error {
	client, err := compute.NewBackendServicesRESTClient(ctx)
	if err != nil {
		return err
	}
	defer client.Close()

	op, err := client.Delete(ctx, &computepb.DeleteBackendServiceRequest{
		Project:        cfg.Project,
		BackendService: name,
	})
	if err != nil {
		if strings.Contains(err.Error(), "not found") || strings.Contains(err.Error(), "404") {
			return nil
		}
		return err
	}
	return op.Wait(ctx)
}

func deleteInstanceGroup(ctx context.Context, cfg *GCPConfig, name string) error {
	client, err := compute.NewInstanceGroupsRESTClient(ctx)
	if err != nil {
		return err
	}
	defer client.Close()

	op, err := client.Delete(ctx, &computepb.DeleteInstanceGroupRequest{
		Project:       cfg.Project,
		Zone:          cfg.Zone,
		InstanceGroup: name,
	})
	if err != nil {
		if strings.Contains(err.Error(), "not found") || strings.Contains(err.Error(), "404") {
			return nil
		}
		return err
	}
	return op.Wait(ctx)
}

func deleteHealthCheck(ctx context.Context, cfg *GCPConfig, name string) error {
	client, err := compute.NewHealthChecksRESTClient(ctx)
	if err != nil {
		return err
	}
	defer client.Close()

	op, err := client.Delete(ctx, &computepb.DeleteHealthCheckRequest{
		Project:     cfg.Project,
		HealthCheck: name,
	})
	if err != nil {
		if strings.Contains(err.Error(), "not found") || strings.Contains(err.Error(), "404") {
			return nil
		}
		return err
	}
	return op.Wait(ctx)
}

func deleteGlobalAddress(ctx context.Context, cfg *GCPConfig, name string) error {
	client, err := compute.NewGlobalAddressesRESTClient(ctx)
	if err != nil {
		return err
	}
	defer client.Close()

	op, err := client.Delete(ctx, &computepb.DeleteGlobalAddressRequest{
		Project: cfg.Project,
		Address: name,
	})
	if err != nil {
		if strings.Contains(err.Error(), "not found") || strings.Contains(err.Error(), "404") {
			return nil
		}
		return err
	}
	return op.Wait(ctx)
}

// WaitForLoadBalancerActive waits for the load balancer to be ready
func WaitForLoadBalancerActive(ctx context.Context, cfg *GCPConfig, installationKey string) error {
	// For GCP, we check if the forwarding rule exists and the backend service is healthy
	forwardingRulesClient, err := compute.NewGlobalForwardingRulesRESTClient(ctx)
	if err != nil {
		return fmt.Errorf("failed to create forwarding rules client: %w", err)
	}
	defer forwardingRulesClient.Close()

	fwdName := fmt.Sprintf("kl-%s-fwd", shortKey(installationKey))

	deadline := time.Now().Add(5 * time.Minute)
	for time.Now().Before(deadline) {
		_, err := forwardingRulesClient.Get(ctx, &computepb.GetGlobalForwardingRuleRequest{
			Project:        cfg.Project,
			ForwardingRule: fwdName,
		})
		if err == nil {
			return nil
		}

		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(10 * time.Second):
			// Continue polling
		}
	}

	return fmt.Errorf("load balancer did not become active within timeout")
}

// GetLoadBalancerIP returns the external IP of the load balancer
func GetLoadBalancerIP(ctx context.Context, cfg *GCPConfig, installationKey string) (string, error) {
	addressesClient, err := compute.NewGlobalAddressesRESTClient(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to create addresses client: %w", err)
	}
	defer addressesClient.Close()

	addressName := fmt.Sprintf("kl-%s-ip", shortKey(installationKey))

	result, err := addressesClient.Get(ctx, &computepb.GetGlobalAddressRequest{
		Project: cfg.Project,
		Address: addressName,
	})
	if err != nil {
		return "", fmt.Errorf("failed to get IP address: %w", err)
	}

	return *result.Address, nil
}

func ptrFloat32(f float32) *float32 {
	return &f
}
