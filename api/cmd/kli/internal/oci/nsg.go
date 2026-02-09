package oci

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/oracle/oci-go-sdk/v65/core"
)

// NSGName returns the NSG name for an installation
func NSGName(installationKey string) string {
	return fmt.Sprintf("kl-%s-nsg", shortKey(installationKey))
}

// EnsureNSG creates a Network Security Group with required rules
func EnsureNSG(ctx context.Context, cfg *OCIConfig, vcnID, vpcCIDR, installationKey string) (string, error) {
	vnClient, err := core.NewVirtualNetworkClientWithConfigurationProvider(cfg.ConfigProvider)
	if err != nil {
		return "", fmt.Errorf("failed to create virtual network client: %w", err)
	}

	nsgName := NSGName(installationKey)

	// Check if NSG already exists
	existingID, err := findNSGByName(ctx, vnClient, cfg.CompartmentOCID, vcnID, nsgName)
	if err == nil && existingID != "" {
		return existingID, nil
	}

	// Create NSG
	tags := freeformTags(installationKey)
	resp, err := vnClient.CreateNetworkSecurityGroup(ctx, core.CreateNetworkSecurityGroupRequest{
		CreateNetworkSecurityGroupDetails: core.CreateNetworkSecurityGroupDetails{
			CompartmentId: &cfg.CompartmentOCID,
			VcnId:         &vcnID,
			DisplayName:   &nsgName,
			FreeformTags:  tags,
		},
	})
	if err != nil {
		if strings.Contains(err.Error(), "already exists") || strings.Contains(err.Error(), "Conflict") {
			existingID, findErr := findNSGByName(ctx, vnClient, cfg.CompartmentOCID, vcnID, nsgName)
			if findErr == nil && existingID != "" {
				return existingID, nil
			}
		}
		return "", fmt.Errorf("failed to create NSG: %w", err)
	}

	nsgID := *resp.Id

	// Add security rules
	if err := CreateNSGRules(ctx, cfg, nsgID, vpcCIDR); err != nil {
		return "", fmt.Errorf("failed to create NSG rules: %w", err)
	}

	return nsgID, nil
}

// CreateNSGRules adds ingress rules to the NSG
func CreateNSGRules(ctx context.Context, cfg *OCIConfig, nsgID, vpcCIDR string) error {
	vnClient, err := core.NewVirtualNetworkClientWithConfigurationProvider(cfg.ConfigProvider)
	if err != nil {
		return fmt.Errorf("failed to create virtual network client: %w", err)
	}

	// Build security rules
	rules := []core.AddSecurityRuleDetails{
		// Allow ALL egress (outbound) traffic — required for apt-get, K3s install, etc.
		{
			Direction:       core.AddSecurityRuleDetailsDirectionEgress,
			Protocol:        strPtr("all"),
			Destination:     strPtr("0.0.0.0/0"),
			DestinationType: core.AddSecurityRuleDetailsDestinationTypeCidrBlock,
			Description:     strPtr("Allow all outbound traffic"),
		},
		// SSH from anywhere
		{
			Direction:   core.AddSecurityRuleDetailsDirectionIngress,
			Protocol:    strPtr("6"), // TCP
			Source:      strPtr("0.0.0.0/0"),
			SourceType:  core.AddSecurityRuleDetailsSourceTypeCidrBlock,
			Description: strPtr("Allow SSH from anywhere"),
			TcpOptions: &core.TcpOptions{
				DestinationPortRange: &core.PortRange{
					Min: intPtr(22),
					Max: intPtr(22),
				},
			},
		},
		// HTTP from anywhere
		{
			Direction:   core.AddSecurityRuleDetailsDirectionIngress,
			Protocol:    strPtr("6"), // TCP
			Source:      strPtr("0.0.0.0/0"),
			SourceType:  core.AddSecurityRuleDetailsSourceTypeCidrBlock,
			Description: strPtr("Allow HTTP from anywhere"),
			TcpOptions: &core.TcpOptions{
				DestinationPortRange: &core.PortRange{
					Min: intPtr(80),
					Max: intPtr(80),
				},
			},
		},
		// HTTPS from anywhere
		{
			Direction:   core.AddSecurityRuleDetailsDirectionIngress,
			Protocol:    strPtr("6"), // TCP
			Source:      strPtr("0.0.0.0/0"),
			SourceType:  core.AddSecurityRuleDetailsSourceTypeCidrBlock,
			Description: strPtr("Allow HTTPS from anywhere"),
			TcpOptions: &core.TcpOptions{
				DestinationPortRange: &core.PortRange{
					Min: intPtr(443),
					Max: intPtr(443),
				},
			},
		},
		// K3s API from VPC
		{
			Direction:   core.AddSecurityRuleDetailsDirectionIngress,
			Protocol:    strPtr("6"), // TCP
			Source:      &vpcCIDR,
			SourceType:  core.AddSecurityRuleDetailsSourceTypeCidrBlock,
			Description: strPtr("Allow K3s API server from VPC"),
			TcpOptions: &core.TcpOptions{
				DestinationPortRange: &core.PortRange{
					Min: intPtr(6443),
					Max: intPtr(6443),
				},
			},
		},
		// Kubelet from VPC
		{
			Direction:   core.AddSecurityRuleDetailsDirectionIngress,
			Protocol:    strPtr("6"), // TCP
			Source:      &vpcCIDR,
			SourceType:  core.AddSecurityRuleDetailsSourceTypeCidrBlock,
			Description: strPtr("Allow Kubelet from VPC"),
			TcpOptions: &core.TcpOptions{
				DestinationPortRange: &core.PortRange{
					Min: intPtr(10250),
					Max: intPtr(10250),
				},
			},
		},
		// Registry from VPC
		{
			Direction:   core.AddSecurityRuleDetailsDirectionIngress,
			Protocol:    strPtr("6"), // TCP
			Source:      &vpcCIDR,
			SourceType:  core.AddSecurityRuleDetailsSourceTypeCidrBlock,
			Description: strPtr("Allow registry from VPC"),
			TcpOptions: &core.TcpOptions{
				DestinationPortRange: &core.PortRange{
					Min: intPtr(5001),
					Max: intPtr(5001),
				},
			},
		},
		// VXLAN (Flannel) from VPC
		{
			Direction:   core.AddSecurityRuleDetailsDirectionIngress,
			Protocol:    strPtr("17"), // UDP
			Source:      &vpcCIDR,
			SourceType:  core.AddSecurityRuleDetailsSourceTypeCidrBlock,
			Description: strPtr("Allow VXLAN from VPC"),
			UdpOptions: &core.UdpOptions{
				DestinationPortRange: &core.PortRange{
					Min: intPtr(8472),
					Max: intPtr(8472),
				},
			},
		},
	}

	_, err = vnClient.AddNetworkSecurityGroupSecurityRules(ctx, core.AddNetworkSecurityGroupSecurityRulesRequest{
		NetworkSecurityGroupId: &nsgID,
		AddNetworkSecurityGroupSecurityRulesDetails: core.AddNetworkSecurityGroupSecurityRulesDetails{
			SecurityRules: rules,
		},
	})
	if err != nil {
		return fmt.Errorf("failed to add NSG rules: %w", err)
	}

	return nil
}

