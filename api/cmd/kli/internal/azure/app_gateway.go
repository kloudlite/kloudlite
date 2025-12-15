package azure

import (
	"context"
	"fmt"
	"time"

	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/network/armnetwork/v4"
)

// AppGatewayInfo contains information about the created Application Gateway
type AppGatewayInfo struct {
	ID       string
	PublicIP string
}

// CreateAppGatewayPublicIP creates a public IP for the Application Gateway
func CreateAppGatewayPublicIP(ctx context.Context, cfg *AzureConfig, installationKey string) (string, error) {
	client, err := armnetwork.NewPublicIPAddressesClient(cfg.SubscriptionID, cfg.Credential, nil)
	if err != nil {
		return "", fmt.Errorf("failed to create public IP client: %w", err)
	}

	pipName := fmt.Sprintf("kl-%s-appgw-pip", installationKey)

	// Check if public IP already exists
	existing, err := client.Get(ctx, cfg.ResourceGroup, pipName, nil)
	if err == nil {
		return *existing.ID, nil
	}

	tags := map[string]*string{
		"Name":                      &pipName,
		"ManagedBy":                 strPtr("kloudlite"),
		"Project":                   strPtr("kloudlite"),
		"Purpose":                   strPtr("kloudlite-installation-appgw"),
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
		return "", fmt.Errorf("failed to create App Gateway public IP: %w", err)
	}

	result, err := poller.PollUntilDone(ctx, nil)
	if err != nil {
		return "", fmt.Errorf("failed to wait for App Gateway public IP creation: %w", err)
	}

	return *result.ID, nil
}

