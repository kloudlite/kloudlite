package azure

import (
	"context"
	"fmt"

	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/network/armnetwork/v4"
)

// LoadBalancerInfo contains information about the created Load Balancer
type LoadBalancerInfo struct {
	ID       string
	PublicIP string
}

// CreateLBPublicIP creates a public IP for the Load Balancer
func CreateLBPublicIP(ctx context.Context, cfg *AzureConfig, installationKey string) (string, error) {
	client, err := armnetwork.NewPublicIPAddressesClient(cfg.SubscriptionID, cfg.Credential, nil)
	if err != nil {
		return "", fmt.Errorf("failed to create public IP client: %w", err)
	}

	pipName := fmt.Sprintf("kl-%s-lb-pip", installationKey)

	// Check if public IP already exists
	existing, err := client.Get(ctx, cfg.ResourceGroup, pipName, nil)
	if err == nil {
		return *existing.ID, nil
	}

	tags := map[string]*string{
		"Name":                      &pipName,
		"ManagedBy":                 strPtr("kloudlite"),
		"Project":                   strPtr("kloudlite"),
		"Purpose":                   strPtr("kloudlite-installation-lb"),
		"kloudlite-installation-id": &installationKey,
	}

	poller, err := client.BeginCreateOrUpdate(ctx, cfg.ResourceGroup, pipName, armnetwork.PublicIPAddress{
		Location: &cfg.Location,
		Tags:     tags,
		Properties: &armnetwork.PublicIPAddressPropertiesFormat{
			PublicIPAllocationMethod: toIPAllocationMethod(armnetwork.IPAllocationMethodStatic),
			PublicIPAddressVersion:   toIPVersion(armnetwork.IPVersionIPv4),
		},
		SKU: &armnetwork.PublicIPAddressSKU{
			Name: toPublicIPSKUName(armnetwork.PublicIPAddressSKUNameStandard),
		},
	}, nil)
	if err != nil {
		return "", fmt.Errorf("failed to create LB public IP: %w", err)
	}

	result, err := poller.PollUntilDone(ctx, nil)
	if err != nil {
		return "", fmt.Errorf("failed to wait for LB public IP creation: %w", err)
	}

	return *result.ID, nil
}

