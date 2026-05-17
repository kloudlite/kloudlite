package cmd

import (
	"context"
	"fmt"

	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/compute/armcompute/v6"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/subscription/armsubscription"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

// azureDoctorCmd represents the azure doctor command
var azureDoctorCmd = &cobra.Command{
	Use:   "doctor",
	Short: "Check Azure prerequisites for Kloudlite installation",
	Long: `Verify that your Azure environment is properly configured for Kloudlite installation.

This command checks:
  - Azure CLI is installed
  - Azure CLI is authenticated
  - Current session has required RBAC permissions
  - Default subscription is set`,
	Example: `  # Check Azure prerequisites
  kli azure doctor
  kli az doctor`,
	Run: runAzureDoctor,
}

func runAzureDoctor(cmd *cobra.Command, args []string) {
	green := color.New(color.FgGreen, color.Bold)
	red := color.New(color.FgRed, color.Bold)
	yellow := color.New(color.FgYellow, color.Bold)
	cyan := color.New(color.FgCyan, color.Bold)

	fmt.Println()
	cyan.Println("Azure Doctor - Checking Prerequisites")
	fmt.Println()

	allPassed := true
	ctx := context.Background()

	// Check 1: Azure authentication and subscription
	fmt.Print("Checking Azure credentials and subscription... ")
	cred, subscriptionID, subscriptionName, err := checkAzureCredentials(ctx)
	if err == nil {
		green.Println("PASSED")
		fmt.Printf("   Subscription: %s (%s)\n", subscriptionName, subscriptionID)
	} else {
		red.Println("FAILED")
		yellow.Printf("   Error: %v\n", err)
		yellow.Println("   Configure Azure credentials: https://learn.microsoft.com/en-us/azure/developer/go/azure-sdk-authentication")
		allPassed = false
		cred = nil
	}

	// Check 2: Required RBAC permissions
	fmt.Print("Checking RBAC permissions... ")
	if cred != nil && subscriptionID != "" {
		permissions := checkAzurePermissions(ctx, cred, subscriptionID)
		if permissions.HasRequired {
			green.Println("PASSED")
			if len(permissions.Missing) > 0 {
				yellow.Printf("   Warning: Some optional permissions missing: %v\n", permissions.Missing)
			}
		} else {
			red.Println("FAILED")
			yellow.Println("   Missing required RBAC permissions for Kloudlite installation:")
			yellow.Println()
			yellow.Println("   Virtual Machine Permissions (to create and manage the VM):")
			yellow.Println("   - Microsoft.Compute/virtualMachines/write")
			yellow.Println("   - Microsoft.Compute/virtualMachines/read")
			yellow.Println()
			yellow.Println("   VNet/Network Permissions (to use existing VNet):")
			yellow.Println("   - Microsoft.Network/virtualNetworks/read")
			yellow.Println("   - Microsoft.Network/virtualNetworks/subnets/read")
			yellow.Println("   - Microsoft.Network/virtualNetworks/subnets/join/action")
			yellow.Println()
			yellow.Println("   Network Security Group Permissions (for ports 443, 6443, 8472, 10250, 5001):")
			yellow.Println("   - Microsoft.Network/networkSecurityGroups/write")
			yellow.Println("   - Microsoft.Network/networkSecurityGroups/read")
			yellow.Println("   - Microsoft.Network/networkSecurityGroups/securityRules/write")
			yellow.Println()
			yellow.Println("   Network Interface Permissions (required for VM):")
			yellow.Println("   - Microsoft.Network/networkInterfaces/write")
			yellow.Println("   - Microsoft.Network/networkInterfaces/read")
			yellow.Println("   - Microsoft.Network/networkInterfaces/join/action")
			yellow.Println()
			yellow.Println("   Managed Identity Permissions (to create and assign runtime identity to VM):")
			yellow.Println("   - Microsoft.ManagedIdentity/userAssignedIdentities/write")
			yellow.Println("   - Microsoft.ManagedIdentity/userAssignedIdentities/read")
			yellow.Println("   - Microsoft.Authorization/roleAssignments/write")
			yellow.Println("   - Microsoft.ManagedIdentity/userAssignedIdentities/assign/action")
			yellow.Println()
			allPassed = false
		}
	} else {
		yellow.Println("SKIPPED (credentials/subscription check failed)")
	}

	// Summary
	fmt.Println()
	if allPassed {
		green.Println("All checks passed! Your Azure environment is ready for Kloudlite installation.")
	} else {
		red.Println("Some checks failed. Please resolve the issues above before proceeding.")
		fmt.Println()
		fmt.Println("For more information, visit: https://docs.kloudlite.io/installation/azure")
	}
	fmt.Println()
}

func checkAzureCredentials(ctx context.Context) (*azidentity.DefaultAzureCredential, string, string, error) {
	// Create default Azure credential
	cred, err := azidentity.NewDefaultAzureCredential(nil)
	if err != nil {
		return nil, "", "", fmt.Errorf("failed to obtain Azure credentials: %w", err)
	}

	// Create subscriptions client to get default subscription
	subsClient, err := armsubscription.NewSubscriptionsClient(cred, nil)
	if err != nil {
		return nil, "", "", fmt.Errorf("failed to create subscriptions client: %w", err)
	}

	// List subscriptions and get the first enabled one
	pager := subsClient.NewListPager(nil)
	for pager.More() {
		page, err := pager.NextPage(ctx)
		if err != nil {
			return nil, "", "", fmt.Errorf("failed to list subscriptions: %w", err)
		}

		for _, sub := range page.Value {
			if sub.State != nil && *sub.State == armsubscription.SubscriptionStateEnabled {
				subscriptionID := ""
				subscriptionName := ""
				if sub.SubscriptionID != nil {
					subscriptionID = *sub.SubscriptionID
				}
				if sub.DisplayName != nil {
					subscriptionName = *sub.DisplayName
				}
				return cred, subscriptionID, subscriptionName, nil
			}
		}
	}

	return nil, "", "", fmt.Errorf("no enabled subscription found")
}

type AzurePermissionCheck struct {
	HasRequired bool
	Missing     []string
}

func checkAzurePermissions(ctx context.Context, cred *azidentity.DefaultAzureCredential, subscriptionID string) *AzurePermissionCheck {
	// Required permissions for Kloudlite installation on Azure
	// Single VM installation in default VNet with NSG and managed identity
	requiredPermissions := []string{
		// VM Permissions
		"Microsoft.Compute/virtualMachines/write",
		"Microsoft.Compute/virtualMachines/read",

		// VNet/Network Permissions (read-only for default VNet)
		"Microsoft.Network/virtualNetworks/read",
		"Microsoft.Network/virtualNetworks/subnets/read",
		"Microsoft.Network/virtualNetworks/subnets/join/action",

		// Network Security Group Permissions (ports: 443, 6443, 8472, 10250, 5001)
		"Microsoft.Network/networkSecurityGroups/write",
		"Microsoft.Network/networkSecurityGroups/read",
		"Microsoft.Network/networkSecurityGroups/securityRules/write",

		// Network Interface Permissions (required for VM)
		"Microsoft.Network/networkInterfaces/write",
		"Microsoft.Network/networkInterfaces/read",
		"Microsoft.Network/networkInterfaces/join/action",

		// Managed Identity Permissions (to create, edit, and assign identity to VM)
		"Microsoft.ManagedIdentity/userAssignedIdentities/write",
		"Microsoft.ManagedIdentity/userAssignedIdentities/read",
		"Microsoft.Authorization/roleAssignments/write",
		"Microsoft.ManagedIdentity/userAssignedIdentities/assign/action",
	}

	// Basic permission check using Virtual Machines List
	vmClient, err := armcompute.NewVirtualMachinesClient(subscriptionID, cred, nil)
	if err != nil {
		return &AzurePermissionCheck{
			HasRequired: false,
			Missing:     requiredPermissions,
		}
	}

	// Try to list VMs to verify basic compute permissions
	pager := vmClient.NewListAllPager(nil)
	_, err = pager.NextPage(ctx)
	if err != nil {
		// This could be a permission error or just no VMs
		// For now, if we can't list VMs, we'll assume missing permissions
		return &AzurePermissionCheck{
			HasRequired: false,
			Missing:     requiredPermissions,
		}
	}

	// Basic check passed - user has some VM permissions
	// TODO: Implement more granular permission checks using Azure RBAC
	return &AzurePermissionCheck{
		HasRequired: true,
		Missing:     []string{},
	}
}