// CreateApplicationGateway creates an Application Gateway for load balancing
func CreateApplicationGateway(ctx context.Context, cfg *AzureConfig, installationKey, vnetID string, subnetIDs []string) (*AppGatewayInfo, error) {
	client, err := armnetwork.NewApplicationGatewaysClient(cfg.SubscriptionID, cfg.Credential, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create app gateway client: %w", err)
	}

	appGwName := fmt.Sprintf("kl-%s-appgw", shortKey(installationKey))

	// Check if Application Gateway already exists
	existing, err := client.Get(ctx, cfg.ResourceGroup, appGwName, nil)
	if err == nil {
		publicIP, _ := GetAppGatewayPublicIP(ctx, cfg, *existing.ID)
		return &AppGatewayInfo{
			ID:       *existing.ID,
			PublicIP: publicIP,
		}, nil
	}

	// Create public IP for App Gateway
	publicIPID, err := CreateAppGatewayPublicIP(ctx, cfg, installationKey)
	if err != nil {
		return nil, err
	}

	// Get the App Gateway subnet (needs a dedicated subnet)
	vnetName := extractResourceName(vnetID)
	appGwSubnetID, err := EnsureAppGatewaySubnet(ctx, cfg, vnetName, installationKey)
	if err != nil {
		return nil, fmt.Errorf("failed to create App Gateway subnet: %w", err)
	}

	tags := map[string]*string{
		"Name":                      &appGwName,
		"ManagedBy":                 strPtr("kloudlite"),
		"Project":                   strPtr("kloudlite"),
		"Purpose":                   strPtr("kloudlite-installation"),
		"kloudlite-installation-id": &installationKey,
	}

	// Configuration names
	gatewayIPConfigName := fmt.Sprintf("kl-%s-gwip", shortKey(installationKey))
	frontendIPConfigName := fmt.Sprintf("kl-%s-feip", shortKey(installationKey))
	frontendPortHTTPName := fmt.Sprintf("kl-%s-feport-http", shortKey(installationKey))
	backendPoolName := fmt.Sprintf("kl-%s-bepool", shortKey(installationKey))
	backendSettingsName := fmt.Sprintf("kl-%s-besettings", shortKey(installationKey))
	httpListenerName := fmt.Sprintf("kl-%s-listener-http", shortKey(installationKey))
	requestRoutingRuleName := fmt.Sprintf("kl-%s-rule", shortKey(installationKey))

	poller, err := client.BeginCreateOrUpdate(ctx, cfg.ResourceGroup, appGwName, armnetwork.ApplicationGateway{
		Location: &cfg.Location,
		Tags:     tags,
		Properties: &armnetwork.ApplicationGatewayPropertiesFormat{
			SKU: &armnetwork.ApplicationGatewaySKU{
				Name:     toAppGwSKUName(armnetwork.ApplicationGatewaySKUNameStandardV2),
				Tier:     toAppGwTier(armnetwork.ApplicationGatewayTierStandardV2),
				Capacity: int32Ptr(1),
			},
			GatewayIPConfigurations: []*armnetwork.ApplicationGatewayIPConfiguration{
				{
					Name: &gatewayIPConfigName,
					Properties: &armnetwork.ApplicationGatewayIPConfigurationPropertiesFormat{
						Subnet: &armnetwork.SubResource{
							ID: &appGwSubnetID,
						},
					},
				},
			},
			FrontendIPConfigurations: []*armnetwork.ApplicationGatewayFrontendIPConfiguration{
				{
					Name: &frontendIPConfigName,
					Properties: &armnetwork.ApplicationGatewayFrontendIPConfigurationPropertiesFormat{
						PublicIPAddress: &armnetwork.SubResource{
							ID: &publicIPID,
						},
					},
				},
			},
			FrontendPorts: []*armnetwork.ApplicationGatewayFrontendPort{
				{
					Name: &frontendPortHTTPName,
					Properties: &armnetwork.ApplicationGatewayFrontendPortPropertiesFormat{
						Port: int32Ptr(80),
					},
				},
			},
			BackendAddressPools: []*armnetwork.ApplicationGatewayBackendAddressPool{
				{
					Name:       &backendPoolName,
					Properties: &armnetwork.ApplicationGatewayBackendAddressPoolPropertiesFormat{
						// Backend addresses will be added later
					},
				},
			},
			BackendHTTPSettingsCollection: []*armnetwork.ApplicationGatewayBackendHTTPSettings{
				{
					Name: &backendSettingsName,
					Properties: &armnetwork.ApplicationGatewayBackendHTTPSettingsPropertiesFormat{
						Port:                int32Ptr(80),
						Protocol:            toAppGwProtocol(armnetwork.ApplicationGatewayProtocolHTTP),
						CookieBasedAffinity: toAppGwCookieAffinity(armnetwork.ApplicationGatewayCookieBasedAffinityDisabled),
						RequestTimeout:      int32Ptr(30),
					},
				},
			},
			HTTPListeners: []*armnetwork.ApplicationGatewayHTTPListener{
				{
					Name: &httpListenerName,
					Properties: &armnetwork.ApplicationGatewayHTTPListenerPropertiesFormat{
						FrontendIPConfiguration: &armnetwork.SubResource{
							ID: strPtr(fmt.Sprintf("/subscriptions/%s/resourceGroups/%s/providers/Microsoft.Network/applicationGateways/%s/frontendIPConfigurations/%s",
								cfg.SubscriptionID, cfg.ResourceGroup, appGwName, frontendIPConfigName)),
						},
						FrontendPort: &armnetwork.SubResource{
							ID: strPtr(fmt.Sprintf("/subscriptions/%s/resourceGroups/%s/providers/Microsoft.Network/applicationGateways/%s/frontendPorts/%s",
								cfg.SubscriptionID, cfg.ResourceGroup, appGwName, frontendPortHTTPName)),
						},
						Protocol: toAppGwProtocol(armnetwork.ApplicationGatewayProtocolHTTP),
					},
				},
			},
			RequestRoutingRules: []*armnetwork.ApplicationGatewayRequestRoutingRule{
				{
					Name: &requestRoutingRuleName,
					Properties: &armnetwork.ApplicationGatewayRequestRoutingRulePropertiesFormat{
						RuleType: toAppGwRuleType(armnetwork.ApplicationGatewayRequestRoutingRuleTypeBasic),
						Priority: int32Ptr(100),
						HTTPListener: &armnetwork.SubResource{
							ID: strPtr(fmt.Sprintf("/subscriptions/%s/resourceGroups/%s/providers/Microsoft.Network/applicationGateways/%s/httpListeners/%s",
								cfg.SubscriptionID, cfg.ResourceGroup, appGwName, httpListenerName)),
						},
						BackendAddressPool: &armnetwork.SubResource{
							ID: strPtr(fmt.Sprintf("/subscriptions/%s/resourceGroups/%s/providers/Microsoft.Network/applicationGateways/%s/backendAddressPools/%s",
								cfg.SubscriptionID, cfg.ResourceGroup, appGwName, backendPoolName)),
						},
						BackendHTTPSettings: &armnetwork.SubResource{
							ID: strPtr(fmt.Sprintf("/subscriptions/%s/resourceGroups/%s/providers/Microsoft.Network/applicationGateways/%s/backendHttpSettingsCollection/%s",
								cfg.SubscriptionID, cfg.ResourceGroup, appGwName, backendSettingsName)),
						},
					},
				},
			},
		},
	}, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create Application Gateway: %w", err)
	}

	result, err := poller.PollUntilDone(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to wait for Application Gateway creation: %w", err)
	}

	// Get the public IP address
	publicIP, err := GetAppGatewayPublicIP(ctx, cfg, *result.ID)
	if err != nil {
		publicIP = ""
	}

	return &AppGatewayInfo{
		ID:       *result.ID,
		PublicIP: publicIP,
	}, nil
}

