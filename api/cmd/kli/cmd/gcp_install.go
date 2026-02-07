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
	gcpinternal "github.com/kloudlite/kloudlite/api/cmd/kli/internal/gcp"
	k8sinternal "github.com/kloudlite/kloudlite/api/cmd/kli/internal/k8s"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

var gcpInstallCmd = &cobra.Command{
	Use:   "install",
	Short: "Install Kloudlite on GCP",
	Long: `Install Kloudlite on GCP by creating all necessary resources.

This command will:
  - Find Ubuntu 24.04 LTS image
  - Create Service Account with required IAM roles
  - Create Cloud Storage bucket for K3s database backups
  - Create VPC firewall rules for K3s and HTTP traffic
  - Launch e2-medium VM with 100GB storage
  - Configure instance with public IP
  - Setup automated K3s SQLite backup to GCS every 30 minutes
  - Create HTTP(S) Load Balancer
  - Configure DNS with Cloudflare proxy mode (TLS termination at Cloudflare edge)

NOTE: The subdomain must be reserved in the console (console.kloudlite.io)
before running this command. The installation will fail if no subdomain
has been configured for the installation key.`,
	Example: `  # Install using gcloud defaults (project and region from ~/.config/gcloud)
  kli gcp install --installation-key prod

  # Install with specific region
  kli gcp install --installation-key prod --region us-central1

  # Install with specific project and zone
  kli gcp install --installation-key staging --project my-project --region us-central1 --zone us-central1-a

  # Install without Load Balancer (direct VM access only)
  kli gcp install --installation-key dev --skip-lb`,
	Run: runGCPInstall,
}

var (
	gcpProject                  string
	gcpRegion                   string
	gcpZone                     string
	gcpInstallationKey          string
	gcpEnableDeletionProtection bool
	gcpSkipLB                   bool
)

func init() {
	gcpInstallCmd.Flags().StringVar(&gcpProject, "project", "", "GCP project ID (reads from GOOGLE_CLOUD_PROJECT or ~/.config/gcloud)")
	gcpInstallCmd.Flags().StringVar(&gcpRegion, "region", "", "GCP region (reads from CLOUDSDK_COMPUTE_REGION or ~/.config/gcloud)")
	gcpInstallCmd.Flags().StringVar(&gcpZone, "zone", "", "GCP zone (auto-selected from region if not specified)")
	gcpInstallCmd.Flags().StringVar(&gcpInstallationKey, "installation-key", "", "Installation key to identify this installation (required)")
	gcpInstallCmd.Flags().BoolVar(&gcpEnableDeletionProtection, "enable-deletion-protection", true, "Enable VM deletion protection (default: true)")
	gcpInstallCmd.Flags().BoolVar(&gcpSkipLB, "skip-lb", false, "Skip Load Balancer setup (direct VM access only)")
	gcpInstallCmd.MarkFlagRequired("installation-key")
}

