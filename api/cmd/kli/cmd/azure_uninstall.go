package cmd

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	azureinternal "github.com/kloudlite/kloudlite/api/cmd/kli/internal/azure"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

var azureUninstallCmd = &cobra.Command{
	Use:   "uninstall",
	Short: "Uninstall Kloudlite from Azure",
	Long: `Uninstall Kloudlite from Azure by deleting all associated resources.

This command will delete:
  - Application Gateway (if exists)
  - Azure VM and associated disks
  - Workmachine VMs, NICs, and Public IPs
  - Network Interface
  - Public IP addresses
  - Network Security Groups
  - User-Assigned Managed Identity
  - Storage Account and blob containers
  - Virtual Network and Subnets
  - Resource Group (if it was auto-created)

WARNING: This operation cannot be undone. All data will be permanently lost.`,
	Example: `  # Uninstall using defaults from ~/.azure/config
  kli azure uninstall --installation-key prod

  # Uninstall from a specific location
  kli azure uninstall --installation-key staging --location westus2

  # Uninstall from a specific resource group
  kli azure uninstall --installation-key dev --resource-group my-rg

  # Force delete resource group and all resources
  kli azure uninstall --installation-key dev --delete-resource-group`,
	Run: runAzureUninstall,
}

var (
	azureUninstallLocation        string
	azureUninstallResourceGroup   string
	azureUninstallInstallationKey string
	azureDeleteResourceGroup      bool
)

func init() {
	azureUninstallCmd.Flags().StringVar(&azureUninstallLocation, "location", "", "Azure location (reads from AZURE_LOCATION or ~/.azure/config)")
	azureUninstallCmd.Flags().StringVar(&azureUninstallResourceGroup, "resource-group", "", "Azure resource group (required if different from default)")
	azureUninstallCmd.Flags().StringVar(&azureUninstallInstallationKey, "installation-key", "", "Installation key to identify this installation (required)")
	azureUninstallCmd.Flags().BoolVar(&azureDeleteResourceGroup, "delete-resource-group", false, "Also delete the resource group (default: false)")
	azureUninstallCmd.MarkFlagRequired("installation-key")
}

