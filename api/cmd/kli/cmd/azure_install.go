package cmd

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	azureinternal "github.com/kloudlite/kloudlite/api/cmd/kli/internal/azure"
	"github.com/kloudlite/kloudlite/api/cmd/kli/internal/console"
	k8sinternal "github.com/kloudlite/kloudlite/api/cmd/kli/internal/k8s"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

var azureInstallCmd = &cobra.Command{
	Use:   "install",
	Short: "Install Kloudlite on Azure",
	Long: `Install Kloudlite on Azure by creating all necessary resources.

This command will:
  - Find Ubuntu 24.04 LTS image in the region
  - Create Resource Group (if not specified)
  - Create VNet and Subnets
  - Create Network Security Groups for VM and Application Gateway
  - Create User-Assigned Managed Identity with required permissions
  - Create Storage Account for K3s database backups
  - Launch Azure VM with 100GB Premium SSD
  - Setup automated K3s backup to Azure Blob Storage every 30 minutes
  - Create Application Gateway (HTTP listener)
  - Configure DNS with Cloudflare proxy mode (TLS termination at Cloudflare edge)

NOTE: The subdomain must be reserved in the console (console.kloudlite.io)
before running this command. The installation will fail if no subdomain
has been configured for the installation key.`,
	Example: `  # Install using defaults from ~/.azure/config
  kli azure install --installation-key prod

  # Install in a specific location
  kli azure install --installation-key staging --location westus2

  # Install in an existing resource group
  kli azure install --installation-key dev --resource-group my-rg

  # Install with custom VM size
  kli azure install --installation-key prod --vm-size Standard_D4s_v3

  # Install without Application Gateway (direct VM access only)
  kli azure install --installation-key dev --skip-appgw`,
	Run: runAzureInstall,
}

var (
	azureLocation                    string
	azureResourceGroup               string
	azureInstallationKey             string
	azureVMSize                      string
	azureEnableTerminationProtection bool
	azureSkipAppGw                   bool
)

func init() {
	azureInstallCmd.Flags().StringVar(&azureLocation, "location", "", "Azure location (reads from AZURE_LOCATION or ~/.azure/config)")
	azureInstallCmd.Flags().StringVar(&azureResourceGroup, "resource-group", "", "Azure resource group (auto-created if not specified)")
	azureInstallCmd.Flags().StringVar(&azureInstallationKey, "installation-key", "", "Installation key to identify this installation (required)")
	azureInstallCmd.Flags().StringVar(&azureVMSize, "vm-size", "Standard_B2ms", "Azure VM size (default: Standard_B2ms)")
	azureInstallCmd.Flags().BoolVar(&azureEnableTerminationProtection, "enable-delete-protection", true, "Enable VM delete protection (default: true)")
	azureInstallCmd.Flags().BoolVar(&azureSkipAppGw, "skip-appgw", false, "Skip Application Gateway setup (direct VM access only)")
	azureInstallCmd.MarkFlagRequired("installation-key")
}

