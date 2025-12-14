package azure

import (
	"context"
	"fmt"
	"strings"

	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/storage/armstorage"
)

// EnsureStorageAccount creates a Storage Account if it doesn't exist
func EnsureStorageAccount(ctx context.Context, cfg *AzureConfig, installationKey string) (string, string, error) {
	client, err := armstorage.NewAccountsClient(cfg.SubscriptionID, cfg.Credential, nil)
	if err != nil {
		return "", "", fmt.Errorf("failed to create storage accounts client: %w", err)
	}

	// Storage account names must be 3-24 characters, lowercase letters and numbers only
	// Remove hyphens and truncate
	accountName := sanitizeStorageAccountName(installationKey)

	// Check if storage account already exists
	existing, err := client.GetProperties(ctx, cfg.ResourceGroup, accountName, nil)
	if err == nil {
		return *existing.ID, accountName, nil
	}

	// Create storage account
	tags := map[string]*string{
		"Name":                         &accountName,
		"ManagedBy":                    strPtr("kloudlite"),
		"Project":                      strPtr("kloudlite"),
		"Purpose":                      strPtr("kloudlite-installation"),
		"kloudlite-installation-id": &installationKey,
	}

	poller, err := client.BeginCreate(ctx, cfg.ResourceGroup, accountName, armstorage.AccountCreateParameters{
		Location: &cfg.Location,
		Tags:     tags,
		Kind:     toStorageKind(armstorage.KindStorageV2),
		SKU: &armstorage.SKU{
			Name: toSKUName(armstorage.SKUNameStandardLRS),
		},
		Properties: &armstorage.AccountPropertiesCreateParameters{
			AccessTier:            toAccessTier(armstorage.AccessTierHot),
			AllowBlobPublicAccess: boolPtr(false),
			MinimumTLSVersion:     toMinimumTLSVersion(armstorage.MinimumTLSVersionTLS12),
		},
	}, nil)
	if err != nil {
		return "", "", fmt.Errorf("failed to create storage account: %w", err)
	}

	result, err := poller.PollUntilDone(ctx, nil)
	if err != nil {
		return "", "", fmt.Errorf("failed to wait for storage account creation: %w", err)
	}

	return *result.ID, accountName, nil
}

// EnsureBlobContainer creates a blob container for K3s backups if it doesn't exist
func EnsureBlobContainer(ctx context.Context, cfg *AzureConfig, accountName string) (string, error) {
	client, err := armstorage.NewBlobContainersClient(cfg.SubscriptionID, cfg.Credential, nil)
	if err != nil {
		return "", fmt.Errorf("failed to create blob containers client: %w", err)
	}

	containerName := "k3s-backups"

	// Check if container already exists
	existing, err := client.Get(ctx, cfg.ResourceGroup, accountName, containerName, nil)
	if err == nil {
		return *existing.ID, nil
	}

	// Create container
	result, err := client.Create(ctx, cfg.ResourceGroup, accountName, containerName, armstorage.BlobContainer{
		ContainerProperties: &armstorage.ContainerProperties{
			PublicAccess: toPublicAccess(armstorage.PublicAccessNone),
		},
	}, nil)
	if err != nil {
		return "", fmt.Errorf("failed to create blob container: %w", err)
	}

	return *result.ID, nil
}

// EnableBlobVersioning enables blob versioning on the storage account
func EnableBlobVersioning(ctx context.Context, cfg *AzureConfig, accountName string) error {
	client, err := armstorage.NewBlobServicesClient(cfg.SubscriptionID, cfg.Credential, nil)
	if err != nil {
		return fmt.Errorf("failed to create blob services client: %w", err)
	}

	props := armstorage.BlobServiceProperties{}
	props.BlobServiceProperties = &armstorage.BlobServicePropertiesProperties{
		IsVersioningEnabled: boolPtr(true),
		DeleteRetentionPolicy: &armstorage.DeleteRetentionPolicy{
			Enabled: boolPtr(true),
			Days:    int32Ptr(7),
		},
	}

	_, err = client.SetServiceProperties(ctx, cfg.ResourceGroup, accountName, props, nil)
	if err != nil {
		return fmt.Errorf("failed to enable blob versioning: %w", err)
	}

	return nil
}

