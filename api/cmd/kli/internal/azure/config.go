package azure

import (
	"context"
	"fmt"

	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/resources/armresources"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/subscription/armsubscription"
)

// AzureConfig holds Azure configuration for the provider
type AzureConfig struct {
	Credential     *azidentity.DefaultAzureCredential
	SubscriptionID string
	Location       string
	ResourceGroup  string
}

// LoadAzureConfig creates an Azure configuration with credentials and subscription
func LoadAzureConfig(ctx context.Context, location, resourceGroup string) (*AzureConfig, error) {
	// Create DefaultAzureCredential
	cred, err := azidentity.NewDefaultAzureCredential(nil)
	if err != nil {
		return nil, fmt.Errorf("failed to obtain Azure credentials: %w", err)
	}

	// Get first enabled subscription
	subscriptionID, err := getDefaultSubscription(ctx, cred)
	if err != nil {
		return nil, err
	}

	// Use default location if not specified
	if location == "" {
		location = "eastus"
	}

	return &AzureConfig{
		Credential:     cred,
		SubscriptionID: subscriptionID,
		Location:       location,
		ResourceGroup:  resourceGroup,
	}, nil
}

// getDefaultSubscription retrieves the first enabled subscription
func getDefaultSubscription(ctx context.Context, cred *azidentity.DefaultAzureCredential) (string, error) {
	subsClient, err := armsubscription.NewSubscriptionsClient(cred, nil)
	if err != nil {
		return "", fmt.Errorf("failed to create subscriptions client: %w", err)
	}

	pager := subsClient.NewListPager(nil)
	for pager.More() {
		page, err := pager.NextPage(ctx)
		if err != nil {
			return "", fmt.Errorf("failed to list subscriptions: %w", err)
		}

		for _, sub := range page.Value {
			if sub.State != nil && *sub.State == armsubscription.SubscriptionStateEnabled {
				if sub.SubscriptionID != nil {
					return *sub.SubscriptionID, nil
				}
			}
		}
	}

	return "", fmt.Errorf("no enabled subscription found")
}

// EnsureResourceGroup creates a resource group if it doesn't exist
func EnsureResourceGroup(ctx context.Context, cfg *AzureConfig, installationKey string) (string, error) {
	rgName := cfg.ResourceGroup
	if rgName == "" {
		rgName = fmt.Sprintf("kl-%s-rg", installationKey)
	}

	client, err := armresources.NewResourceGroupsClient(cfg.SubscriptionID, cfg.Credential, nil)
	if err != nil {
		return "", fmt.Errorf("failed to create resource groups client: %w", err)
	}

	// Check if resource group exists
	_, err = client.Get(ctx, rgName, nil)
	if err == nil {
		// Resource group already exists
		return rgName, nil
	}

	// Create resource group
	tags := map[string]*string{
		"Name":                         &rgName,
		"ManagedBy":                    strPtr("kloudlite"),
		"Project":                      strPtr("kloudlite"),
		"Purpose":                      strPtr("kloudlite-installation"),
		"kloudlite-installation-id": &installationKey,
	}

	_, err = client.CreateOrUpdate(ctx, rgName, armresources.ResourceGroup{
		Location: &cfg.Location,
		Tags:     tags,
	}, nil)
	if err != nil {
		return "", fmt.Errorf("failed to create resource group: %w", err)
	}

	return rgName, nil
}

// DeleteResourceGroup deletes a resource group and all its resources
func DeleteResourceGroup(ctx context.Context, cfg *AzureConfig, rgName string) error {
	client, err := armresources.NewResourceGroupsClient(cfg.SubscriptionID, cfg.Credential, nil)
	if err != nil {
		return fmt.Errorf("failed to create resource groups client: %w", err)
	}

	// Check if resource group exists
	_, err = client.Get(ctx, rgName, nil)
	if err != nil {
		// Resource group doesn't exist
		return nil
	}

	// Delete resource group (this deletes all resources in it)
	poller, err := client.BeginDelete(ctx, rgName, nil)
	if err != nil {
		return fmt.Errorf("failed to start resource group deletion: %w", err)
	}

	// Wait for deletion to complete
	_, err = poller.PollUntilDone(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to delete resource group: %w", err)
	}

	return nil
}

// strPtr returns a pointer to a string
func strPtr(s string) *string {
	return &s
}

// int32Ptr returns a pointer to an int32
func int32Ptr(i int32) *int32 {
	return &i
}

// boolPtr returns a pointer to a bool
func boolPtr(b bool) *bool {
	return &b
}