// CreateLoadBalancer creates a Standard Load Balancer for TCP port 80 forwarding
func CreateLoadBalancer(ctx context.Context, cfg *AzureConfig, installationKey string) (*LoadBalancerInfo, error) {
	client, err := armnetwork.NewLoadBalancersClient(cfg.SubscriptionID, cfg.Credential, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create load balancer client: %w", err)
	}

	lbName := fmt.Sprintf("kl-%s-lb", shortKey(installationKey))

	// Check if Load Balancer already exists
	existing, err := client.Get(ctx, cfg.ResourceGroup, lbName, nil)
	if err == nil {
		publicIP, _ := GetLBPublicIP(ctx, cfg, installationKey)
		return &LoadBalancerInfo{
			ID:       *existing.ID,
			PublicIP: publicIP,
		}, nil
	}

	// Create public IP for LB
	publicIPID, err := CreateLBPublicIP(ctx, cfg, installationKey)
	if err != nil {
		return nil, err
	}

	tags := map[string]*string{
		"Name":                      &lbName,
		"ManagedBy":                 strPtr("kloudlite"),
		"Project":                   strPtr("kloudlite"),
		"Purpose":                   strPtr("kloudlite-installation"),
		"kloudlite-installation-id": &installationKey,
	}

	key := shortKey(installationKey)
	frontendIPConfigName := fmt.Sprintf("kl-%s-feip", key)
	backendPoolName := fmt.Sprintf("kl-%s-bepool", key)
	healthProbeName := fmt.Sprintf("kl-%s-probe", key)
	lbRuleName := fmt.Sprintf("kl-%s-rule-http", key)

	frontendIPConfigID := fmt.Sprintf("/subscriptions/%s/resourceGroups/%s/providers/Microsoft.Network/loadBalancers/%s/frontendIPConfigurations/%s",
		cfg.SubscriptionID, cfg.ResourceGroup, lbName, frontendIPConfigName)
	backendPoolID := fmt.Sprintf("/subscriptions/%s/resourceGroups/%s/providers/Microsoft.Network/loadBalancers/%s/backendAddressPools/%s",
		cfg.SubscriptionID, cfg.ResourceGroup, lbName, backendPoolName)
	healthProbeID := fmt.Sprintf("/subscriptions/%s/resourceGroups/%s/providers/Microsoft.Network/loadBalancers/%s/probes/%s",
		cfg.SubscriptionID, cfg.ResourceGroup, lbName, healthProbeName)

	poller, err := client.BeginCreateOrUpdate(ctx, cfg.ResourceGroup, lbName, armnetwork.LoadBalancer{
		Location: &cfg.Location,
		Tags:     tags,
		SKU: &armnetwork.LoadBalancerSKU{
			Name: toLBSKUName(armnetwork.LoadBalancerSKUNameStandard),
		},
		Properties: &armnetwork.LoadBalancerPropertiesFormat{
			FrontendIPConfigurations: []*armnetwork.FrontendIPConfiguration{
				{
					Name: &frontendIPConfigName,
					Properties: &armnetwork.FrontendIPConfigurationPropertiesFormat{
						PublicIPAddress: &armnetwork.PublicIPAddress{
							ID: &publicIPID,
						},
					},
				},
			},
			BackendAddressPools: []*armnetwork.BackendAddressPool{
				{
					Name: &backendPoolName,
				},
			},
			Probes: []*armnetwork.Probe{
				{
					Name: &healthProbeName,
					Properties: &armnetwork.ProbePropertiesFormat{
						Protocol:          toProbeProtocol(armnetwork.ProbeProtocolTCP),
						Port:              int32Ptr(80),
						IntervalInSeconds: int32Ptr(15),
						NumberOfProbes:    int32Ptr(2),
					},
				},
			},
			LoadBalancingRules: []*armnetwork.LoadBalancingRule{
				{
					Name: &lbRuleName,
					Properties: &armnetwork.LoadBalancingRulePropertiesFormat{
						FrontendIPConfiguration: &armnetwork.SubResource{
							ID: &frontendIPConfigID,
						},
						BackendAddressPool: &armnetwork.SubResource{
							ID: &backendPoolID,
						},
						Probe: &armnetwork.SubResource{
							ID: &healthProbeID,
						},
						Protocol:             toTransportProtocol(armnetwork.TransportProtocolTCP),
						FrontendPort:         int32Ptr(80),
						BackendPort:          int32Ptr(80),
						IdleTimeoutInMinutes: int32Ptr(4),
						EnableFloatingIP:     boolPtr(false),
					},
				},
			},
		},
	}, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create Load Balancer: %w", err)
	}

	result, err := poller.PollUntilDone(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to wait for Load Balancer creation: %w", err)
	}

	// Get the public IP address
	publicIP, err := GetLBPublicIP(ctx, cfg, installationKey)
	if err != nil {
		publicIP = ""
	}

	return &LoadBalancerInfo{
		ID:       *result.ID,
		PublicIP: publicIP,
	}, nil
}

// AddNICToBackendPool adds a NIC to the Load Balancer backend pool
func AddNICToBackendPool(ctx context.Context, cfg *AzureConfig, installationKey, nicID string) error {
	nicClient, err := armnetwork.NewInterfacesClient(cfg.SubscriptionID, cfg.Credential, nil)
	if err != nil {
		return fmt.Errorf("failed to create NIC client: %w", err)
	}

	nicName := extractResourceName(nicID)

	nic, err := nicClient.Get(ctx, cfg.ResourceGroup, nicName, nil)
	if err != nil {
		return fmt.Errorf("failed to get NIC: %w", err)
	}

	key := shortKey(installationKey)
	lbName := fmt.Sprintf("kl-%s-lb", key)
	backendPoolName := fmt.Sprintf("kl-%s-bepool", key)
	backendPoolID := fmt.Sprintf("/subscriptions/%s/resourceGroups/%s/providers/Microsoft.Network/loadBalancers/%s/backendAddressPools/%s",
		cfg.SubscriptionID, cfg.ResourceGroup, lbName, backendPoolName)

	// Update the first IP configuration to include the backend pool
	if nic.Properties != nil && len(nic.Properties.IPConfigurations) > 0 {
		ipConfig := nic.Properties.IPConfigurations[0]
		if ipConfig.Properties != nil {
			ipConfig.Properties.LoadBalancerBackendAddressPools = []*armnetwork.BackendAddressPool{
				{
					ID: &backendPoolID,
				},
			}
		}
	}

	poller, err := nicClient.BeginCreateOrUpdate(ctx, cfg.ResourceGroup, nicName, nic.Interface, nil)
	if err != nil {
		return fmt.Errorf("failed to update NIC with backend pool: %w", err)
	}

	_, err = poller.PollUntilDone(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to wait for NIC update: %w", err)
	}

	return nil
}

