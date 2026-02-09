package cmd

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/kloudlite/kloudlite/api/cmd/kli/internal/console"
	k8sinternal "github.com/kloudlite/kloudlite/api/cmd/kli/internal/k8s"
	ociinternal "github.com/kloudlite/kloudlite/api/cmd/kli/internal/oci"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

var ociInstallCmd = &cobra.Command{
	Use:   "install",
	Short: "Install Kloudlite on OCI",
	Long: `Install Kloudlite on OCI by creating all necessary resources.

This command will:
  - Find Ubuntu 24.04 image
  - Create VCN, Subnet, and Network Security Group
  - Create Object Storage bucket for K3s database backups
  - Create Dynamic Group and IAM Policy
  - Launch VM.Standard.E4.Flex instance (1 OCPU, 8GB RAM) with 100GB boot volume
  - Create Network Load Balancer (Layer 4, TCP passthrough)
  - Configure DNS with Cloudflare proxy mode (TLS termination at Cloudflare edge)

NOTE: The subdomain must be reserved in the console (console.kloudlite.io)
before running this command. The installation will fail if no subdomain
has been configured for the installation key.`,
	Example: `  # Install using OCI defaults (from ~/.oci/config)
  kli oci install --installation-key prod

  # Install with specific compartment and region
  kli oci install --installation-key prod --compartment ocid1.compartment.oc1..xxx --region us-ashburn-1

  # Install without Load Balancer (direct VM access only)
  kli oci install --installation-key dev --skip-lb`,
	Run: runOCIInstall,
}

var (
	ociTenancy                  string
	ociUser                     string
	ociRegion                   string
	ociCompartment              string
	ociFingerprint              string
	ociKeyFile                  string
	ociInstallationKey          string
	ociEnableDeletionProtection bool
	ociSkipLB                   bool
	ociDevMode                  bool
)

func init() {
	ociInstallCmd.Flags().StringVar(&ociTenancy, "tenancy", "", "OCI tenancy OCID (reads from OCI_CLI_TENANCY or ~/.oci/config)")
	ociInstallCmd.Flags().StringVar(&ociUser, "user", "", "OCI user OCID (reads from OCI_CLI_USER or ~/.oci/config)")
	ociInstallCmd.Flags().StringVar(&ociRegion, "region", "", "OCI region (reads from OCI_CLI_REGION or ~/.oci/config)")
	ociInstallCmd.Flags().StringVar(&ociCompartment, "compartment", "", "OCI compartment OCID (defaults to tenancy)")
	ociInstallCmd.Flags().StringVar(&ociFingerprint, "fingerprint", "", "OCI API key fingerprint")
	ociInstallCmd.Flags().StringVar(&ociKeyFile, "key-file", "", "OCI API private key file path")
	ociInstallCmd.Flags().StringVar(&ociInstallationKey, "installation-key", "", "Installation key to identify this installation (required)")
	ociInstallCmd.Flags().BoolVar(&ociEnableDeletionProtection, "enable-deletion-protection", true, "Enable VM deletion protection (default: true)")
	ociInstallCmd.Flags().BoolVar(&ociSkipLB, "skip-lb", false, "Skip Load Balancer setup (direct VM access only)")
	ociInstallCmd.Flags().BoolVar(&ociDevMode, "dev", false, "Development mode: inject local SSH public key for instance access")
	ociInstallCmd.MarkFlagRequired("installation-key")
}

