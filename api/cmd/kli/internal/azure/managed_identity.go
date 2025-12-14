package azure

import (
	"context"
	"fmt"
	"time"

	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/authorization/armauthorization/v2"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/msi/armmsi"
	"github.com/google/uuid"
)

// Azure built-in role definition IDs
const (
	// Storage Blob Data Contributor - for backup operations
	StorageBlobDataContributorRoleID = "ba92f5b4-2d11-453d-a403-e96b0029c9fe"
	// Virtual Machine Contributor - for VM lifecycle management
	VirtualMachineContributorRoleID = "9980e02c-c2be-4d73-94e8-173b1dc7cf3c"
	// Network Contributor - for NIC and Public IP management
	NetworkContributorRoleID = "4d97b98b-1d4f-4787-a291-c67834d212e7"
)

// EnsureManagedIdentity creates a User-Assigned Managed Identity if it doesn't exist
func EnsureManagedIdentity(ctx context.Context, cfg *AzureConfig, installationKey string) (string, string, error) {
	client, err := armmsi.NewUserAssignedIdentitiesClient(cfg.SubscriptionID, cfg.Credential, nil)
	if err != nil {
		return "", "", fmt.Errorf("failed to create managed identity client: %w", err)
	}

	identityName := fmt.Sprintf("kl-%s-identity", installationKey)

	// Check if identity already exists
	existing, err := client.Get(ctx, cfg.ResourceGroup, identityName, nil)
	if err == nil {
		return *existing.ID, *existing.Properties.PrincipalID, nil
	}

	// Create managed identity
	tags := map[string]*string{
		"Name":                      &identityName,
		"ManagedBy":                 strPtr("kloudlite"),
		"Project":                   strPtr("kloudlite"),
		"Purpose":                   strPtr("kloudlite-installation"),
		"kloudlite-installation-id": &installationKey,
	}

	result, err := client.CreateOrUpdate(ctx, cfg.ResourceGroup, identityName, armmsi.Identity{
		Location: &cfg.Location,
		Tags:     tags,
	}, nil)
	if err != nil {
		return "", "", fmt.Errorf("failed to create managed identity: %w", err)
	}

	return *result.ID, *result.Properties.PrincipalID, nil
}

// AssignStorageBlobRole assigns the Storage Blob Data Contributor role to the managed identity
func AssignStorageBlobRole(ctx context.Context, cfg *AzureConfig, principalID, storageAccountID string) error {
	client, err := armauthorization.NewRoleAssignmentsClient(cfg.SubscriptionID, cfg.Credential, nil)
	if err != nil {
		return fmt.Errorf("failed to create role assignments client: %w", err)
	}

	// Generate a unique name for the role assignment
	roleAssignmentName := uuid.New().String()

	// Build the role definition ID
	roleDefinitionID := fmt.Sprintf("/subscriptions/%s/providers/Microsoft.Authorization/roleDefinitions/%s",
		cfg.SubscriptionID, StorageBlobDataContributorRoleID)

	// Check if role assignment already exists by listing assignments for the principal
	pager := client.NewListForScopePager(storageAccountID, &armauthorization.RoleAssignmentsClientListForScopeOptions{
		Filter: strPtr(fmt.Sprintf("principalId eq '%s'", principalID)),
	})

	for pager.More() {
		page, err := pager.NextPage(ctx)
		if err != nil {
			break // Continue to create assignment
		}

		for _, assignment := range page.Value {
			if assignment.Properties != nil && assignment.Properties.RoleDefinitionID != nil {
				if *assignment.Properties.RoleDefinitionID == roleDefinitionID {
					// Role assignment already exists
					return nil
				}
			}
		}
	}

	// Create role assignment
	_, err = client.Create(ctx, storageAccountID, roleAssignmentName, armauthorization.RoleAssignmentCreateParameters{
		Properties: &armauthorization.RoleAssignmentProperties{
			PrincipalID:      &principalID,
			RoleDefinitionID: &roleDefinitionID,
			PrincipalType:    toServicePrincipal(),
		},
	}, nil)
	if err != nil {
		// Check if it's a conflict error (role already assigned)
		if contains(err.Error(), "RoleAssignmentExists") {
			return nil
		}
		return fmt.Errorf("failed to create role assignment: %w", err)
	}

	// Wait a bit for the role assignment to propagate
	time.Sleep(5 * time.Second)

	return nil
}