func runAzureInstall(cmd *cobra.Command, args []string) {
	green := color.New(color.FgGreen, color.Bold)
	red := color.New(color.FgRed, color.Bold)
	yellow := color.New(color.FgYellow, color.Bold)
	cyan := color.New(color.FgCyan, color.Bold)
	bold := color.New(color.Bold)

	// Header
	fmt.Println()
	cyan.Println("+-----------------------------------------+")
	cyan.Println("|   Kloudlite Azure Installation          |")
	cyan.Println("+-----------------------------------------+")
	fmt.Println()

	ctx := context.Background()

	// Setup signal handling for cleanup on interruption
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	var createdResources struct {
		sync.Mutex
		cfg                *azureinternal.AzureConfig
		resourceGroup      string
		vnetID             string
		subnetID           string
		nsgID              string
		masterNsgID        string
		appGwNsgID         string
		identityID         string
		storageAccountID   string
		storageAccountName string
		publicIPID         string
		nicID              string
		vmID               string
		appGwID            string
	}

	go func() {
		<-sigChan
		fmt.Println()
		yellow.Println("\nInstallation interrupted! Cleaning up resources...")

		createdResources.Lock()
		defer createdResources.Unlock()

		if createdResources.cfg == nil {
			os.Exit(130)
		}

		cfg := createdResources.cfg
		cleanupCtx := context.Background()

		// Cleanup in reverse order
		if createdResources.appGwID != "" {
			fmt.Printf("  Deleting Application Gateway...\n")
			azureinternal.DeleteApplicationGateway(cleanupCtx, cfg, azureInstallationKey)
		}
		if createdResources.vmID != "" {
			fmt.Printf("  Terminating VM...\n")
			azureinternal.TerminateVM(cleanupCtx, cfg, azureInstallationKey)
		}
		if createdResources.nicID != "" {
			fmt.Printf("  Deleting Network Interface...\n")
			azureinternal.DeleteNetworkInterface(cleanupCtx, cfg, azureInstallationKey)
		}
		if createdResources.publicIPID != "" {
			fmt.Printf("  Deleting Public IP...\n")
			azureinternal.DeletePublicIP(cleanupCtx, cfg, azureInstallationKey)
		}
		if createdResources.masterNsgID != "" {
			fmt.Printf("  Deleting master NSG...\n")
			azureinternal.DeleteNSGByName(cleanupCtx, cfg, fmt.Sprintf("kl-%s-master-nsg", azureInstallationKey))
		}
		if createdResources.appGwNsgID != "" {
			fmt.Printf("  Deleting App Gateway NSG...\n")
			azureinternal.DeleteNSGByName(cleanupCtx, cfg, fmt.Sprintf("kl-%s-appgw-nsg", azureInstallationKey))
		}
		if createdResources.nsgID != "" {
			fmt.Printf("  Deleting NSG...\n")
			azureinternal.DeleteNSGByName(cleanupCtx, cfg, fmt.Sprintf("kl-%s-nsg", azureInstallationKey))
		}
		if createdResources.identityID != "" {
			fmt.Printf("  Deleting Managed Identity...\n")
			azureinternal.DeleteManagedIdentity(cleanupCtx, cfg, azureInstallationKey)
		}
		if createdResources.storageAccountName != "" {
			fmt.Printf("  Deleting Storage Account...\n")
			azureinternal.DeleteStorageAccount(cleanupCtx, cfg, createdResources.storageAccountName)
		}
		if createdResources.vnetID != "" {
			fmt.Printf("  Deleting VNet...\n")
			azureinternal.DeleteVNet(cleanupCtx, cfg, azureInstallationKey)
		}

		yellow.Println("Cleanup completed. Exiting...")
		os.Exit(130) // Standard exit code for SIGINT
	}()

	// Configuration
	bold.Println("Configuration")
	bold.Println("-------------")
	fmt.Printf("  Installation Key: %s\n", azureInstallationKey)
	fmt.Printf("  Location:         ")

	cfg, err := azureinternal.LoadAzureConfig(ctx, azureLocation, azureResourceGroup)
	if err != nil {
		red.Printf("x\n")
		yellow.Printf("  Error: %v\n\n", err)
		os.Exit(1)
	}
	green.Printf("+ %s\n", cfg.Location)
	fmt.Printf("  Subscription:     %s\n", cfg.SubscriptionID)
	fmt.Printf("  VM Size:          %s\n", azureVMSize)
	fmt.Println()

	createdResources.Lock()
	createdResources.cfg = cfg
	createdResources.Unlock()

	// Console API client
	consoleClient := console.NewClient()

	// Verify Installation and get subdomain
	bold.Println("Verifying Installation")
	bold.Println("----------------------")

	fmt.Printf("  o Verifying installation key with registration API...")
	verifyResult, err := k8sinternal.VerifyInstallation(ctx, azureInstallationKey)
	if err != nil {
		red.Printf(" x\n")
		yellow.Printf("    Error: %v\n\n", err)
		os.Exit(1)
	}
	green.Printf(" +\n")
	fmt.Printf("    Secret key obtained successfully\n")

	secretKey := verifyResult.SecretKey
	var fullDomain string

	// Check if subdomain was configured in console (required for App Gateway)
	if !azureSkipAppGw {
		if verifyResult.Subdomain == "" {
			red.Printf("\n  Error: No subdomain configured for this installation.\n")
			yellow.Printf("  Please configure a subdomain in the console (console.kloudlite.io)\n")
			yellow.Printf("  before running this installation command.\n\n")
			os.Exit(1)
		}

		fullDomain = console.GetFullDomain(verifyResult.Subdomain)
		fmt.Printf("    Subdomain: %s\n", verifyResult.Subdomain)
		cyan.Printf("    Your URL: https://%s\n", fullDomain)
	}
	fmt.Println()

	// Infrastructure Setup
	bold.Println("Infrastructure Setup")
	bold.Println("--------------------")

	// Create Resource Group
	fmt.Printf("  o Creating resource group...")
	resourceGroup, err := azureinternal.EnsureResourceGroup(ctx, cfg, azureInstallationKey)
	if err != nil {
		red.Printf(" x\n")
		yellow.Printf("    Error: %v\n\n", err)
		os.Exit(1)
	}
	cfg.ResourceGroup = resourceGroup
	createdResources.Lock()
	createdResources.resourceGroup = resourceGroup
	createdResources.Unlock()
	green.Printf(" +\n")
	fmt.Printf("    %s\n", resourceGroup)

	// Find Ubuntu Image
	fmt.Printf("  o Finding Ubuntu image...")
	imageRef, err := azureinternal.FindUbuntuImage(ctx, cfg)
	if err != nil {
		red.Printf(" x\n")
		yellow.Printf("    Error: %v\n\n", err)
		os.Exit(1)
	}
	green.Printf(" +\n")
	fmt.Printf("    %s\n", imageRef.String())

	// Pace API calls
	time.Sleep(1 * time.Second)

	// Network Resources
	fmt.Printf("  o Setting up network...")
	vnetID, vnetCIDR, err := azureinternal.EnsureVNet(ctx, cfg, azureInstallationKey)
	if err != nil {
		red.Printf(" x\n")
		yellow.Printf("    Error: %v\n\n", err)
		os.Exit(1)
	}
	createdResources.Lock()
	createdResources.vnetID = vnetID
	createdResources.Unlock()

	vnetName := azureinternal.ExtractResourceName(vnetID)
	subnetID, _, err := azureinternal.EnsureSubnet(ctx, cfg, vnetName, azureInstallationKey)
	if err != nil {
		red.Printf(" x\n")
		yellow.Printf("    Error: %v\n\n", err)
		os.Exit(1)
	}
	createdResources.Lock()
	createdResources.subnetID = subnetID
	createdResources.Unlock()

	green.Printf(" +\n")
	fmt.Printf("    VNet: kl-%s-vnet (%s)\n", azureInstallationKey, vnetCIDR)
	fmt.Printf("    Subnet: kl-%s-subnet\n", azureInstallationKey)

	// Pace API calls
	time.Sleep(1 * time.Second)

	// Parallel Resource Creation
	fmt.Printf("  o Creating resources in parallel...\n")

	var wg sync.WaitGroup
	var nsgID, appGwNsgID, identityID, principalID, storageAccountID, storageAccountName string
	var nsgErr, appGwNsgErr, identityErr, storageErr error

	startTime := time.Now()

	// NSG (parallel)
	wg.Add(1)
	go func() {
		defer wg.Done()
		fmt.Printf("    [%s] Starting: NSG creation\n", time.Now().Format("15:04:05"))
		nsgID, nsgErr = azureinternal.EnsureNetworkSecurityGroup(ctx, cfg, vnetCIDR, azureInstallationKey)
		if nsgErr != nil {
			fmt.Printf("    [%s] Failed: NSG - %v\n", time.Now().Format("15:04:05"), nsgErr)
		} else {
			createdResources.Lock()
			createdResources.nsgID = nsgID
			createdResources.Unlock()
			fmt.Printf("    [%s] Completed: NSG\n", time.Now().Format("15:04:05"))
		}
	}()

	// App Gateway NSG (parallel, only if not skipping App Gateway)
	if !azureSkipAppGw {
		wg.Add(1)
		go func() {
			defer wg.Done()
			fmt.Printf("    [%s] Starting: App Gateway NSG creation\n", time.Now().Format("15:04:05"))
			appGwNsgID, appGwNsgErr = azureinternal.CreateAppGatewayNSG(ctx, cfg, azureInstallationKey)
			if appGwNsgErr != nil {
				fmt.Printf("    [%s] Failed: App Gateway NSG - %v\n", time.Now().Format("15:04:05"), appGwNsgErr)
			} else {
				createdResources.Lock()
				createdResources.appGwNsgID = appGwNsgID
				createdResources.Unlock()
				fmt.Printf("    [%s] Completed: App Gateway NSG\n", time.Now().Format("15:04:05"))
			}
		}()
	}

	// Managed Identity (parallel)
	wg.Add(1)
	go func() {
		defer wg.Done()
		fmt.Printf("    [%s] Starting: Managed Identity creation\n", time.Now().Format("15:04:05"))
		identityID, principalID, identityErr = azureinternal.EnsureManagedIdentity(ctx, cfg, azureInstallationKey)
		if identityErr != nil {
			fmt.Printf("    [%s] Failed: Managed Identity - %v\n", time.Now().Format("15:04:05"), identityErr)
		} else {
			createdResources.Lock()
			createdResources.identityID = identityID
			createdResources.Unlock()
			fmt.Printf("    [%s] Completed: Managed Identity\n", time.Now().Format("15:04:05"))
		}
	}()

	// Storage Account (parallel)
	wg.Add(1)
	go func() {
		defer wg.Done()
		fmt.Printf("    [%s] Starting: Storage Account creation\n", time.Now().Format("15:04:05"))
		storageAccountID, storageAccountName, storageErr = azureinternal.EnsureStorageAccount(ctx, cfg, azureInstallationKey)
		if storageErr != nil {
			fmt.Printf("    [%s] Failed: Storage Account - %v\n", time.Now().Format("15:04:05"), storageErr)
		} else {
			createdResources.Lock()
			createdResources.storageAccountID = storageAccountID
			createdResources.storageAccountName = storageAccountName
			createdResources.Unlock()
			fmt.Printf("    [%s] Completed: Storage Account\n", time.Now().Format("15:04:05"))
		}
	}()

	wg.Wait()
	elapsed := time.Since(startTime)
	fmt.Printf("    Parallel operations completed in %.1fs\n", elapsed.Seconds())

	// Check for errors
	if nsgErr != nil {
		red.Printf(" x\n")
		yellow.Printf("    NSG Error: %v\n\n", nsgErr)
		os.Exit(1)
	}
	if !azureSkipAppGw && appGwNsgErr != nil {
		red.Printf(" x\n")
		yellow.Printf("    App Gateway NSG Error: %v\n\n", appGwNsgErr)
		os.Exit(1)
	}
	if identityErr != nil {
		red.Printf(" x\n")
		yellow.Printf("    Managed Identity Error: %v\n\n", identityErr)
		os.Exit(1)
	}
	if storageErr != nil {
		red.Printf(" x\n")
		yellow.Printf("    Storage Account Error: %v\n\n", storageErr)
		os.Exit(1)
	}

	green.Printf(" +\n")
	fmt.Printf("    NSG: kl-%s-nsg\n", azureInstallationKey)
	if !azureSkipAppGw {
		fmt.Printf("    App Gateway NSG: kl-%s-appgw-nsg\n", azureInstallationKey)
	}
	fmt.Printf("    Managed Identity: kl-%s-identity\n", azureInstallationKey)
	fmt.Printf("    Storage Account: %s\n", storageAccountName)

	// Pace API calls
	time.Sleep(2 * time.Second)

	// Assign Storage Blob role to Managed Identity
	fmt.Printf("  o Assigning Storage Blob role...")
	err = azureinternal.AssignStorageBlobRole(ctx, cfg, principalID, storageAccountID)
	if err != nil {
		red.Printf(" x\n")
		yellow.Printf("    Error: %v\n\n", err)
		os.Exit(1)
	}
	green.Printf(" +\n")

	// Assign VM and Network roles for WorkMachine controller
	fmt.Printf("  o Assigning VM & Network roles...")
	err = azureinternal.AssignVMAndNetworkRoles(ctx, cfg, principalID)
	if err != nil {
		red.Printf(" x\n")
		yellow.Printf("    Error: %v\n\n", err)
		os.Exit(1)
	}
	green.Printf(" +\n")

	// Create blob container
	fmt.Printf("  o Creating blob container...")
	_, err = azureinternal.EnsureBlobContainer(ctx, cfg, storageAccountName)
	if err != nil {
		red.Printf(" x\n")
		yellow.Printf("    Error: %v\n\n", err)
		os.Exit(1)
	}
	green.Printf(" +\n")

	// Enable blob versioning
	fmt.Printf("  o Enabling blob versioning...")
	err = azureinternal.EnableBlobVersioning(ctx, cfg, storageAccountName)
	if err != nil {
		yellow.Printf(" (warning: %v)\n", err)
	} else {
		green.Printf(" +\n")
	}

	// Create master NSG (depends on App Gateway NSG)
	var masterNsgID string
	if !azureSkipAppGw {
		fmt.Printf("  o Creating master NSG...")
		masterNsgID, err = azureinternal.EnsureMasterNSG(ctx, cfg, vnetCIDR, azureinternal.AppGatewaySubnetCIDR, azureInstallationKey)
		if err != nil {
			red.Printf(" x\n")
			yellow.Printf("    Error: %v\n\n", err)
			os.Exit(1)
		}
		createdResources.Lock()
		createdResources.masterNsgID = masterNsgID
		createdResources.Unlock()
		green.Printf(" +\n")
		fmt.Printf("    Master NSG: kl-%s-master-nsg\n", azureInstallationKey)
	}

	// Instance Deployment
	bold.Println("\nInstance Deployment")
	bold.Println("-------------------")

	// Generate K3s agent token
	k3sToken, err := azureinternal.GenerateK3sToken()
	if err != nil {
		red.Printf(" x\n")
		yellow.Printf("    Error generating K3s token: %v\n\n", err)
		os.Exit(1)
	}

	// Create public IP
	fmt.Printf("  o Creating public IP...")
	publicIPID, err := azureinternal.CreatePublicIP(ctx, cfg, azureInstallationKey)
	if err != nil {
		red.Printf(" x\n")
		yellow.Printf("    Error: %v\n\n", err)
		os.Exit(1)
	}
	createdResources.Lock()
	createdResources.publicIPID = publicIPID
	createdResources.Unlock()
	green.Printf(" +\n")

	// Use master NSG for VM when App Gateway is enabled
	instanceNsgID := nsgID
	if !azureSkipAppGw && masterNsgID != "" {
		instanceNsgID = masterNsgID
	}

	// Create network interface
	fmt.Printf("  o Creating network interface...")
	nicID, err := azureinternal.CreateNetworkInterface(ctx, cfg, subnetID, instanceNsgID, publicIPID, azureInstallationKey)
	if err != nil {
		red.Printf(" x\n")
		yellow.Printf("    Error: %v\n\n", err)
		os.Exit(1)
	}
	createdResources.Lock()
	createdResources.nicID = nicID
	createdResources.Unlock()
	green.Printf(" +\n")

	// Generate SSH key pair
	fmt.Printf("  o Generating SSH key pair...")
	sshKeyPair, err := azureinternal.GenerateSSHKeyPair()
	if err != nil {
		red.Printf(" x\n")
		yellow.Printf("    Error: %v\n\n", err)
		os.Exit(1)
	}
	green.Printf(" +\n")

	// Launch VM
	fmt.Printf("  o Launching Azure VM (%s)...", azureVMSize)
	vmID, err := azureinternal.LaunchVM(ctx, cfg, imageRef, nicID, identityID,
		secretKey, storageAccountName, k3sToken, azureInstallationKey, azureVMSize, sshKeyPair.PublicKey, azureEnableTerminationProtection, fullDomain,
		subnetID, nsgID)
	if err != nil {
		red.Printf(" x\n")
		yellow.Printf("    Error: %v\n\n", err)
		os.Exit(1)
	}
	createdResources.Lock()
	createdResources.vmID = vmID
	createdResources.Unlock()
	green.Printf(" +\n")
	fmt.Printf("    VM: kl-%s-vm\n", azureInstallationKey)

	fmt.Printf("  o Waiting for VM to be ready...")
	publicIP, privateIP, err := azureinternal.WaitForVM(ctx, cfg, vmID)
	if err != nil {
		red.Printf(" x\n")
		yellow.Printf("    Error: %v\n\n", err)
		os.Exit(1)
	}
	green.Printf(" +\n")
	fmt.Printf("    Public IP: %s\n", publicIP)
	fmt.Printf("    Private IP: %s\n", privateIP)

	// Application Gateway Setup (unless skipping)
	var appGwPublicIP string
	if !azureSkipAppGw {
		bold.Println("\nApplication Gateway Setup")
		bold.Println("-------------------------")

		// Create Application Gateway
		fmt.Printf("  o Creating Application Gateway...")
		appGwInfo, err := azureinternal.CreateApplicationGateway(ctx, cfg, azureInstallationKey, vnetID, []string{subnetID})
		if err != nil {
			red.Printf(" x\n")
			yellow.Printf("    Error: %v\n\n", err)
			os.Exit(1)
		}
		createdResources.Lock()
		createdResources.appGwID = appGwInfo.ID
		createdResources.Unlock()
		appGwPublicIP = appGwInfo.PublicIP
		green.Printf(" +\n")
		fmt.Printf("    App Gateway IP: %s\n", appGwPublicIP)

		// Wait for Application Gateway to be active
		fmt.Printf("  o Waiting for Application Gateway to become active...")
		err = azureinternal.WaitForAppGatewayActive(ctx, cfg, appGwInfo.ID)
		if err != nil {
			red.Printf(" x\n")
			yellow.Printf("    Error: %v\n\n", err)
			os.Exit(1)
		}
		green.Printf(" +\n")

		// Register VM with backend pool
		fmt.Printf("  o Registering VM with backend pool...")
		backendPoolID, _ := azureinternal.CreateBackendPool(ctx, cfg, azureInstallationKey)
		err = azureinternal.RegisterBackendTargets(ctx, cfg, backendPoolID, vmID)
		if err != nil {
			red.Printf(" x\n")
			yellow.Printf("    Error: %v\n\n", err)
			os.Exit(1)
		}
		green.Printf(" +\n")

		// Register App Gateway IP with console for DNS (proxied for Cloudflare TLS)
		bold.Println("\nDNS Configuration")
		bold.Println("-----------------")
		fmt.Printf("  o Configuring DNS for %s (Cloudflare proxied)...", fullDomain)
		_, err = consoleClient.ConfigureRootDNS(ctx, azureInstallationKey, secretKey, appGwPublicIP, "a", true)
		if err != nil {
			yellow.Printf(" !\n")
			yellow.Printf("    Warning: Automatic DNS configuration failed: %v\n", err)
			fmt.Println()
			bold.Println("    Manual DNS Configuration Required:")
			fmt.Printf("    Create an A record in your DNS provider:\n")
			fmt.Printf("      Name:    %s\n", verifyResult.Subdomain)
			fmt.Printf("      Type:    A\n")
			fmt.Printf("      Value:   %s\n", appGwPublicIP)
			fmt.Printf("      Proxied: Yes (if using Cloudflare)\n")
			fmt.Println()
		} else {
			green.Printf(" +\n")
		}
	}

	// Success Summary
	fmt.Println()
	green.Println("+-----------------------------------------+")
	green.Println("|   + Installation Complete!              |")
	green.Println("+-----------------------------------------+")
	fmt.Println()

	bold.Println("Instance Details")
	bold.Println("----------------")
	fmt.Printf("  VM Name:        kl-%s-vm\n", azureInstallationKey)
	fmt.Printf("  Public IP:      %s\n", publicIP)
	fmt.Printf("  Private IP:     %s\n", privateIP)
	fmt.Printf("  Location:       %s\n", cfg.Location)
	fmt.Printf("  Resource Group: %s\n", cfg.ResourceGroup)

	if !azureSkipAppGw {
		fmt.Println()
		bold.Println("Application Gateway Details")
		bold.Println("---------------------------")
		fmt.Printf("  App Gateway IP: %s\n", appGwPublicIP)
		fmt.Printf("  Custom Domain:  https://%s\n", fullDomain)
		fmt.Printf("  Wildcard:       https://*.%s\n", fullDomain)
	}

	fmt.Println()
	bold.Println("Instance Access")
	bold.Println("---------------")
	fmt.Println("  Via SSH:")
	cyan.Printf("    ssh -i ~/.ssh/kl-%s kloudlite@%s\n", azureInstallationKey, publicIP)

	fmt.Println()
	fmt.Println("  Via Azure Portal:")
	cyan.Printf("    https://portal.azure.com/#@/resource/subscriptions/%s/resourceGroups/%s/providers/Microsoft.Compute/virtualMachines/kl-%s-vm\n",
		cfg.SubscriptionID, cfg.ResourceGroup, azureInstallationKey)

	fmt.Println()
	fmt.Println("  Via Azure CLI (Serial Console):")
	cyan.Printf("    az serial-console connect -g %s -n kl-%s-vm\n", cfg.ResourceGroup, azureInstallationKey)

	// Save SSH private key to file
	sshKeyPath := fmt.Sprintf("%s/.ssh/kl-%s", os.Getenv("HOME"), azureInstallationKey)
	if err := os.WriteFile(sshKeyPath, []byte(sshKeyPair.PrivateKey), 0600); err != nil {
		yellow.Printf("\n  Warning: Could not save SSH key to %s: %v\n", sshKeyPath, err)
		fmt.Println()
		bold.Println("SSH Private Key (save this to connect to your VM):")
		bold.Println("--------------------------------------------------")
		fmt.Println(sshKeyPair.PrivateKey)
	} else {
		green.Printf("\n  SSH private key saved to: %s\n", sshKeyPath)
	}

	if !azureSkipAppGw {
		fmt.Println()
		bold.Println("Web Access")
		bold.Println("----------")
		cyan.Printf("    https://%s\n", fullDomain)
	}

	fmt.Println()
}