func runAzureUninstall(cmd *cobra.Command, args []string) {
	green := color.New(color.FgGreen, color.Bold)
	red := color.New(color.FgRed, color.Bold)
	yellow := color.New(color.FgYellow, color.Bold)
	cyan := color.New(color.FgCyan, color.Bold)
	bold := color.New(color.Bold)

	// Header
	fmt.Println()
	cyan.Println("+-----------------------------------------+")
	cyan.Println("|   Kloudlite Azure Uninstallation        |")
	cyan.Println("+-----------------------------------------+")
	fmt.Println()

	// Warning about interruption
	yellow.Println("WARNING: This operation cannot be interrupted safely.")
	yellow.Println("         Interrupting may leave orphaned resources.")
	fmt.Println()

	ctx := context.Background()

	// Setup signal handling - warn but don't exit
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-sigChan
		fmt.Println()
		yellow.Println("\nInterrupt received, but continuing cleanup to avoid orphaned resources...")
		yellow.Println("Press Ctrl+C again to force exit (not recommended).")

		<-sigChan
		fmt.Println()
		red.Println("Force exit - some resources may be orphaned!")
		os.Exit(130)
	}()

	// Configuration
	bold.Println("Configuration")
	bold.Println("-------------")
	fmt.Printf("  Installation Key: %s\n", azureUninstallInstallationKey)
	fmt.Printf("  Location:         ")

	// If resource group not specified, use default naming
	if azureUninstallResourceGroup == "" {
		azureUninstallResourceGroup = fmt.Sprintf("kl-%s-rg", azureUninstallInstallationKey)
	}

	cfg, err := azureinternal.LoadAzureConfig(ctx, azureUninstallLocation, azureUninstallResourceGroup)
	if err != nil {
		red.Printf("x\n")
		yellow.Printf("  Error: %v\n\n", err)
		os.Exit(1)
	}
	green.Printf("+ %s\n", cfg.Location)
	fmt.Printf("  Subscription:     %s\n", cfg.SubscriptionID)
	fmt.Printf("  Resource Group:   %s\n", cfg.ResourceGroup)
	fmt.Println()

	// Start deletion
	bold.Println("Deleting Resources")
	bold.Println("------------------")

	var deletionErrors []error

	// Delete Application Gateway first (releases public IP)
	fmt.Printf("  o Deleting Application Gateway...")
	err = azureinternal.DeleteApplicationGateway(ctx, cfg, azureUninstallInstallationKey)
	if err != nil {
		yellow.Printf(" (warning: %v)\n", err)
		deletionErrors = append(deletionErrors, err)
	} else {
		green.Printf(" +\n")
	}

	// Wait for Application Gateway deletion to propagate
	time.Sleep(5 * time.Second)

	// Delete VM (includes NIC and OS disk due to DeleteOption)
	fmt.Printf("  o Deleting VM...")
	err = azureinternal.TerminateVM(ctx, cfg, azureUninstallInstallationKey)
	if err != nil {
		yellow.Printf(" (warning: %v)\n", err)
		deletionErrors = append(deletionErrors, err)
	} else {
		green.Printf(" +\n")
	}

	// Wait for VM deletion to complete
	time.Sleep(5 * time.Second)

	// Delete Workmachine VMs, NICs, and PIPs
	fmt.Printf("  o Deleting Workmachine resources...")
	wmCount, wmErr := azureinternal.DeleteWorkmachineResources(ctx, cfg)
	if wmErr != nil {
		yellow.Printf(" (warning: %v)\n", wmErr)
		deletionErrors = append(deletionErrors, wmErr)
	} else if wmCount > 0 {
		green.Printf(" + (%d resources)\n", wmCount)
	} else {
		green.Printf(" + (none found)\n")
	}

	time.Sleep(5 * time.Second)

	// Delete Network Interface (may already be deleted with VM)
	fmt.Printf("  o Deleting Network Interface...")
	err = azureinternal.DeleteNetworkInterface(ctx, cfg, azureUninstallInstallationKey)
	if err != nil {
		yellow.Printf(" (warning: %v)\n", err)
	} else {
		green.Printf(" +\n")
	}

	// Delete Public IP
	fmt.Printf("  o Deleting Public IP...")
	err = azureinternal.DeletePublicIP(ctx, cfg, azureUninstallInstallationKey)
	if err != nil {
		yellow.Printf(" (warning: %v)\n", err)
		deletionErrors = append(deletionErrors, err)
	} else {
		green.Printf(" +\n")
	}

	// Delete NSGs (retry if needed due to dependencies)
	fmt.Printf("  o Deleting master NSG...")
	for retries := 0; retries < 5; retries++ {
		err = azureinternal.DeleteNSGByName(ctx, cfg, fmt.Sprintf("kl-%s-master-nsg", azureUninstallInstallationKey))
		if err == nil {
			green.Printf(" +\n")
			break
		}
		if retries < 4 {
			time.Sleep(10 * time.Second)
		} else {
			yellow.Printf(" (warning: %v)\n", err)
		}
	}

	fmt.Printf("  o Deleting App Gateway NSG...")
	for retries := 0; retries < 5; retries++ {
		err = azureinternal.DeleteNSGByName(ctx, cfg, fmt.Sprintf("kl-%s-appgw-nsg", azureUninstallInstallationKey))
		if err == nil {
			green.Printf(" +\n")
			break
		}
		if retries < 4 {
			time.Sleep(10 * time.Second)
		} else {
			yellow.Printf(" (warning: %v)\n", err)
		}
	}

	fmt.Printf("  o Deleting instance NSG...")
	for retries := 0; retries < 5; retries++ {
		err = azureinternal.DeleteNSGByName(ctx, cfg, fmt.Sprintf("kl-%s-nsg", azureUninstallInstallationKey))
		if err == nil {
			green.Printf(" +\n")
			break
		}
		if retries < 4 {
			time.Sleep(10 * time.Second)
		} else {
			yellow.Printf(" (warning: %v)\n", err)
			deletionErrors = append(deletionErrors, err)
		}
	}

	// Delete Managed Identity
	fmt.Printf("  o Deleting Managed Identity...")
	err = azureinternal.DeleteManagedIdentity(ctx, cfg, azureUninstallInstallationKey)
	if err != nil {
		yellow.Printf(" (warning: %v)\n", err)
		deletionErrors = append(deletionErrors, err)
	} else {
		green.Printf(" +\n")
	}

	// Delete Storage Account (includes all blobs)
	fmt.Printf("  o Deleting Storage Account...")
	_, storageAccountName, _ := azureinternal.FindStorageAccountByInstallationKey(ctx, cfg, azureUninstallInstallationKey)
	if storageAccountName != "" {
		err = azureinternal.DeleteStorageAccount(ctx, cfg, storageAccountName)
		if err != nil {
			yellow.Printf(" (warning: %v)\n", err)
			deletionErrors = append(deletionErrors, err)
		} else {
			green.Printf(" +\n")
		}
	} else {
		yellow.Printf(" (not found)\n")
	}

	// Delete VNet (includes subnets)
	fmt.Printf("  o Deleting VNet...")
	err = azureinternal.DeleteVNet(ctx, cfg, azureUninstallInstallationKey)
	if err != nil {
		yellow.Printf(" (warning: %v)\n", err)
		deletionErrors = append(deletionErrors, err)
	} else {
		green.Printf(" +\n")
	}

	// Optionally delete resource group
	if azureDeleteResourceGroup {
		fmt.Printf("  o Deleting Resource Group (this may take a while)...")
		err = azureinternal.DeleteResourceGroup(ctx, cfg, cfg.ResourceGroup)
		if err != nil {
			yellow.Printf(" (warning: %v)\n", err)
			deletionErrors = append(deletionErrors, err)
		} else {
			green.Printf(" +\n")
		}
	}

	// Summary
	fmt.Println()
	if len(deletionErrors) > 0 {
		yellow.Println("+-----------------------------------------+")
		yellow.Println("|   Uninstallation completed with warnings|")
		yellow.Println("+-----------------------------------------+")
		fmt.Println()
		yellow.Println("Some resources may not have been deleted. Please check the Azure Portal")
		yellow.Println("for any remaining resources in the resource group:")
		fmt.Printf("  Resource Group: %s\n", cfg.ResourceGroup)
		fmt.Println()
		yellow.Println("You can manually delete the resource group to remove all resources:")
		cyan.Printf("  az group delete --name %s --yes\n", cfg.ResourceGroup)
	} else {
		green.Println("+-----------------------------------------+")
		green.Println("|   + Uninstallation Complete!            |")
		green.Println("+-----------------------------------------+")
		fmt.Println()
		fmt.Println("All Kloudlite resources have been successfully deleted.")
		if !azureDeleteResourceGroup {
			fmt.Println()
			fmt.Printf("The resource group '%s' was not deleted.\n", cfg.ResourceGroup)
			fmt.Println("To delete it and any remaining resources:")
			cyan.Printf("  az group delete --name %s --yes\n", cfg.ResourceGroup)
		}
	}
	fmt.Println()
}