func runOCIInstall(cmd *cobra.Command, args []string) {
	green := color.New(color.FgGreen, color.Bold)
	red := color.New(color.FgRed, color.Bold)
	yellow := color.New(color.FgYellow, color.Bold)
	cyan := color.New(color.FgCyan, color.Bold)
	bold := color.New(color.Bold)

	// Header
	fmt.Println()
	cyan.Println("+-----------------------------------------+")
	cyan.Println("|   Kloudlite OCI Installation            |")
	cyan.Println("+-----------------------------------------+")
	fmt.Println()

	ctx := context.Background()

	// Setup signal handling for cleanup on interruption
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	var createdResources struct {
		sync.Mutex
		instanceID   string
		nsgID        string
		vcnID        string
		bucketName   string
		dgCreated    bool
		policyCreated bool
		nlbCreated   bool
	}

	var ociCfg *ociinternal.OCIConfig

	go func() {
		<-sigChan
		fmt.Println()
		yellow.Println("\nInstallation interrupted! Cleaning up resources...")

		createdResources.Lock()
		defer createdResources.Unlock()

		if ociCfg == nil {
			os.Exit(130)
		}

		// Cleanup in reverse order
		if createdResources.nlbCreated {
			fmt.Printf("  Deleting Network Load Balancer...\n")
			ociinternal.DeleteNetworkLoadBalancer(context.Background(), ociCfg, ociInstallationKey)
		}
		if createdResources.instanceID != "" {
			fmt.Printf("  Terminating instance...\n")
			ociinternal.DeleteInstance(context.Background(), ociCfg, createdResources.instanceID)
		}
		if createdResources.nsgID != "" && createdResources.vcnID != "" {
			fmt.Printf("  Deleting NSG...\n")
			ociinternal.DeleteNSG(context.Background(), ociCfg, createdResources.vcnID, ociInstallationKey)
		}
		if createdResources.dgCreated {
			fmt.Printf("  Deleting Dynamic Group...\n")
			ociinternal.DeleteDynamicGroup(context.Background(), ociCfg, ociInstallationKey)
		}
		if createdResources.policyCreated {
			fmt.Printf("  Deleting Policy...\n")
			ociinternal.DeletePolicy(context.Background(), ociCfg, ociInstallationKey)
		}
		if createdResources.bucketName != "" {
			fmt.Printf("  Deleting storage bucket...\n")
			ociinternal.DeleteStorageBucket(context.Background(), ociCfg, createdResources.bucketName)
		}

		yellow.Println("Cleanup completed. Exiting...")
		os.Exit(130)
	}()

	// Configuration
	bold.Println("Configuration")
	bold.Println("-------------")
	fmt.Printf("  Installation Key: %s\n", ociInstallationKey)

	cfg, err := ociinternal.LoadOCIConfig(ctx, ociTenancy, ociUser, ociRegion, ociCompartment, ociFingerprint, ociKeyFile)
	if err != nil {
		red.Printf("x\n")
		yellow.Printf("  Error: %v\n\n", err)
		os.Exit(1)
	}
	ociCfg = cfg
	green.Printf("  Tenancy:     %s\n", cfg.TenancyOCID)
	green.Printf("  Region:      %s\n", cfg.Region)
	green.Printf("  Compartment: %s\n", cfg.CompartmentOCID)
	fmt.Println()

	// Console API client
	consoleClient := console.NewClient()

	// Verify Installation and get subdomain
	bold.Println("Verifying Installation")
	bold.Println("----------------------")

	fmt.Printf("  o Verifying installation key with registration API...")
	verifyResult, err := k8sinternal.VerifyInstallation(ctx, ociInstallationKey, &k8sinternal.VerifyInstallationOptions{
		Provider: "oci",
		Region:   cfg.Region,
	})
	if err != nil {
		red.Printf(" x\n")
		yellow.Printf("    Error: %v\n\n", err)
		os.Exit(1)
	}
	green.Printf(" +\n")
	fmt.Printf("    Secret key obtained successfully\n")

	secretKey := verifyResult.SecretKey
	var fullDomain string

	// Check if subdomain was configured in console (required for LB)
	if !ociSkipLB {
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

	// Find Ubuntu image
	fmt.Printf("  o Finding Ubuntu 24.04 image...")
	imageID, imageName, err := ociinternal.FindUbuntuImage(ctx, cfg)
	if err != nil {
		red.Printf(" x\n")
		yellow.Printf("    Error: %v\n\n", err)
		os.Exit(1)
	}
	green.Printf(" +\n")
	fmt.Printf("    %s\n", imageName)

	time.Sleep(1 * time.Second)

	// Network Resources
	fmt.Printf("  o Setting up network...")
	vcnID, vcnCIDR, err := ociinternal.GetDefaultVCN(ctx, cfg)
	if err != nil {
		red.Printf(" x\n")
		yellow.Printf("    Error: %v\n\n", err)
		os.Exit(1)
	}

	subnetID, subnetCIDR, err := ociinternal.GetDefaultSubnet(ctx, cfg, vcnID)
	if err != nil {
		red.Printf(" x\n")
		yellow.Printf("    Error: %v\n\n", err)
		os.Exit(1)
	}
	green.Printf(" +\n")
	fmt.Printf("    VCN: %s\n", vcnID)
	fmt.Printf("    Subnet: %s (CIDR: %s)\n", subnetID, subnetCIDR)

	createdResources.Lock()
	createdResources.vcnID = vcnID
	createdResources.Unlock()

	time.Sleep(1 * time.Second)

	// Parallel Resource Creation
	fmt.Printf("  o Creating resources in parallel...\n")

	var wg sync.WaitGroup
	var nsgID, bucketName string
	var nsgErr, storageErr, dgErr, policyErr error
	bucketName = ociinternal.GetBucketName(ociInstallationKey)

	startTime := time.Now()

	// NSG (parallel)
	wg.Add(1)
	go func() {
		defer wg.Done()
		fmt.Printf("    [%s] Starting: NSG creation\n", time.Now().Format("15:04:05"))
		nsgID, nsgErr = ociinternal.EnsureNSG(ctx, cfg, vcnID, vcnCIDR, ociInstallationKey)
		if nsgErr != nil {
			fmt.Printf("    [%s] Failed: NSG - %v\n", time.Now().Format("15:04:05"), nsgErr)
		} else {
			createdResources.Lock()
			createdResources.nsgID = nsgID
			createdResources.Unlock()
			fmt.Printf("    [%s] Completed: NSG\n", time.Now().Format("15:04:05"))
		}
	}()

	// Storage Bucket (parallel)
	wg.Add(1)
	go func() {
		defer wg.Done()
		fmt.Printf("    [%s] Starting: Storage Bucket creation\n", time.Now().Format("15:04:05"))
		storageErr = ociinternal.EnsureStorageBucket(ctx, cfg, bucketName, ociInstallationKey)
		if storageErr != nil {
			fmt.Printf("    [%s] Failed: Storage Bucket - %v\n", time.Now().Format("15:04:05"), storageErr)
		} else {
			createdResources.Lock()
			createdResources.bucketName = bucketName
			createdResources.Unlock()
			fmt.Printf("    [%s] Completed: Storage Bucket\n", time.Now().Format("15:04:05"))
		}
	}()

	// Dynamic Group + Policy (parallel)
	wg.Add(1)
	go func() {
		defer wg.Done()
		fmt.Printf("    [%s] Starting: Dynamic Group + Policy creation\n", time.Now().Format("15:04:05"))
		_, dgErr = ociinternal.EnsureDynamicGroup(ctx, cfg, ociInstallationKey)
		if dgErr != nil {
			fmt.Printf("    [%s] Failed: Dynamic Group - %v\n", time.Now().Format("15:04:05"), dgErr)
			return
		}
		createdResources.Lock()
		createdResources.dgCreated = true
		createdResources.Unlock()

		_, policyErr = ociinternal.EnsurePolicy(ctx, cfg, ociInstallationKey)
		if policyErr != nil {
			fmt.Printf("    [%s] Failed: Policy - %v\n", time.Now().Format("15:04:05"), policyErr)
		} else {
			createdResources.Lock()
			createdResources.policyCreated = true
			createdResources.Unlock()
			fmt.Printf("    [%s] Completed: Dynamic Group + Policy\n", time.Now().Format("15:04:05"))
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
	if storageErr != nil {
		red.Printf(" x\n")
		yellow.Printf("    Storage Bucket Error: %v\n\n", storageErr)
		os.Exit(1)
	}
	if dgErr != nil {
		red.Printf(" x\n")
		yellow.Printf("    Dynamic Group Error: %v\n\n", dgErr)
		os.Exit(1)
	}
	if policyErr != nil {
		red.Printf(" x\n")
		yellow.Printf("    Policy Error: %v\n\n", policyErr)
		os.Exit(1)
	}

	green.Printf(" +\n")
	fmt.Printf("    NSG:            %s\n", nsgID)
	fmt.Printf("    Storage Bucket: %s\n", bucketName)
	fmt.Printf("    Dynamic Group:  Created\n")
	fmt.Printf("    Policy:         Created\n")

	// Instance Launch
	bold.Println("\nInstance Deployment")
	bold.Println("-------------------")

	// Generate K3s agent token
	k3sToken, err := ociinternal.GenerateK3sToken()
	if err != nil {
		red.Printf(" x\n")
		yellow.Printf("    Error generating K3s token: %v\n\n", err)
		os.Exit(1)
	}

	// Read local SSH public key in dev mode
	sshPubKey := ""
	if ociDevMode {
		sshPubKey = readLocalSSHPublicKey()
		if sshPubKey != "" {
			fmt.Printf("  o Dev mode: SSH public key will be injected\n")
		} else {
			yellow.Printf("  o Dev mode: No SSH public key found in ~/.ssh/\n")
		}
	}

	fmt.Printf("  o Launching OCI instance (VM.Standard.E4.Flex 1 OCPU / 8GB)...")
	instanceID, err := ociinternal.LaunchInstance(ctx, cfg, imageID, subnetID, nsgID, secretKey, bucketName, k3sToken, ociInstallationKey, fullDomain, sshPubKey, ociEnableDeletionProtection)
	if err != nil {
		red.Printf(" x\n")
		yellow.Printf("    Error: %v\n\n", err)
		os.Exit(1)
	}
	createdResources.Lock()
	createdResources.instanceID = instanceID
	createdResources.Unlock()
	green.Printf(" +\n")
	fmt.Printf("    %s\n", instanceID)

	fmt.Printf("  o Waiting for instance to be ready...")
	publicIP, privateIP, err := ociinternal.WaitForInstance(ctx, cfg, instanceID)
	if err != nil {
		red.Printf(" x\n")
		yellow.Printf("    Error: %v\n\n", err)
		os.Exit(1)
	}
	green.Printf(" +\n")
	fmt.Printf("    Public IP: %s\n", publicIP)
	fmt.Printf("    Private IP: %s\n", privateIP)

	// Load Balancer Setup (unless skipping)
	var lbIP string
	if !ociSkipLB {
		bold.Println("\nLoad Balancer Setup")
		bold.Println("-------------------")

		fmt.Printf("  o Creating Network Load Balancer...")
		_, nlbIP, err := ociinternal.CreateNetworkLoadBalancer(ctx, cfg, subnetID, instanceID, privateIP, ociInstallationKey)
		if err != nil {
			red.Printf(" x\n")
			yellow.Printf("    Error: %v\n\n", err)
			os.Exit(1)
		}
		lbIP = nlbIP
		createdResources.Lock()
		createdResources.nlbCreated = true
		createdResources.Unlock()
		green.Printf(" +\n")
		fmt.Printf("    NLB IP: %s\n", lbIP)

		// Register LB IP with console for DNS configuration (A record, Cloudflare proxied for TLS)
		bold.Println("\nDNS Configuration")
		bold.Println("-----------------")
		fmt.Printf("  o Configuring DNS for %s (Cloudflare proxied)...", fullDomain)
		_, err = consoleClient.ConfigureRootDNS(ctx, ociInstallationKey, secretKey, lbIP, "a", true)
		if err != nil {
			red.Printf(" x\n")
			yellow.Printf("    Error: %v\n\n", err)
			os.Exit(1)
		}
		green.Printf(" +\n")
	}

	// Success Summary
	fmt.Println()
	green.Println("+-----------------------------------------+")
	green.Println("|   + Installation Complete!              |")
	green.Println("+-----------------------------------------+")
	fmt.Println()

	bold.Println("Instance Details")
	bold.Println("----------------")
	fmt.Printf("  Instance ID:    %s\n", instanceID)
	fmt.Printf("  Public IP:      %s\n", publicIP)
	fmt.Printf("  Private IP:     %s\n", privateIP)
	fmt.Printf("  Region:         %s\n", cfg.Region)
	fmt.Printf("  Compartment:    %s\n", cfg.CompartmentOCID)

	if !ociSkipLB {
		fmt.Println()
		bold.Println("Load Balancer Details")
		bold.Println("---------------------")
		fmt.Printf("  NLB IP:         %s\n", lbIP)
		fmt.Printf("  Custom Domain:  https://%s\n", fullDomain)
		fmt.Printf("  Wildcard:       https://*.%s\n", fullDomain)
	}

	fmt.Println()
	bold.Println("Instance Access")
	bold.Println("---------------")
	fmt.Println("  Via OCI CLI:")
	cyan.Printf("    oci compute instance get --instance-id %s\n", instanceID)
	fmt.Println("  Via SSH:")
	cyan.Printf("    ssh ubuntu@%s\n", publicIP)

	if !ociSkipLB {
		fmt.Println()
		bold.Println("Web Access")
		bold.Println("----------")
		cyan.Printf("    https://%s\n", fullDomain)
	}

	fmt.Println()
}

func readLocalSSHPublicKey() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	for _, name := range []string{"id_ed25519.pub", "id_rsa.pub", "id_ecdsa.pub"} {
		path := home + "/.ssh/" + name
		data, err := os.ReadFile(path)
		if err == nil && len(data) > 0 {
			return string(data)
		}
	}
	return ""
}
