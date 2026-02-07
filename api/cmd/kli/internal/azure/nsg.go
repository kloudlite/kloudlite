package azure

import (
	"context"
	"fmt"

	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/network/armnetwork/v4"
)

// EnsureNetworkSecurityGroup creates a Network Security Group for the VM if it doesn't exist
func EnsureNetworkSecurityGroup(ctx context.Context, cfg *AzureConfig, vpcCIDR, installationKey string) (string, error) {
	client, err := armnetwork.NewSecurityGroupsClient(cfg.SubscriptionID, cfg.Credential, nil)
	if err != nil {
		return "", fmt.Errorf("failed to create NSG client: %w", err)
	}

	nsgName := fmt.Sprintf("kl-%s-nsg", installationKey)

	// Check if NSG already exists
	existing, err := client.Get(ctx, cfg.ResourceGroup, nsgName, nil)
	if err == nil {
		return *existing.ID, nil
	}

	// Create NSG with security rules
	tags := map[string]*string{
		"Name":                      &nsgName,
		"ManagedBy":                 strPtr("kloudlite"),
		"Project":                   strPtr("kloudlite"),
		"Purpose":                   strPtr("kloudlite-installation"),
		"kloudlite-installation-id": &installationKey,
	}

	// Define security rules matching AWS security group
	securityRules := []*armnetwork.SecurityRule{
		// HTTPS from anywhere (for worker node services/tunnels)
		{
			Name: strPtr("Allow-HTTPS"),
			Properties: &armnetwork.SecurityRulePropertiesFormat{
				Protocol:                 toProtocol("Tcp"),
				SourceAddressPrefix:      strPtr("*"),
				SourcePortRange:          strPtr("*"),
				DestinationAddressPrefix: strPtr("*"),
				DestinationPortRange:     strPtr("443"),
				Access:                   toAccess("Allow"),
				Direction:                toDirection("Inbound"),
				Priority:                 int32Ptr(100),
				Description:              strPtr("HTTPS for worker services"),
			},
		},
		// K3s API from VNet CIDR
		{
			Name: strPtr("Allow-K3s-API"),
			Properties: &armnetwork.SecurityRulePropertiesFormat{
				Protocol:                 toProtocol("Tcp"),
				SourceAddressPrefix:      strPtr(vpcCIDR),
				SourcePortRange:          strPtr("*"),
				DestinationAddressPrefix: strPtr("*"),
				DestinationPortRange:     strPtr("6443"),
				Access:                   toAccess("Allow"),
				Direction:                toDirection("Inbound"),
				Priority:                 int32Ptr(110),
				Description:              strPtr("K3s API from VNet"),
			},
		},
		// Flannel VXLAN from VNet CIDR
		{
			Name: strPtr("Allow-Flannel-VXLAN"),
			Properties: &armnetwork.SecurityRulePropertiesFormat{
				Protocol:                 toProtocol("Udp"),
				SourceAddressPrefix:      strPtr(vpcCIDR),
				SourcePortRange:          strPtr("*"),
				DestinationAddressPrefix: strPtr("*"),
				DestinationPortRange:     strPtr("8472"),
				Access:                   toAccess("Allow"),
				Direction:                toDirection("Inbound"),
				Priority:                 int32Ptr(120),
				Description:              strPtr("Flannel VXLAN from VNet"),
			},
		},
		// Kubelet API from VNet CIDR
		{
			Name: strPtr("Allow-Kubelet"),
			Properties: &armnetwork.SecurityRulePropertiesFormat{
				Protocol:                 toProtocol("Tcp"),
				SourceAddressPrefix:      strPtr(vpcCIDR),
				SourcePortRange:          strPtr("*"),
				DestinationAddressPrefix: strPtr("*"),
				DestinationPortRange:     strPtr("10250"),
				Access:                   toAccess("Allow"),
				Direction:                toDirection("Inbound"),
				Priority:                 int32Ptr(130),
				Description:              strPtr("Kubelet API from VNet"),
			},
		},
		// Internal service from VNet CIDR
		{
			Name: strPtr("Allow-Internal-Service"),
			Properties: &armnetwork.SecurityRulePropertiesFormat{
				Protocol:                 toProtocol("Tcp"),
				SourceAddressPrefix:      strPtr(vpcCIDR),
				SourcePortRange:          strPtr("*"),
				DestinationAddressPrefix: strPtr("*"),
				DestinationPortRange:     strPtr("5001"),
				Access:                   toAccess("Allow"),
				Direction:                toDirection("Inbound"),
				Priority:                 int32Ptr(140),
				Description:              strPtr("Internal service from VNet"),
			},
		},
		// Allow all outbound
		{
			Name: strPtr("Allow-All-Outbound"),
			Properties: &armnetwork.SecurityRulePropertiesFormat{
				Protocol:                 toProtocol("*"),
				SourceAddressPrefix:      strPtr("*"),
				SourcePortRange:          strPtr("*"),
				DestinationAddressPrefix: strPtr("*"),
				DestinationPortRange:     strPtr("*"),
				Access:                   toAccess("Allow"),
				Direction:                toDirection("Outbound"),
				Priority:                 int32Ptr(100),
				Description:              strPtr("Allow all outbound traffic"),
			},
		},
	}

	poller, err := client.BeginCreateOrUpdate(ctx, cfg.ResourceGroup, nsgName, armnetwork.SecurityGroup{
		Location: &cfg.Location,
		Tags:     tags,
		Properties: &armnetwork.SecurityGroupPropertiesFormat{
			SecurityRules: securityRules,
		},
	}, nil)
	if err != nil {
		return "", fmt.Errorf("failed to create NSG: %w", err)
	}

	result, err := poller.PollUntilDone(ctx, nil)
	if err != nil {
		return "", fmt.Errorf("failed to wait for NSG creation: %w", err)
	}

	return *result.ID, nil
}

