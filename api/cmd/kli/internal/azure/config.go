package azure

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

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

	// Get subscription from env var or API
	subscriptionID := os.Getenv("AZURE_SUBSCRIPTION_ID")
	if subscriptionID == "" {
		var err error
		subscriptionID, err = getDefaultSubscription(ctx, cred)
		if err != nil {
			return nil, err
		}
	}

	// Get location from env var or flag
	if location == "" {
		location = GetDefaultLocation()
	}

	return &AzureConfig{
		Credential:     cred,
		SubscriptionID: subscriptionID,
		Location:       location,
		ResourceGroup:  resourceGroup,
	}, nil
}

// GetDefaultLocation returns location from environment variables, Azure config, or Azure IMDS
func GetDefaultLocation() string {
	// Try environment variables first
	envVars := []string{"AZURE_LOCATION", "AZURE_REGION", "AZURE_DEFAULTS_LOCATION"}
	for _, envVar := range envVars {
		if location := os.Getenv(envVar); location != "" {
			return location
		}
	}

	// Try reading from Azure CLI config file (~/.azure/config)
	location := readAzureConfig("defaults", "location")
	if location != "" {
		return location
	}

	// Try Azure Instance Metadata Service (IMDS)
	if location := getAzureIMDSLocation(); location != "" {
		return location
	}

	return ""
}

// getAzureIMDSLocation gets the location from Azure Instance Metadata Service
func getAzureIMDSLocation() string {
	client := &http.Client{Timeout: 1 * time.Second}

	req, err := http.NewRequest("GET", "http://169.254.169.254/metadata/instance/compute?api-version=2021-02-01", nil)
	if err != nil {
		return ""
	}
	req.Header.Set("Metadata", "true")

	resp, err := client.Do(req)
	if err != nil {
		return ""
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return ""
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return ""
	}

	// Parse JSON response to get location
	var metadata struct {
		Location string `json:"location"`
	}
	if err := json.Unmarshal(body, &metadata); err != nil {
		return ""
	}

	return metadata.Location
}

// readAzureConfig reads a value from Azure CLI config file
func readAzureConfig(section, key string) string {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return ""
	}

	configPath := filepath.Join(homeDir, ".azure", "config")
	return readAzureINIValue(configPath, section, key)
}

// readAzureINIValue reads a value from an INI-style config file
func readAzureINIValue(filePath, section, key string) string {
	file, err := os.Open(filePath)
	if err != nil {
		return ""
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	inTargetSection := false
	targetSection := "[" + section + "]"

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		// Skip comments and empty lines
		if line == "" || strings.HasPrefix(line, "#") || strings.HasPrefix(line, ";") {
			continue
		}

		// Check for section headers
		if strings.HasPrefix(line, "[") {
			inTargetSection = strings.EqualFold(line, targetSection)
			continue
		}

		// Look for key=value in target section
		if inTargetSection && strings.Contains(line, "=") {
			parts := strings.SplitN(line, "=", 2)
			if len(parts) == 2 && strings.TrimSpace(parts[0]) == key {
				return strings.TrimSpace(parts[1])
			}
		}
	}

	return ""
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
		"Name":                      &rgName,
		"ManagedBy":                 strPtr("kloudlite"),
		"Project":                   strPtr("kloudlite"),
		"Purpose":                   strPtr("kloudlite-installation"),
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
