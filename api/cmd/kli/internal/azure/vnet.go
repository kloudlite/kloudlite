package azure

import (
	"context"
	"fmt"

	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/network/armnetwork/v4"
	"github.com/kloudlite/kloudlite/api/cmd/kli/internal/provider"
)

const (
	// Default VNet CIDR
	DefaultVNetCIDR = "10.0.0.0/16"
	// Default Subnet CIDR
	DefaultSubnetCIDR = "10.0.0.0/24"
)

// GetDefaultVNet finds or creates a default VNet for Kloudlite
func GetDefaultVNet(ctx context.Context, cfg *AzureConfig) (string, string, error) {
	client, err := armnetwork.NewVirtualNetworksClient(cfg.SubscriptionID, cfg.Credential, nil)
	if err != nil {
		return "", "", fmt.Errorf("failed to create vnet client: %w", err)
	}

	// Try to find existing VNet with kloudlite tag in the resource group
	pager := client.NewListPager(cfg.ResourceGroup, nil)
	for pager.More() {
		page, err := pager.NextPage(ctx)
		if err != nil {
			break // No VNets found, will create one
		}

		for _, vnet := range page.Value {
			if vnet.Tags != nil {
				if _, ok := vnet.Tags["ManagedBy"]; ok {
					if vnet.Properties != nil && vnet.Properties.AddressSpace != nil &&
						len(vnet.Properties.AddressSpace.AddressPrefixes) > 0 {
						return *vnet.ID, *vnet.Properties.AddressSpace.AddressPrefixes[0], nil
					}
				}
			}
		}
	}

	return "", "", fmt.Errorf("no VNet found in resource group %s. Use EnsureVNet to create one", cfg.ResourceGroup)
}

// EnsureVNet creates a VNet if it doesn't exist
func EnsureVNet(ctx context.Context, cfg *AzureConfig, installationKey string) (string, string, error) {
	client, err := armnetwork.NewVirtualNetworksClient(cfg.SubscriptionID, cfg.Credential, nil)
	if err != nil {
		return "", "", fmt.Errorf("failed to create vnet client: %w", err)
	}

	vnetName := fmt.Sprintf("kl-%s-vnet", installationKey)

	// Check if VNet already exists
	existing, err := client.Get(ctx, cfg.ResourceGroup, vnetName, nil)
	if err == nil {
		// VNet exists
		cidr := ""
		if existing.Properties != nil && existing.Properties.AddressSpace != nil &&
			len(existing.Properties.AddressSpace.AddressPrefixes) > 0 {
			cidr = *existing.Properties.AddressSpace.AddressPrefixes[0]
		}
		return *existing.ID, cidr, nil
	}

	// Create VNet
	tags := map[string]*string{
		"Name":                      &vnetName,
		"ManagedBy":                 strPtr("kloudlite"),
		"Project":                   strPtr("kloudlite"),
		"Purpose":                   strPtr("kloudlite-installation"),
		"kloudlite-installation-id": &installationKey,
	}

	poller, err := client.BeginCreateOrUpdate(ctx, cfg.ResourceGroup, vnetName, armnetwork.VirtualNetwork{
		Location: &cfg.Location,
		Tags:     tags,
		Properties: &armnetwork.VirtualNetworkPropertiesFormat{
			AddressSpace: &armnetwork.AddressSpace{
				AddressPrefixes: []*string{strPtr(DefaultVNetCIDR)},
			},
		},
	}, nil)
	if err != nil {
		return "", "", fmt.Errorf("failed to create VNet: %w", err)
	}

	result, err := poller.PollUntilDone(ctx, nil)
	if err != nil {
		return "", "", fmt.Errorf("failed to wait for VNet creation: %w", err)
	}

	return *result.ID, DefaultVNetCIDR, nil
}

// EnsureSubnet creates a subnet if it doesn't exist
func EnsureSubnet(ctx context.Context, cfg *AzureConfig, vnetName, installationKey string) (string, string, error) {
	client, err := armnetwork.NewSubnetsClient(cfg.SubscriptionID, cfg.Credential, nil)
	if err != nil {
		return "", "", fmt.Errorf("failed to create subnet client: %w", err)
	}

	subnetName := fmt.Sprintf("kl-%s-subnet", installationKey)

	// Check if subnet already exists
	existing, err := client.Get(ctx, cfg.ResourceGroup, vnetName, subnetName, nil)
	if err == nil {
		// Subnet exists
		return *existing.ID, cfg.Location, nil
	}

	// Create subnet
	poller, err := client.BeginCreateOrUpdate(ctx, cfg.ResourceGroup, vnetName, subnetName, armnetwork.Subnet{
		Properties: &armnetwork.SubnetPropertiesFormat{
			AddressPrefix: strPtr(DefaultSubnetCIDR),
		},
	}, nil)
	if err != nil {
		return "", "", fmt.Errorf("failed to create subnet: %w", err)
	}

	result, err := poller.PollUntilDone(ctx, nil)
	if err != nil {
		return "", "", fmt.Errorf("failed to wait for subnet creation: %w", err)
	}

	return *result.ID, cfg.Location, nil
}