// DeleteLoadBalancer deletes a Load Balancer and its public IP
func DeleteLoadBalancer(ctx context.Context, cfg *AzureConfig, installationKey string) error {
	client, err := armnetwork.NewLoadBalancersClient(cfg.SubscriptionID, cfg.Credential, nil)
	if err != nil {
		return fmt.Errorf("failed to create load balancer client: %w", err)
	}

	lbName := fmt.Sprintf("kl-%s-lb", shortKey(installationKey))

	// Check if LB exists
	_, err = client.Get(ctx, cfg.ResourceGroup, lbName, nil)
	if err != nil {
		// LB doesn't exist, still try to clean up public IP
		return deleteLBPublicIP(ctx, cfg, installationKey)
	}

	poller, err := client.BeginDelete(ctx, cfg.ResourceGroup, lbName, nil)
	if err != nil {
		return fmt.Errorf("failed to start Load Balancer deletion: %w", err)
	}

	_, err = poller.PollUntilDone(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to delete Load Balancer: %w", err)
	}

	return deleteLBPublicIP(ctx, cfg, installationKey)
}

// deleteLBPublicIP deletes the Load Balancer public IP
func deleteLBPublicIP(ctx context.Context, cfg *AzureConfig, installationKey string) error {
	client, err := armnetwork.NewPublicIPAddressesClient(cfg.SubscriptionID, cfg.Credential, nil)
	if err != nil {
		return fmt.Errorf("failed to create public IP client: %w", err)
	}

	pipName := fmt.Sprintf("kl-%s-lb-pip", installationKey)

	_, err = client.Get(ctx, cfg.ResourceGroup, pipName, nil)
	if err != nil {
		return nil // doesn't exist
	}

	poller, err := client.BeginDelete(ctx, cfg.ResourceGroup, pipName, nil)
	if err != nil {
		return fmt.Errorf("failed to start LB public IP deletion: %w", err)
	}

	_, err = poller.PollUntilDone(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to delete LB public IP: %w", err)
	}

	return nil
}

// GetLBPublicIP gets the public IP address of the Load Balancer
func GetLBPublicIP(ctx context.Context, cfg *AzureConfig, installationKey string) (string, error) {
	client, err := armnetwork.NewPublicIPAddressesClient(cfg.SubscriptionID, cfg.Credential, nil)
	if err != nil {
		return "", fmt.Errorf("failed to create public IP client: %w", err)
	}

	pipName := fmt.Sprintf("kl-%s-lb-pip", installationKey)

	pip, err := client.Get(ctx, cfg.ResourceGroup, pipName, nil)
	if err != nil {
		return "", fmt.Errorf("LB public IP not found: %w", err)
	}

	if pip.Properties != nil && pip.Properties.IPAddress != nil {
		return *pip.Properties.IPAddress, nil
	}

	return "", fmt.Errorf("LB public IP address not yet assigned")
}

// FindLoadBalancerByInstallationKey finds a Load Balancer by installation key
func FindLoadBalancerByInstallationKey(ctx context.Context, cfg *AzureConfig, installationKey string) (string, error) {
	client, err := armnetwork.NewLoadBalancersClient(cfg.SubscriptionID, cfg.Credential, nil)
	if err != nil {
		return "", fmt.Errorf("failed to create load balancer client: %w", err)
	}

	lbName := fmt.Sprintf("kl-%s-lb", shortKey(installationKey))

	result, err := client.Get(ctx, cfg.ResourceGroup, lbName, nil)
	if err != nil {
		return "", nil // Not found
	}

	return *result.ID, nil
}

// Helper functions for typed pointers

func toLBSKUName(n armnetwork.LoadBalancerSKUName) *armnetwork.LoadBalancerSKUName {
	return &n
}

func toProbeProtocol(p armnetwork.ProbeProtocol) *armnetwork.ProbeProtocol {
	return &p
}

func toTransportProtocol(p armnetwork.TransportProtocol) *armnetwork.TransportProtocol {
	return &p
}