// EnsureMasterNSG creates a Network Security Group for the master node
func EnsureMasterNSG(ctx context.Context, cfg *AzureConfig, vpcCIDR, installationKey string) (string, error) {
	client, err := armnetwork.NewSecurityGroupsClient(cfg.SubscriptionID, cfg.Credential, nil)
	if err != nil {
		return "", fmt.Errorf("failed to create NSG client: %w", err)
	}

	nsgName := fmt.Sprintf("kl-%s-master-nsg", installationKey)

	// Check if NSG already exists
	existing, err := client.Get(ctx, cfg.ResourceGroup, nsgName, nil)
	if err == nil {
		return *existing.ID, nil
	}

	// Create NSG with security rules
	tags := map[string]*string{
		"Name":                      &nsgName,
		"ManagedBy":                 strPtr("kloudlite"),
		"Project":                   strPtr("kloudlite"),
		"Purpose":                   strPtr("kloudlite-installation-master"),
		"kloudlite-installation-id": &installationKey,
	}

	// Define security rules for master node
	securityRules := []*armnetwork.SecurityRule{
		// Allow Azure Load Balancer health probes
		{
			Name: strPtr("Allow-LB-HealthProbe"),
			Properties: &armnetwork.SecurityRulePropertiesFormat{
				Protocol:                 toProtocol("Tcp"),
				SourceAddressPrefix:      strPtr("AzureLoadBalancer"),
				SourcePortRange:          strPtr("*"),
				DestinationAddressPrefix: strPtr("*"),
				DestinationPortRange:     strPtr("80"),
				Access:                   toAccess("Allow"),
				Direction:                toDirection("Inbound"),
				Priority:                 int32Ptr(95),
				Description:              strPtr("Azure LB health probes"),
			},
		},
		// HTTP from anywhere (Cloudflare → LB → VM)
		{
			Name: strPtr("Allow-HTTP"),
			Properties: &armnetwork.SecurityRulePropertiesFormat{
				Protocol:                 toProtocol("Tcp"),
				SourceAddressPrefix:      strPtr("*"),
				SourcePortRange:          strPtr("*"),
				DestinationAddressPrefix: strPtr("*"),
				DestinationPortRange:     strPtr("80"),
				Access:                   toAccess("Allow"),
				Direction:                toDirection("Inbound"),
				Priority:                 int32Ptr(100),
				Description:              strPtr("HTTP from anywhere"),
			},
		},
		// HTTPS from anywhere
		{
			Name: strPtr("Allow-HTTPS"),
			Properties: &armnetwork.SecurityRulePropertiesFormat{
				Protocol:                 toProtocol("Tcp"),
				SourceAddressPrefix:      strPtr("*"),
				SourcePortRange:          strPtr("*"),
				DestinationAddressPrefix: strPtr("*"),
				DestinationPortRange:     strPtr("443"),
				Access:                   toAccess("Allow"),
				Direction:                toDirection("Inbound"),
				Priority:                 int32Ptr(110),
				Description:              strPtr("HTTPS from anywhere"),
			},
		},
		// K3s API from VNet CIDR
		{
			Name: strPtr("Allow-K3s-API"),
			Properties: &armnetwork.SecurityRulePropertiesFormat{
				Protocol:                 toProtocol("Tcp"),
				SourceAddressPrefix:      strPtr(vpcCIDR),
				SourcePortRange:          strPtr("*"),
				DestinationAddressPrefix: strPtr("*"),
				DestinationPortRange:     strPtr("6443"),
				Access:                   toAccess("Allow"),
				Direction:                toDirection("Inbound"),
				Priority:                 int32Ptr(120),
				Description:              strPtr("K3s API from VNet"),
			},
		},
		// Flannel VXLAN from VNet CIDR
		{
			Name: strPtr("Allow-Flannel-VXLAN"),
			Properties: &armnetwork.SecurityRulePropertiesFormat{
				Protocol:                 toProtocol("Udp"),
				SourceAddressPrefix:      strPtr(vpcCIDR),
				SourcePortRange:          strPtr("*"),
				DestinationAddressPrefix: strPtr("*"),
				DestinationPortRange:     strPtr("8472"),
				Access:                   toAccess("Allow"),
				Direction:                toDirection("Inbound"),
				Priority:                 int32Ptr(130),
				Description:              strPtr("Flannel VXLAN from VNet"),
			},
		},
		// Kubelet API from VNet CIDR
		{
			Name: strPtr("Allow-Kubelet"),
			Properties: &armnetwork.SecurityRulePropertiesFormat{
				Protocol:                 toProtocol("Tcp"),
				SourceAddressPrefix:      strPtr(vpcCIDR),
				SourcePortRange:          strPtr("*"),
				DestinationAddressPrefix: strPtr("*"),
				DestinationPortRange:     strPtr("10250"),
				Access:                   toAccess("Allow"),
				Direction:                toDirection("Inbound"),
				Priority:                 int32Ptr(140),
				Description:              strPtr("Kubelet API from VNet"),
			},
		},
		// Internal service from VNet CIDR
		{
			Name: strPtr("Allow-Internal-Service"),
			Properties: &armnetwork.SecurityRulePropertiesFormat{
				Protocol:                 toProtocol("Tcp"),
				SourceAddressPrefix:      strPtr(vpcCIDR),
				SourcePortRange:          strPtr("*"),
				DestinationAddressPrefix: strPtr("*"),
				DestinationPortRange:     strPtr("5001"),
				Access:                   toAccess("Allow"),
				Direction:                toDirection("Inbound"),
				Priority:                 int32Ptr(150),
				Description:              strPtr("Internal service from VNet"),
			},
		},
		// Allow all outbound
		{
			Name: strPtr("Allow-All-Outbound"),
			Properties: &armnetwork.SecurityRulePropertiesFormat{
				Protocol:                 toProtocol("*"),
				SourceAddressPrefix:      strPtr("*"),
				SourcePortRange:          strPtr("*"),
				DestinationAddressPrefix: strPtr("*"),
				DestinationPortRange:     strPtr("*"),
				Access:                   toAccess("Allow"),
				Direction:                toDirection("Outbound"),
				Priority:                 int32Ptr(100),
				Description:              strPtr("Allow all outbound traffic"),
			},
		},
	}

	poller, err := client.BeginCreateOrUpdate(ctx, cfg.ResourceGroup, nsgName, armnetwork.SecurityGroup{
		Location: &cfg.Location,
		Tags:     tags,
		Properties: &armnetwork.SecurityGroupPropertiesFormat{
			SecurityRules: securityRules,
		},
	}, nil)
	if err != nil {
		return "", fmt.Errorf("failed to create master NSG: %w", err)
	}

	result, err := poller.PollUntilDone(ctx, nil)
	if err != nil {
		return "", fmt.Errorf("failed to wait for master NSG creation: %w", err)
	}

	return *result.ID, nil
}

