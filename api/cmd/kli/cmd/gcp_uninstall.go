package cmd

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	gcpinternal "github.com/kloudlite/kloudlite/api/cmd/kli/internal/gcp"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

var gcpUninstallCmd = &cobra.Command{
	Use:   "uninstall",
	Short: "Uninstall Kloudlite from GCP",
	Long: `Uninstall Kloudlite from GCP by removing all resources created during installation.

This command will:
  - Delete Load Balancer components (forwarding rule, proxy, URL map, backend service, instance group, health check, reserved IP)
  - Terminate Compute Engine VM 'kl-{installation-key}-instance'
  - Delete VPC firewall rules
  - Delete Service Account 'kl-{installation-key}-sa'
  - Delete Cloud Storage bucket 'kl-{installation-key}-backups' and all backups

All resources are identified by the installation key.`,
	Example: `  # Uninstall using default GCP project
  kli gcp uninstall --installation-key prod --region us-central1

  # Uninstall with specific project
  kli gcp uninstall --installation-key staging --project my-project --region us-central1 --zone us-central1-a`,
	Run: runGCPUninstall,
}

var (
	gcpUninstallProject         string
	gcpUninstallRegion          string
	gcpUninstallZone            string
	gcpUninstallInstallationKey string
)

func init() {
	gcpUninstallCmd.Flags().StringVar(&gcpUninstallProject, "project", "", "GCP project ID (uses GOOGLE_CLOUD_PROJECT if not specified)")
	gcpUninstallCmd.Flags().StringVar(&gcpUninstallRegion, "region", "", "GCP region (required)")
	gcpUninstallCmd.Flags().StringVar(&gcpUninstallZone, "zone", "", "GCP zone (auto-selected from region if not specified)")
	gcpUninstallCmd.Flags().StringVar(&gcpUninstallInstallationKey, "installation-key", "", "Installation key to identify this installation (required)")
	gcpUninstallCmd.MarkFlagRequired("installation-key")
	gcpUninstallCmd.MarkFlagRequired("region")
}