// CreateBackendPool creates a backend pool for the Application Gateway
// This is a no-op for Azure as the pool is created with the gateway
func CreateBackendPool(ctx context.Context, cfg *AzureConfig, installationKey string) (string, error) {
	appGwName := fmt.Sprintf("kl-%s-appgw", shortKey(installationKey))
	backendPoolName := fmt.Sprintf("kl-%s-bepool", shortKey(installationKey))

	return fmt.Sprintf("/subscriptions/%s/resourceGroups/%s/providers/Microsoft.Network/applicationGateways/%s/backendAddressPools/%s",
		cfg.SubscriptionID, cfg.ResourceGroup, appGwName, backendPoolName), nil
}

// RegisterBackendTargets adds VM IP addresses to the backend pool
func RegisterBackendTargets(ctx context.Context, cfg *AzureConfig, backendPoolID string, instanceIDs ...string) error {
	client, err := armnetwork.NewApplicationGatewaysClient(cfg.SubscriptionID, cfg.Credential, nil)
	if err != nil {
		return fmt.Errorf("failed to create app gateway client: %w", err)
	}

	// Extract app gateway name from backend pool ID
	appGwName := extractAppGatewayName(backendPoolID)
	backendPoolName := extractResourceName(backendPoolID)

	// Get current Application Gateway configuration
	appGw, err := client.Get(ctx, cfg.ResourceGroup, appGwName, nil)
	if err != nil {
		return fmt.Errorf("failed to get Application Gateway: %w", err)
	}

	// Get VM IPs for the instances
	var backendAddresses []*armnetwork.ApplicationGatewayBackendAddress
	for _, vmID := range instanceIDs {
		vmName := extractResourceName(vmID)
		_, privateIP, err := getVMIPs(ctx, cfg, vmName)
		if err != nil {
			return fmt.Errorf("failed to get VM IP: %w", err)
		}
		backendAddresses = append(backendAddresses, &armnetwork.ApplicationGatewayBackendAddress{
			IPAddress: &privateIP,
		})
	}

	// Update the backend pool with the VM addresses
	for i, pool := range appGw.Properties.BackendAddressPools {
		if pool.Name != nil && *pool.Name == backendPoolName {
			appGw.Properties.BackendAddressPools[i].Properties.BackendAddresses = backendAddresses
			break
		}
	}

	// Update the Application Gateway
	poller, err := client.BeginCreateOrUpdate(ctx, cfg.ResourceGroup, appGwName, appGw.ApplicationGateway, nil)
	if err != nil {
		return fmt.Errorf("failed to update Application Gateway: %w", err)
	}

	_, err = poller.PollUntilDone(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to wait for Application Gateway update: %w", err)
	}

	return nil
}