// DeleteNSG deletes a Network Security Group by ID
func DeleteNSG(ctx context.Context, cfg *AzureConfig, nsgID string) error {
	client, err := armnetwork.NewSecurityGroupsClient(cfg.SubscriptionID, cfg.Credential, nil)
	if err != nil {
		return fmt.Errorf("failed to create NSG client: %w", err)
	}

	nsgName := extractResourceName(nsgID)

	// Check if NSG exists
	_, err = client.Get(ctx, cfg.ResourceGroup, nsgName, nil)
	if err != nil {
		// NSG doesn't exist
		return nil
	}

	poller, err := client.BeginDelete(ctx, cfg.ResourceGroup, nsgName, nil)
	if err != nil {
		return fmt.Errorf("failed to start NSG deletion: %w", err)
	}

	_, err = poller.PollUntilDone(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to delete NSG: %w", err)
	}

	return nil
}

// DeleteNSGByName deletes a Network Security Group by name
func DeleteNSGByName(ctx context.Context, cfg *AzureConfig, nsgName string) error {
	client, err := armnetwork.NewSecurityGroupsClient(cfg.SubscriptionID, cfg.Credential, nil)
	if err != nil {
		return fmt.Errorf("failed to create NSG client: %w", err)
	}

	// Check if NSG exists
	_, err = client.Get(ctx, cfg.ResourceGroup, nsgName, nil)
	if err != nil {
		// NSG doesn't exist
		return nil
	}

	poller, err := client.BeginDelete(ctx, cfg.ResourceGroup, nsgName, nil)
	if err != nil {
		return fmt.Errorf("failed to start NSG deletion: %w", err)
	}

	_, err = poller.PollUntilDone(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to delete NSG: %w", err)
	}

	return nil
}