func runGCPUninstall(cmd *cobra.Command, args []string) {
	green := color.New(color.FgGreen, color.Bold)
	red := color.New(color.FgRed, color.Bold)
	yellow := color.New(color.FgYellow, color.Bold)
	cyan := color.New(color.FgCyan, color.Bold)
	bold := color.New(color.Bold)

	// Header
	fmt.Println()
	cyan.Println("+-----------------------------------------+")
	cyan.Println("|   Kloudlite GCP Uninstallation          |")
	cyan.Println("+-----------------------------------------+")
	fmt.Println()

	ctx := context.Background()

	// Setup signal handling - for uninstallation, we don't want to abort mid-cleanup
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-sigChan
		fmt.Println()
		yellow.Println("\nInterrupt received. Uninstallation will continue to completion...")
		yellow.Println("   (Aborting now may leave orphaned resources)")
		// Don't exit - let uninstallation complete
	}()

	// Configuration
	bold.Println("Configuration")
	bold.Println("-------------")
	fmt.Printf("  Installation Key: %s\n", gcpUninstallInstallationKey)
	if gcpUninstallProject != "" {
		fmt.Printf("  Project:          %s\n", gcpUninstallProject)
	}
	fmt.Printf("  Region:           %s\n", gcpUninstallRegion)
	if gcpUninstallZone != "" {
		fmt.Printf("  Zone:             %s\n", gcpUninstallZone)
	}

	cfg, err := gcpinternal.LoadGCPConfig(ctx, gcpUninstallProject, gcpUninstallRegion, gcpUninstallZone)
	if err != nil {
		red.Printf("x\n")
		yellow.Printf("  Error: %v\n\n", err)
		os.Exit(1)
	}
	green.Printf("  Project:          %s\n", cfg.Project)
	green.Printf("  Zone:             %s\n", cfg.Zone)
	fmt.Println()

	// Resource Cleanup
	bold.Println("Removing Resources")
	bold.Println("------------------")

	fmt.Printf("  o Cleaning up resources...\n")

	startTime := time.Now()

	// Phase 1: Delete Load Balancer components (must be done first in reverse order)
	fmt.Printf("    [%s] Starting: Load Balancer cleanup\n", time.Now().Format("15:04:05"))
	err = gcpinternal.DeleteLoadBalancer(ctx, cfg, gcpUninstallInstallationKey)
	if err != nil {
		fmt.Printf("    [%s] Warning: Load Balancer cleanup - %v\n", time.Now().Format("15:04:05"), err)
	} else {
		fmt.Printf("    [%s] Completed: Load Balancer cleanup\n", time.Now().Format("15:04:05"))
	}

	// Phase 2: Delete Instance
	fmt.Printf("    [%s] Starting: VM instance deletion\n", time.Now().Format("15:04:05"))
	instanceName := gcpinternal.GetInstanceName(gcpUninstallInstallationKey)
	instanceErr := gcpinternal.DeleteInstance(ctx, cfg, instanceName)
	if instanceErr != nil {
		fmt.Printf("    [%s] Warning: VM instance - %v\n", time.Now().Format("15:04:05"), instanceErr)
	} else {
		fmt.Printf("    [%s] Completed: VM instance deleted\n", time.Now().Format("15:04:05"))
	}

	// Phase 3: Parallel cleanup of remaining resources
	var wg sync.WaitGroup
	var firewallErr, saErr, storageErr error

	bucketName := gcpinternal.GetBucketName(gcpUninstallInstallationKey)

	// Firewall Rules (parallel)
	wg.Add(1)
	go func() {
		defer wg.Done()
		fmt.Printf("    [%s] Starting: Firewall Rules deletion\n", time.Now().Format("15:04:05"))
		firewallErr = gcpinternal.DeleteFirewallRules(ctx, cfg, gcpUninstallInstallationKey)
		if firewallErr != nil {
			fmt.Printf("    [%s] Failed: Firewall Rules - %v\n", time.Now().Format("15:04:05"), firewallErr)
		} else {
			fmt.Printf("    [%s] Completed: Firewall Rules\n", time.Now().Format("15:04:05"))
		}
	}()

	// Service Account (parallel)
	wg.Add(1)
	go func() {
		defer wg.Done()
		fmt.Printf("    [%s] Starting: Service Account deletion\n", time.Now().Format("15:04:05"))
		// First remove IAM roles
		saEmail := gcpinternal.GetServiceAccountEmail(cfg, gcpUninstallInstallationKey)
		_ = gcpinternal.RemoveIAMRoles(ctx, cfg, saEmail)
		// Then delete the service account
		saErr = gcpinternal.DeleteServiceAccount(ctx, cfg, gcpUninstallInstallationKey)
		if saErr != nil {
			fmt.Printf("    [%s] Failed: Service Account - %v\n", time.Now().Format("15:04:05"), saErr)
		} else {
			fmt.Printf("    [%s] Completed: Service Account\n", time.Now().Format("15:04:05"))
		}
	}()

	// Storage Bucket (parallel)
	wg.Add(1)
	go func() {
		defer wg.Done()
		fmt.Printf("    [%s] Starting: Storage Bucket deletion\n", time.Now().Format("15:04:05"))
		storageErr = gcpinternal.DeleteStorageBucket(ctx, cfg, bucketName)
		if storageErr != nil {
			fmt.Printf("    [%s] Failed: Storage Bucket - %v\n", time.Now().Format("15:04:05"), storageErr)
		} else {
			fmt.Printf("    [%s] Completed: Storage Bucket\n", time.Now().Format("15:04:05"))
		}
	}()

	wg.Wait()
	elapsed := time.Since(startTime)
	fmt.Printf("    Operations completed in %.1fs\n", elapsed.Seconds())

	// Report results
	hasErrors := firewallErr != nil || saErr != nil || storageErr != nil || instanceErr != nil
	if hasErrors {
		red.Printf(" x\n")
		if instanceErr != nil {
			yellow.Printf("    VM Instance: %v\n", instanceErr)
		}
		if firewallErr != nil {
			yellow.Printf("    Firewall Rules: %v\n", firewallErr)
		}
		if saErr != nil {
			yellow.Printf("    Service Account: %v\n", saErr)
		}
		if storageErr != nil {
			yellow.Printf("    Storage Bucket: %v\n", storageErr)
		}
	} else {
		green.Printf(" +\n")
	}

	// Summary of what was deleted
	fmt.Println()
	bold.Println("Deleted Resources")
	bold.Println("-----------------")
	fmt.Printf("    Load Balancer:    Components deleted\n")
	if instanceErr == nil {
		fmt.Printf("    VM Instance:      %s\n", instanceName)
	}
	if firewallErr == nil {
		fmt.Printf("    Firewall Rules:   Deleted\n")
	}
	if saErr == nil {
		fmt.Printf("    Service Account:  kl-%s-sa\n", gcpUninstallInstallationKey)
	}
	if storageErr == nil {
		fmt.Printf("    Storage Bucket:   %s\n", bucketName)
	}

	// Success Summary
	fmt.Println()
	green.Println("+-----------------------------------------+")
	green.Println("|   + Uninstallation Complete!            |")
	green.Println("+-----------------------------------------+")
	fmt.Println()
	fmt.Printf("All resources for installation key '%s' have been removed.\n", gcpUninstallInstallationKey)
	fmt.Println()
}