// AssignVMAndNetworkRoles assigns Virtual Machine Contributor and Network Contributor roles
// to the managed identity on the resource group scope (needed for WorkMachine controller)
func AssignVMAndNetworkRoles(ctx context.Context, cfg *AzureConfig, principalID string) error {
	client, err := armauthorization.NewRoleAssignmentsClient(cfg.SubscriptionID, cfg.Credential, nil)
	if err != nil {
		return fmt.Errorf("failed to create role assignments client: %w", err)
	}

	// Resource group scope
	rgScope := fmt.Sprintf("/subscriptions/%s/resourceGroups/%s", cfg.SubscriptionID, cfg.ResourceGroup)

	roles := []struct {
		name string
		id   string
	}{
		{"Virtual Machine Contributor", VirtualMachineContributorRoleID},
		{"Network Contributor", NetworkContributorRoleID},
	}

	for _, role := range roles {
		roleDefinitionID := fmt.Sprintf("/subscriptions/%s/providers/Microsoft.Authorization/roleDefinitions/%s",
			cfg.SubscriptionID, role.id)

		// Check if role assignment already exists
		exists := false
		pager := client.NewListForScopePager(rgScope, &armauthorization.RoleAssignmentsClientListForScopeOptions{
			Filter: strPtr(fmt.Sprintf("principalId eq '%s'", principalID)),
		})

		for pager.More() {
			page, err := pager.NextPage(ctx)
			if err != nil {
				break
			}
			for _, assignment := range page.Value {
				if assignment.Properties != nil && assignment.Properties.RoleDefinitionID != nil {
					if *assignment.Properties.RoleDefinitionID == roleDefinitionID {
						exists = true
						break
					}
				}
			}
			if exists {
				break
			}
		}

		if exists {
			continue
		}

		// Create role assignment
		roleAssignmentName := uuid.New().String()
		_, err = client.Create(ctx, rgScope, roleAssignmentName, armauthorization.RoleAssignmentCreateParameters{
			Properties: &armauthorization.RoleAssignmentProperties{
				PrincipalID:      &principalID,
				RoleDefinitionID: &roleDefinitionID,
				PrincipalType:    toServicePrincipal(),
			},
		}, nil)
		if err != nil {
			if !contains(err.Error(), "RoleAssignmentExists") {
				return fmt.Errorf("failed to assign %s role: %w", role.name, err)
			}
		}
	}

	// Wait for role assignments to propagate
	time.Sleep(5 * time.Second)

	return nil
}

// DeleteManagedIdentity deletes a User-Assigned Managed Identity
func DeleteManagedIdentity(ctx context.Context, cfg *AzureConfig, installationKey string) error {
	client, err := armmsi.NewUserAssignedIdentitiesClient(cfg.SubscriptionID, cfg.Credential, nil)
	if err != nil {
		return fmt.Errorf("failed to create managed identity client: %w", err)
	}

	identityName := fmt.Sprintf("kl-%s-identity", installationKey)

	// Check if identity exists
	_, err = client.Get(ctx, cfg.ResourceGroup, identityName, nil)
	if err != nil {
		// Identity doesn't exist
		return nil
	}

	_, err = client.Delete(ctx, cfg.ResourceGroup, identityName, nil)
	if err != nil {
		return fmt.Errorf("failed to delete managed identity: %w", err)
	}

	return nil
}

// FindManagedIdentityByInstallationKey finds a managed identity by installation key
func FindManagedIdentityByInstallationKey(ctx context.Context, cfg *AzureConfig, installationKey string) (string, error) {
	client, err := armmsi.NewUserAssignedIdentitiesClient(cfg.SubscriptionID, cfg.Credential, nil)
	if err != nil {
		return "", fmt.Errorf("failed to create managed identity client: %w", err)
	}

	identityName := fmt.Sprintf("kl-%s-identity", installationKey)

	result, err := client.Get(ctx, cfg.ResourceGroup, identityName, nil)
	if err != nil {
		return "", nil // Not found
	}

	return *result.ID, nil
}

func toServicePrincipal() *armauthorization.PrincipalType {
	pt := armauthorization.PrincipalTypeServicePrincipal
	return &pt
}