// DeleteNSG deletes a Network Security Group, retrying if VNICs are still attached
func DeleteNSG(ctx context.Context, cfg *OCIConfig, vcnID, installationKey string) error {
	vnClient, err := core.NewVirtualNetworkClientWithConfigurationProvider(cfg.ConfigProvider)
	if err != nil {
		return fmt.Errorf("failed to create virtual network client: %w", err)
	}

	nsgName := NSGName(installationKey)
	nsgID, err := findNSGByName(ctx, vnClient, cfg.CompartmentOCID, vcnID, nsgName)
	if err != nil || nsgID == "" {
		return nil // Already gone
	}

	// Retry deletion — instances may still be terminating with VNICs attached
	deadline := time.Now().Add(3 * time.Minute)
	for time.Now().Before(deadline) {
		_, err = vnClient.DeleteNetworkSecurityGroup(ctx, core.DeleteNetworkSecurityGroupRequest{
			NetworkSecurityGroupId: &nsgID,
		})
		if err == nil {
			return nil
		}
		if strings.Contains(err.Error(), "404") || strings.Contains(err.Error(), "not found") {
			return nil
		}
		// If VNICs still attached, wait and retry
		if strings.Contains(err.Error(), "PreconditionFailed") || strings.Contains(err.Error(), "vnics attached") {
			fmt.Printf("    NSG still has VNICs attached, waiting for instance termination...\n")
			time.Sleep(15 * time.Second)
			continue
		}
		return fmt.Errorf("failed to delete NSG: %w", err)
	}

	return fmt.Errorf("failed to delete NSG after retries: %w", err)
}

// findNSGByName finds an NSG by display name in the compartment/VCN
func findNSGByName(ctx context.Context, vnClient core.VirtualNetworkClient, compartmentID, vcnID, name string) (string, error) {
	resp, err := vnClient.ListNetworkSecurityGroups(ctx, core.ListNetworkSecurityGroupsRequest{
		CompartmentId: &compartmentID,
		VcnId:         &vcnID,
		DisplayName:   &name,
	})
	if err != nil {
		return "", err
	}

	for _, nsg := range resp.Items {
		if nsg.Id != nil && nsg.LifecycleState == core.NetworkSecurityGroupLifecycleStateAvailable {
			return *nsg.Id, nil
		}
	}

	return "", fmt.Errorf("NSG %s not found", name)
}

func intPtr(i int) *int {
	return &i
}