func runGCPInstall(cmd *cobra.Command, args []string) {
	green := color.New(color.FgGreen, color.Bold)
	red := color.New(color.FgRed, color.Bold)
	yellow := color.New(color.FgYellow, color.Bold)
	cyan := color.New(color.FgCyan, color.Bold)
	bold := color.New(color.Bold)

	// Header
	fmt.Println()
	cyan.Println("+-----------------------------------------+")
	cyan.Println("|   Kloudlite GCP Installation            |")
	cyan.Println("+-----------------------------------------+")
	fmt.Println()

	ctx := context.Background()

	// Setup signal handling for cleanup on interruption
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	var createdResources struct {
		sync.Mutex
		instanceName     string
		saEmail          string
		bucketName       string
		firewallsCreated bool
		lbCreated        bool
	}

	var gcpCfg *gcpinternal.GCPConfig

	go func() {
		<-sigChan
		fmt.Println()
		yellow.Println("\nInstallation interrupted! Cleaning up resources...")

		createdResources.Lock()
		defer createdResources.Unlock()

		if gcpCfg == nil {
			os.Exit(130)
		}

		// Cleanup in reverse order
		if createdResources.lbCreated {
			fmt.Printf("  Deleting Load Balancer components...\n")
			gcpinternal.DeleteLoadBalancer(context.Background(), gcpCfg, gcpInstallationKey)
		}
		if createdResources.instanceName != "" {
			fmt.Printf("  Terminating instance %s...\n", createdResources.instanceName)
			gcpinternal.DeleteInstance(context.Background(), gcpCfg, createdResources.instanceName)
		}
		if createdResources.firewallsCreated {
			fmt.Printf("  Deleting firewall rules...\n")
			gcpinternal.DeleteFirewallRules(context.Background(), gcpCfg, gcpInstallationKey)
		}
		if createdResources.saEmail != "" {
			fmt.Printf("  Deleting service account...\n")
			gcpinternal.DeleteServiceAccount(context.Background(), gcpCfg, gcpInstallationKey)
		}
		if createdResources.bucketName != "" {
			fmt.Printf("  Deleting storage bucket...\n")
			gcpinternal.DeleteStorageBucket(context.Background(), gcpCfg, createdResources.bucketName)
		}

		yellow.Println("Cleanup completed. Exiting...")
		os.Exit(130)
	}()

	// Configuration
	bold.Println("Configuration")
	bold.Println("-------------")
	fmt.Printf("  Installation Key: %s\n", gcpInstallationKey)
	if gcpProject != "" {
		fmt.Printf("  Project:          %s\n", gcpProject)
	}
	fmt.Printf("  Region:           %s\n", gcpRegion)
	if gcpZone != "" {
		fmt.Printf("  Zone:             %s\n", gcpZone)
	}

	cfg, err := gcpinternal.LoadGCPConfig(ctx, gcpProject, gcpRegion, gcpZone)
	if err != nil {
		red.Printf("x\n")
		yellow.Printf("  Error: %v\n\n", err)
		os.Exit(1)
	}
	gcpCfg = cfg
	green.Printf("  Project:          %s\n", cfg.Project)
	green.Printf("  Zone:             %s (auto-selected)\n", cfg.Zone)
	fmt.Println()

	// Console API client
	consoleClient := console.NewClient()

	// Verify Installation and get subdomain
	bold.Println("Verifying Installation")
	bold.Println("----------------------")

	fmt.Printf("  o Verifying installation key with registration API...")
	verifyResult, err := k8sinternal.VerifyInstallation(ctx, gcpInstallationKey, &k8sinternal.VerifyInstallationOptions{
		Provider: "gcp",
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
	if !gcpSkipLB {
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

	// Enable required GCP APIs
	fmt.Printf("  o Enabling required GCP APIs...")
	if err := gcpinternal.EnableRequiredAPIs(ctx, cfg.Project); err != nil {
		red.Printf(" x\n")
		yellow.Printf("    Error: %v\n\n", err)
		os.Exit(1)
	}
	green.Printf(" +\n")

	// Find Ubuntu image
	fmt.Printf("  o Finding Ubuntu 24.04 LTS image...")
	imageURL, err := gcpinternal.GetUbuntu2404Image(ctx)
	if err != nil {
		red.Printf(" x\n")
		yellow.Printf("    Error: %v\n\n", err)
		os.Exit(1)
	}
	green.Printf(" +\n")
	fmt.Printf("    %s\n", gcpinternal.GetImageName(imageURL))

	time.Sleep(1 * time.Second)

	// Network Resources
	fmt.Printf("  o Setting up network...")
	networkName, _, err := gcpinternal.GetDefaultVPC(ctx, cfg)
	if err != nil {
		red.Printf(" x\n")
		yellow.Printf("    Error: %v\n\n", err)
		os.Exit(1)
	}

	_, subnetCIDR, err := gcpinternal.GetDefaultSubnet(ctx, cfg, networkName)
	if err != nil {
		red.Printf(" x\n")
		yellow.Printf("    Error: %v\n\n", err)
		os.Exit(1)
	}
	green.Printf(" +\n")
	fmt.Printf("    Network: %s\n", networkName)
	fmt.Printf("    Subnet CIDR: %s\n", subnetCIDR)

	time.Sleep(1 * time.Second)

	// Parallel Resource Creation
	fmt.Printf("  o Creating resources in parallel...\n")

	var wg sync.WaitGroup
	var saEmail, bucketName string
	var saErr, storageErr, firewallErr error
	bucketName = gcpinternal.GetBucketName(gcpInstallationKey)

	startTime := time.Now()

	// Service Account (parallel)
	wg.Add(1)
	go func() {
		defer wg.Done()
		fmt.Printf("    [%s] Starting: Service Account creation\n", time.Now().Format("15:04:05"))
		saEmail, saErr = gcpinternal.EnsureServiceAccount(ctx, cfg, gcpInstallationKey)
		if saErr != nil {
			fmt.Printf("    [%s] Failed: Service Account - %v\n", time.Now().Format("15:04:05"), saErr)
		} else {
			createdResources.Lock()
			createdResources.saEmail = saEmail
			createdResources.Unlock()
			fmt.Printf("    [%s] Completed: Service Account\n", time.Now().Format("15:04:05"))
		}
	}()

	// Storage Bucket (parallel)
	wg.Add(1)
	go func() {
		defer wg.Done()
		fmt.Printf("    [%s] Starting: Storage Bucket creation\n", time.Now().Format("15:04:05"))
		storageErr = gcpinternal.EnsureStorageBucket(ctx, cfg, bucketName, gcpInstallationKey)
		if storageErr != nil {
			fmt.Printf("    [%s] Failed: Storage Bucket - %v\n", time.Now().Format("15:04:05"), storageErr)
		} else {
			createdResources.Lock()
			createdResources.bucketName = bucketName
			createdResources.Unlock()
			fmt.Printf("    [%s] Completed: Storage Bucket\n", time.Now().Format("15:04:05"))
		}
	}()

	// Firewall Rules (parallel)
	wg.Add(1)
	go func() {
		defer wg.Done()
		fmt.Printf("    [%s] Starting: Firewall Rules creation\n", time.Now().Format("15:04:05"))
		firewallErr = gcpinternal.EnsureFirewallRules(ctx, cfg, subnetCIDR, gcpInstallationKey)
		if firewallErr != nil {
			fmt.Printf("    [%s] Failed: Firewall Rules - %v\n", time.Now().Format("15:04:05"), firewallErr)
		} else {
			createdResources.Lock()
			createdResources.firewallsCreated = true
			createdResources.Unlock()
			fmt.Printf("    [%s] Completed: Firewall Rules\n", time.Now().Format("15:04:05"))
		}
	}()

	wg.Wait()
	elapsed := time.Since(startTime)
	fmt.Printf("    Parallel operations completed in %.1fs\n", elapsed.Seconds())

	// Check for errors
	if saErr != nil {
		red.Printf(" x\n")
		yellow.Printf("    Service Account Error: %v\n\n", saErr)
		os.Exit(1)
	}
	if storageErr != nil {
		red.Printf(" x\n")
		yellow.Printf("    Storage Bucket Error: %v\n\n", storageErr)
		os.Exit(1)
	}
	if firewallErr != nil {
		red.Printf(" x\n")
		yellow.Printf("    Firewall Rules Error: %v\n\n", firewallErr)
		os.Exit(1)
	}

	green.Printf(" +\n")
	fmt.Printf("    Service Account: %s\n", saEmail)
	fmt.Printf("    Storage Bucket:  %s\n", bucketName)
	fmt.Printf("    Firewall Rules:  Created\n")

	// Grant IAM roles (depends on service account)
	bold.Println("\nFinalizing IAM Setup")
	bold.Println("--------------------")
	fmt.Printf("  o Granting IAM roles...")
	err = gcpinternal.GrantIAMRoles(ctx, cfg, saEmail, gcpInstallationKey)
	if err != nil {
		red.Printf(" x\n")
		yellow.Printf("    Error: %v\n\n", err)
		os.Exit(1)
	}
	green.Printf(" +\n")

	time.Sleep(2 * time.Second)

	// Instance Launch
	bold.Println("\nInstance Deployment")
	bold.Println("-------------------")

	// Generate K3s agent token
	k3sToken, err := gcpinternal.GenerateK3sToken()
	if err != nil {
		red.Printf(" x\n")
		yellow.Printf("    Error generating K3s token: %v\n\n", err)
		os.Exit(1)
	}

	fmt.Printf("  o Launching Compute Engine VM (e2-medium)...")
	instanceName, err := gcpinternal.LaunchInstance(ctx, cfg, imageURL, saEmail, secretKey, bucketName, k3sToken, gcpInstallationKey, fullDomain, gcpEnableDeletionProtection)
	if err != nil {
		red.Printf(" x\n")
		yellow.Printf("    Error: %v\n\n", err)
		os.Exit(1)
	}
	createdResources.Lock()
	createdResources.instanceName = instanceName
	createdResources.Unlock()
	green.Printf(" +\n")
	fmt.Printf("    %s\n", instanceName)

	fmt.Printf("  o Waiting for instance to be ready...")
	publicIP, privateIP, err := gcpinternal.WaitForInstance(ctx, cfg, instanceName)
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
	if !gcpSkipLB {
		bold.Println("\nLoad Balancer Setup")
		bold.Println("-------------------")

		// Create IP, health check, and instance group in parallel (no dependencies between them)
		fmt.Printf("  o Creating LB prerequisites in parallel...")
		var lbIPResult string
		var hcURL, igURL string
		var ipErr, hcErr, igErr error
		var lbWg sync.WaitGroup
		lbWg.Add(3)
		go func() {
			defer lbWg.Done()
			lbIPResult, ipErr = gcpinternal.ReserveExternalIP(ctx, cfg, gcpInstallationKey)
		}()
		go func() {
			defer lbWg.Done()
			hcURL, hcErr = gcpinternal.CreateHealthCheck(ctx, cfg, gcpInstallationKey)
		}()
		go func() {
			defer lbWg.Done()
			igURL, igErr = gcpinternal.CreateUnmanagedInstanceGroup(ctx, cfg, instanceName, gcpInstallationKey)
		}()
		lbWg.Wait()

		if ipErr != nil {
			red.Printf(" x\n")
			yellow.Printf("    Error reserving IP: %v\n\n", ipErr)
			os.Exit(1)
		}
		if hcErr != nil {
			red.Printf(" x\n")
			yellow.Printf("    Error creating health check: %v\n\n", hcErr)
			os.Exit(1)
		}
		if igErr != nil {
			red.Printf(" x\n")
			yellow.Printf("    Error creating instance group: %v\n\n", igErr)
			os.Exit(1)
		}
		lbIP = lbIPResult
		green.Printf(" +\n")
		fmt.Printf("    IP: %s\n", lbIP)

		// Create Backend Service (depends on health check + instance group)
		fmt.Printf("  o Creating backend service...")
		bsURL, err := gcpinternal.CreateBackendService(ctx, cfg, igURL, hcURL, gcpInstallationKey)
		if err != nil {
			red.Printf(" x\n")
			yellow.Printf("    Error: %v\n\n", err)
			os.Exit(1)
		}
		green.Printf(" +\n")

		// Create URL Map
		fmt.Printf("  o Creating URL map...")
		urlMapURL, err := gcpinternal.CreateURLMap(ctx, cfg, bsURL, gcpInstallationKey)
		if err != nil {
			red.Printf(" x\n")
			yellow.Printf("    Error: %v\n\n", err)
			os.Exit(1)
		}
		green.Printf(" +\n")

		// Create Target HTTP Proxy
		fmt.Printf("  o Creating target HTTP proxy...")
		proxyURL, err := gcpinternal.CreateTargetHTTPProxy(ctx, cfg, urlMapURL, gcpInstallationKey)
		if err != nil {
			red.Printf(" x\n")
			yellow.Printf("    Error: %v\n\n", err)
			os.Exit(1)
		}
		green.Printf(" +\n")

		// Create Global Forwarding Rule
		fmt.Printf("  o Creating forwarding rule...")
		err = gcpinternal.CreateGlobalForwardingRule(ctx, cfg, lbIP, proxyURL, gcpInstallationKey)
		if err != nil {
			red.Printf(" x\n")
			yellow.Printf("    Error: %v\n\n", err)
			os.Exit(1)
		}
		createdResources.Lock()
		createdResources.lbCreated = true
		createdResources.Unlock()
		green.Printf(" +\n")

		// Wait for LB to become active
		fmt.Printf("  o Waiting for load balancer to become active...")
		err = gcpinternal.WaitForLoadBalancerActive(ctx, cfg, gcpInstallationKey)
		if err != nil {
			red.Printf(" x\n")
			yellow.Printf("    Error: %v\n\n", err)
			os.Exit(1)
		}
		green.Printf(" +\n")

		// Register LB IP with console for DNS configuration (A record, Cloudflare proxied for TLS)
		bold.Println("\nDNS Configuration")
		bold.Println("-----------------")
		fmt.Printf("  o Configuring DNS for %s (Cloudflare proxied)...", fullDomain)
		_, err = consoleClient.ConfigureRootDNS(ctx, gcpInstallationKey, secretKey, lbIP, "a", true)
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
	fmt.Printf("  Instance Name:  %s\n", instanceName)
	fmt.Printf("  Public IP:      %s\n", publicIP)
	fmt.Printf("  Private IP:     %s\n", privateIP)
	fmt.Printf("  Project:        %s\n", cfg.Project)
	fmt.Printf("  Zone:           %s\n", cfg.Zone)

	if !gcpSkipLB {
		fmt.Println()
		bold.Println("Load Balancer Details")
		bold.Println("---------------------")
		fmt.Printf("  LB IP:          %s\n", lbIP)
		fmt.Printf("  Custom Domain:  https://%s\n", fullDomain)
		fmt.Printf("  Wildcard:       https://*.%s\n", fullDomain)
	}

	fmt.Println()
	bold.Println("Instance Access")
	bold.Println("---------------")
	fmt.Println("  Via gcloud:")
	cyan.Printf("    gcloud compute ssh %s --zone %s --project %s\n", instanceName, cfg.Zone, cfg.Project)

	if !gcpSkipLB {
		fmt.Println()
		bold.Println("Web Access")
		bold.Println("----------")
		cyan.Printf("    https://%s\n", fullDomain)
	}

	fmt.Println()
}
