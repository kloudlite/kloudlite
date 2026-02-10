package oci

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/oracle/oci-go-sdk/v65/core"
)

// ReservedIPName returns the reserved IP name for an installation
func ReservedIPName(installationKey string) string {
	return fmt.Sprintf("kl-%s-ip", shortKey(installationKey))
}

// AssignReservedIP assigns an existing reserved public IP to an instance's primary VNIC.
// The instance must NOT have an ephemeral public IP (AssignPublicIp must be false at launch).
func AssignReservedIP(ctx context.Context, cfg *OCIConfig, instanceID, installationKey string) (string, error) {
	computeClient, err := core.NewComputeClientWithConfigurationProvider(cfg.ConfigProvider)
	if err != nil {
		return "", fmt.Errorf("failed to create compute client: %w", err)
	}

	vnClient, err := core.NewVirtualNetworkClientWithConfigurationProvider(cfg.ConfigProvider)
	if err != nil {
		return "", fmt.Errorf("failed to create virtual network client: %w", err)
	}

	// 1. Find the instance's primary VNIC
	vnicResp, err := computeClient.ListVnicAttachments(ctx, core.ListVnicAttachmentsRequest{
		CompartmentId: &cfg.CompartmentOCID,
		InstanceId:    &instanceID,
	})
	if err != nil {
		return "", fmt.Errorf("failed to list VNIC attachments: %w", err)
	}

	var vnicID string
	for _, va := range vnicResp.Items {
		if va.VnicId != nil && va.LifecycleState == core.VnicAttachmentLifecycleStateAttached {
			vnicID = *va.VnicId
			break
		}
	}
	if vnicID == "" {
		return "", fmt.Errorf("no attached VNIC found for instance %s", instanceID)
	}

	// 2. Get the primary private IP OCID for this VNIC
	privateIpResp, err := vnClient.ListPrivateIps(ctx, core.ListPrivateIpsRequest{
		VnicId: &vnicID,
	})
	if err != nil {
		return "", fmt.Errorf("failed to list private IPs: %w", err)
	}

	var privateIpID string
	for _, pip := range privateIpResp.Items {
		if pip.Id != nil && pip.IsPrimary != nil && *pip.IsPrimary {
			privateIpID = *pip.Id
			break
		}
	}
	if privateIpID == "" {
		// Fallback: use first private IP if none marked primary
		for _, pip := range privateIpResp.Items {
			if pip.Id != nil {
				privateIpID = *pip.Id
				break
			}
		}
	}
	if privateIpID == "" {
		return "", fmt.Errorf("no private IP found for VNIC %s", vnicID)
	}

	// 3. Find the reserved IP by name
	ipName := ReservedIPName(installationKey)
	ipID, _, err := findReservedIPByName(ctx, vnClient, cfg.CompartmentOCID, ipName)
	if err != nil {
		return "", fmt.Errorf("failed to find reserved IP %s: %w", ipName, err)
	}

	// 4. Assign the reserved IP to the private IP
	updateResp, err := vnClient.UpdatePublicIp(ctx, core.UpdatePublicIpRequest{
		PublicIpId: &ipID,
		UpdatePublicIpDetails: core.UpdatePublicIpDetails{
			PrivateIpId: &privateIpID,
		},
	})
	if err != nil {
		return "", fmt.Errorf("failed to assign reserved IP to instance: %w", err)
	}

	if updateResp.IpAddress != nil {
		return *updateResp.IpAddress, nil
	}

	// Wait for IP assignment to complete
	return waitForPublicIP(ctx, vnClient, ipID)
}

// ReservePublicIP reserves a public IP for the installation
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