// GetDefaultSubnet returns the first subnet in a VNet
func GetDefaultSubnet(ctx context.Context, cfg *AzureConfig, vnetID string) (string, string, error) {
	client, err := armnetwork.NewSubnetsClient(cfg.SubscriptionID, cfg.Credential, nil)
	if err != nil {
		return "", "", fmt.Errorf("failed to create subnet client: %w", err)
	}

	// Extract VNet name from ID
	vnetName := extractResourceName(vnetID)

	pager := client.NewListPager(cfg.ResourceGroup, vnetName, nil)
	for pager.More() {
		page, err := pager.NextPage(ctx)
		if err != nil {
			return "", "", fmt.Errorf("failed to list subnets: %w", err)
		}

		for _, subnet := range page.Value {
			// Skip App Gateway subnet
			if subnet.Name != nil && contains(*subnet.Name, "appgw") {
				continue
			}
			return *subnet.ID, cfg.Location, nil
		}
	}

	return "", "", fmt.Errorf("no subnet found in VNet %s", vnetID)
}

// GetAllSubnets returns all subnets in a VNet (excluding App Gateway subnet)
func GetAllSubnets(ctx context.Context, cfg *AzureConfig, vnetID string) ([]provider.SubnetInfo, error) {
	client, err := armnetwork.NewSubnetsClient(cfg.SubscriptionID, cfg.Credential, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create subnet client: %w", err)
	}

	// Extract VNet name from ID
	vnetName := extractResourceName(vnetID)

	var subnets []provider.SubnetInfo
	pager := client.NewListPager(cfg.ResourceGroup, vnetName, nil)
	for pager.More() {
		page, err := pager.NextPage(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to list subnets: %w", err)
		}

		for _, subnet := range page.Value {
			// Skip App Gateway subnet
			if subnet.Name != nil && contains(*subnet.Name, "appgw") {
				continue
			}

			cidr := ""
			if subnet.Properties != nil && subnet.Properties.AddressPrefix != nil {
				cidr = *subnet.Properties.AddressPrefix
			}

			subnets = append(subnets, provider.SubnetInfo{
				ID:               *subnet.ID,
				AvailabilityZone: cfg.Location, // Azure uses regions, not AZs for basic VNets
				CIDR:             cidr,
			})
		}
	}

	if len(subnets) == 0 {
		return nil, fmt.Errorf("no subnets found in VNet %s", vnetID)
	}

	return subnets, nil
}

// DeleteVNet deletes a VNet
func DeleteVNet(ctx context.Context, cfg *AzureConfig, installationKey string) error {
	client, err := armnetwork.NewVirtualNetworksClient(cfg.SubscriptionID, cfg.Credential, nil)
	if err != nil {
		return fmt.Errorf("failed to create vnet client: %w", err)
	}

	vnetName := fmt.Sprintf("kl-%s-vnet", installationKey)

	// Check if VNet exists
	_, err = client.Get(ctx, cfg.ResourceGroup, vnetName, nil)
	if err != nil {
		// VNet doesn't exist
		return nil
	}

	poller, err := client.BeginDelete(ctx, cfg.ResourceGroup, vnetName, nil)
	if err != nil {
		return fmt.Errorf("failed to start VNet deletion: %w", err)
	}

	_, err = poller.PollUntilDone(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to delete VNet: %w", err)
	}

	return nil
}

// ExtractResourceName extracts the resource name from an Azure resource ID
func ExtractResourceName(resourceID string) string {
	return extractResourceName(resourceID)
}

// extractResourceName is the internal version
func extractResourceName(resourceID string) string {
	// Azure resource IDs are in format: /subscriptions/{sub}/resourceGroups/{rg}/providers/{provider}/{type}/{name}
	// We want the last segment
	parts := splitString(resourceID, "/")
	if len(parts) > 0 {
		return parts[len(parts)-1]
	}
	return resourceID
}

// splitString splits a string by separator
func splitString(s, sep string) []string {
	var result []string
	start := 0
	for i := 0; i < len(s); i++ {
		if i+len(sep) <= len(s) && s[i:i+len(sep)] == sep {
			if i > start {
				result = append(result, s[start:i])
			}
			start = i + len(sep)
		}
	}
	if start < len(s) {
		result = append(result, s[start:])
	}
	return result
}

// contains checks if a string contains a substring
func contains(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