// WaitForAppGatewayActive waits for the Application Gateway to be in running state
func WaitForAppGatewayActive(ctx context.Context, cfg *AzureConfig, appGwID string) error {
	client, err := armnetwork.NewApplicationGatewaysClient(cfg.SubscriptionID, cfg.Credential, nil)
	if err != nil {
		return fmt.Errorf("failed to create app gateway client: %w", err)
	}

	appGwName := extractResourceName(appGwID)

	timeout := 10 * time.Minute
	deadline := time.Now().Add(timeout)

	for time.Now().Before(deadline) {
		result, err := client.Get(ctx, cfg.ResourceGroup, appGwName, nil)
		if err != nil {
			return fmt.Errorf("failed to get Application Gateway: %w", err)
		}

		if result.Properties != nil && result.Properties.ProvisioningState != nil {
			if *result.Properties.ProvisioningState == armnetwork.ProvisioningStateSucceeded {
				return nil
			}
			if *result.Properties.ProvisioningState == armnetwork.ProvisioningStateFailed {
				return fmt.Errorf("Application Gateway provisioning failed")
			}
		}

		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(15 * time.Second):
			// Continue polling
		}
	}

	return fmt.Errorf("Application Gateway did not become active within %v", timeout)
}

// DeleteApplicationGateway deletes an Application Gateway
func DeleteApplicationGateway(ctx context.Context, cfg *AzureConfig, installationKey string) error {
	client, err := armnetwork.NewApplicationGatewaysClient(cfg.SubscriptionID, cfg.Credential, nil)
	if err != nil {
		return fmt.Errorf("failed to create app gateway client: %w", err)
	}

	appGwName := fmt.Sprintf("kl-%s-appgw", shortKey(installationKey))

	// Check if App Gateway exists
	_, err = client.Get(ctx, cfg.ResourceGroup, appGwName, nil)
	if err != nil {
		// App Gateway doesn't exist
		return nil
	}

	poller, err := client.BeginDelete(ctx, cfg.ResourceGroup, appGwName, nil)
	if err != nil {
		return fmt.Errorf("failed to start Application Gateway deletion: %w", err)
	}

	_, err = poller.PollUntilDone(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to delete Application Gateway: %w", err)
	}

	// Delete the App Gateway public IP
	return DeleteAppGatewayPublicIP(ctx, cfg, installationKey)
}

// DeleteAppGatewayPublicIP deletes the Application Gateway public IP
func DeleteAppGatewayPublicIP(ctx context.Context, cfg *AzureConfig, installationKey string) error {
	client, err := armnetwork.NewPublicIPAddressesClient(cfg.SubscriptionID, cfg.Credential, nil)
	if err != nil {
		return fmt.Errorf("failed to create public IP client: %w", err)
	}

	pipName := fmt.Sprintf("kl-%s-appgw-pip", installationKey)

	// Check if public IP exists
	_, err = client.Get(ctx, cfg.ResourceGroup, pipName, nil)
	if err != nil {
		// Public IP doesn't exist
		return nil
	}

	poller, err := client.BeginDelete(ctx, cfg.ResourceGroup, pipName, nil)
	if err != nil {
		return fmt.Errorf("failed to start App Gateway public IP deletion: %w", err)
	}

	_, err = poller.PollUntilDone(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to delete App Gateway public IP: %w", err)
	}

	return nil
}

// FindAppGatewayByInstallationKey finds an Application Gateway by installation key
func FindAppGatewayByInstallationKey(ctx context.Context, cfg *AzureConfig, installationKey string) (string, error) {
	client, err := armnetwork.NewApplicationGatewaysClient(cfg.SubscriptionID, cfg.Credential, nil)
	if err != nil {
		return "", fmt.Errorf("failed to create app gateway client: %w", err)
	}

	appGwName := fmt.Sprintf("kl-%s-appgw", shortKey(installationKey))

	result, err := client.Get(ctx, cfg.ResourceGroup, appGwName, nil)
	if err != nil {
		return "", nil // Not found
	}

	return *result.ID, nil
}