// DeleteStorageAccount deletes a Storage Account and all its contents
func DeleteStorageAccount(ctx context.Context, cfg *AzureConfig, accountName string) error {
	client, err := armstorage.NewAccountsClient(cfg.SubscriptionID, cfg.Credential, nil)
	if err != nil {
		return fmt.Errorf("failed to create storage accounts client: %w", err)
	}

	// Check if storage account exists
	_, err = client.GetProperties(ctx, cfg.ResourceGroup, accountName, nil)
	if err != nil {
		// Storage account doesn't exist
		return nil
	}

	_, err = client.Delete(ctx, cfg.ResourceGroup, accountName, nil)
	if err != nil {
		return fmt.Errorf("failed to delete storage account: %w", err)
	}

	return nil
}

// FindStorageAccountByInstallationKey finds a storage account by installation key
func FindStorageAccountByInstallationKey(ctx context.Context, cfg *AzureConfig, installationKey string) (string, string, error) {
	client, err := armstorage.NewAccountsClient(cfg.SubscriptionID, cfg.Credential, nil)
	if err != nil {
		return "", "", fmt.Errorf("failed to create storage accounts client: %w", err)
	}

	accountName := sanitizeStorageAccountName(installationKey)

	result, err := client.GetProperties(ctx, cfg.ResourceGroup, accountName, nil)
	if err != nil {
		return "", "", nil // Not found
	}

	return *result.ID, accountName, nil
}

// GetStorageAccountKey retrieves the primary access key for a storage account
func GetStorageAccountKey(ctx context.Context, cfg *AzureConfig, accountName string) (string, error) {
	client, err := armstorage.NewAccountsClient(cfg.SubscriptionID, cfg.Credential, nil)
	if err != nil {
		return "", fmt.Errorf("failed to create storage accounts client: %w", err)
	}

	result, err := client.ListKeys(ctx, cfg.ResourceGroup, accountName, nil)
	if err != nil {
		return "", fmt.Errorf("failed to list storage account keys: %w", err)
	}

	if result.Keys != nil && len(result.Keys) > 0 {
		return *result.Keys[0].Value, nil
	}

	return "", fmt.Errorf("no keys found for storage account %s", accountName)
}

// sanitizeStorageAccountName creates a valid storage account name from installation key
func sanitizeStorageAccountName(installationKey string) string {
	// Storage account names: 3-24 chars, lowercase letters and numbers only
	name := "kl" + strings.ToLower(installationKey) + "backup"

	// Remove any non-alphanumeric characters
	var result strings.Builder
	for _, c := range name {
		if (c >= 'a' && c <= 'z') || (c >= '0' && c <= '9') {
			result.WriteRune(c)
		}
	}

	sanitized := result.String()

	// Ensure length is between 3 and 24
	if len(sanitized) < 3 {
		sanitized = sanitized + "storage"
	}
	if len(sanitized) > 24 {
		sanitized = sanitized[:24]
	}

	return sanitized
}

// Helper functions for creating typed pointers

func toStorageKind(k armstorage.Kind) *armstorage.Kind {
	return &k
}

func toSKUName(s armstorage.SKUName) *armstorage.SKUName {
	return &s
}

func toAccessTier(t armstorage.AccessTier) *armstorage.AccessTier {
	return &t
}

func toMinimumTLSVersion(v armstorage.MinimumTLSVersion) *armstorage.MinimumTLSVersion {
	return &v
}

func toPublicAccess(p armstorage.PublicAccess) *armstorage.PublicAccess {
	return &p
}