// FindNSGByInstallationKey finds a Network Security Group by installation key
func FindNSGByInstallationKey(ctx context.Context, cfg *AzureConfig, installationKey string, isLoadBalancer bool) (string, error) {
	if isLoadBalancer {
		// Standard LB has no separate NSG
		return "", nil
	}

	client, err := armnetwork.NewSecurityGroupsClient(cfg.SubscriptionID, cfg.Credential, nil)
	if err != nil {
		return "", fmt.Errorf("failed to create NSG client: %w", err)
	}

	nsgName := fmt.Sprintf("kl-%s-nsg", installationKey)

	result, err := client.Get(ctx, cfg.ResourceGroup, nsgName, nil)
	if err != nil {
		return "", nil // NSG not found
	}

	return *result.ID, nil
}

// Helper functions for creating typed pointers

func toProtocol(p string) *armnetwork.SecurityRuleProtocol {
	var protocol armnetwork.SecurityRuleProtocol
	switch p {
	case "Tcp":
		protocol = armnetwork.SecurityRuleProtocolTCP
	case "Udp":
		protocol = armnetwork.SecurityRuleProtocolUDP
	case "*":
		protocol = armnetwork.SecurityRuleProtocolAsterisk
	default:
		protocol = armnetwork.SecurityRuleProtocol(p)
	}
	return &protocol
}

func toAccess(a string) *armnetwork.SecurityRuleAccess {
	var access armnetwork.SecurityRuleAccess
	switch a {
	case "Allow":
		access = armnetwork.SecurityRuleAccessAllow
	case "Deny":
		access = armnetwork.SecurityRuleAccessDeny
	default:
		access = armnetwork.SecurityRuleAccess(a)
	}
	return &access
}

func toDirection(d string) *armnetwork.SecurityRuleDirection {
	var direction armnetwork.SecurityRuleDirection
	switch d {
	case "Inbound":
		direction = armnetwork.SecurityRuleDirectionInbound
	case "Outbound":
		direction = armnetwork.SecurityRuleDirectionOutbound
	default:
		direction = armnetwork.SecurityRuleDirection(d)
	}
	return &direction
}