// FindBackendPoolByInstallationKey finds a backend pool by installation key
func FindBackendPoolByInstallationKey(ctx context.Context, cfg *AzureConfig, installationKey string) (string, error) {
	appGwID, err := FindAppGatewayByInstallationKey(ctx, cfg, installationKey)
	if err != nil || appGwID == "" {
		return "", err
	}

	backendPoolName := fmt.Sprintf("kl-%s-bepool", shortKey(installationKey))
	return fmt.Sprintf("%s/backendAddressPools/%s", appGwID, backendPoolName), nil
}

// GetAppGatewayPublicIP gets the public IP address of an Application Gateway
func GetAppGatewayPublicIP(ctx context.Context, cfg *AzureConfig, appGwID string) (string, error) {
	client, err := armnetwork.NewApplicationGatewaysClient(cfg.SubscriptionID, cfg.Credential, nil)
	if err != nil {
		return "", fmt.Errorf("failed to create app gateway client: %w", err)
	}

	pipClient, err := armnetwork.NewPublicIPAddressesClient(cfg.SubscriptionID, cfg.Credential, nil)
	if err != nil {
		return "", fmt.Errorf("failed to create public IP client: %w", err)
	}

	appGwName := extractResourceName(appGwID)

	appGw, err := client.Get(ctx, cfg.ResourceGroup, appGwName, nil)
	if err != nil {
		return "", fmt.Errorf("failed to get Application Gateway: %w", err)
	}

	// Find the public IP in the frontend IP configuration
	if appGw.Properties != nil && appGw.Properties.FrontendIPConfigurations != nil {
		for _, feIP := range appGw.Properties.FrontendIPConfigurations {
			if feIP.Properties != nil && feIP.Properties.PublicIPAddress != nil && feIP.Properties.PublicIPAddress.ID != nil {
				pipName := extractResourceName(*feIP.Properties.PublicIPAddress.ID)
				pip, err := pipClient.Get(ctx, cfg.ResourceGroup, pipName, nil)
				if err == nil && pip.Properties != nil && pip.Properties.IPAddress != nil {
					return *pip.Properties.IPAddress, nil
				}
			}
		}
	}

	return "", fmt.Errorf("public IP not found for Application Gateway")
}

// shortKey returns the first 8 characters of an installation key for resource naming
func shortKey(installationKey string) string {
	if len(installationKey) > 8 {
		return installationKey[:8]
	}
	return installationKey
}

// extractAppGatewayName extracts the app gateway name from a backend pool ID
func extractAppGatewayName(backendPoolID string) string {
	// Format: /subscriptions/{sub}/resourceGroups/{rg}/providers/Microsoft.Network/applicationGateways/{name}/backendAddressPools/{poolName}
	parts := splitString(backendPoolID, "/")
	for i, part := range parts {
		if part == "applicationGateways" && i+1 < len(parts) {
			return parts[i+1]
		}
	}
	return ""
}

// Helper functions for creating typed pointers

func toAppGwSKUName(n armnetwork.ApplicationGatewaySKUName) *armnetwork.ApplicationGatewaySKUName {
	return &n
}

func toAppGwTier(t armnetwork.ApplicationGatewayTier) *armnetwork.ApplicationGatewayTier {
	return &t
}

func toAppGwProtocol(p armnetwork.ApplicationGatewayProtocol) *armnetwork.ApplicationGatewayProtocol {
	return &p
}

func toAppGwCookieAffinity(c armnetwork.ApplicationGatewayCookieBasedAffinity) *armnetwork.ApplicationGatewayCookieBasedAffinity {
	return &c
}

func toAppGwRuleType(r armnetwork.ApplicationGatewayRequestRoutingRuleType) *armnetwork.ApplicationGatewayRequestRoutingRuleType {
	return &r
}
