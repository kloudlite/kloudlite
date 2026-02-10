package cmd

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	ociinternal "github.com/kloudlite/kloudlite/api/cmd/kli/internal/oci"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

var ociUninstallCmd = &cobra.Command{
	Use:   "uninstall",
	Short: "Uninstall Kloudlite from OCI",
	Long: `Uninstall Kloudlite from OCI by removing all resources created during installation.

This command will:
  - Delete Reserved Public IP
  - Terminate ALL Compute Instances (including workmachines and sub-resources)
  - Delete Network Security Group
  - Delete Dynamic Group and IAM Policy
  - Delete Object Storage bucket and all backups

All resources are identified by the installation-id freeform tag.`,
	Example: `  # Uninstall using OCI defaults (from ~/.oci/config)
  kli oci uninstall --installation-key prod

  # Uninstall with specific compartment
  kli oci uninstall --installation-key prod --compartment ocid1.compartment.oc1..xxx`,
	Run: runOCIUninstall,
}

var (
	ociUninstallTenancy         string
	ociUninstallUser            string
	ociUninstallRegion          string
	ociUninstallCompartment     string
	ociUninstallFingerprint     string
	ociUninstallKeyFile         string
	ociUninstallInstallationKey string
)

func init() {
	ociUninstallCmd.Flags().StringVar(&ociUninstallTenancy, "tenancy", "", "OCI tenancy OCID (reads from OCI_CLI_TENANCY or ~/.oci/config)")
	ociUninstallCmd.Flags().StringVar(&ociUninstallUser, "user", "", "OCI user OCID (reads from OCI_CLI_USER or ~/.oci/config)")
	ociUninstallCmd.Flags().StringVar(&ociUninstallRegion, "region", "", "OCI region (reads from OCI_CLI_REGION or ~/.oci/config)")
	ociUninstallCmd.Flags().StringVar(&ociUninstallCompartment, "compartment", "", "OCI compartment OCID (defaults to tenancy)")
	ociUninstallCmd.Flags().StringVar(&ociUninstallFingerprint, "fingerprint", "", "OCI API key fingerprint")
	ociUninstallCmd.Flags().StringVar(&ociUninstallKeyFile, "key-file", "", "OCI API private key file path")
	ociUninstallCmd.Flags().StringVar(&ociUninstallInstallationKey, "installation-key", "", "Installation key to identify this installation (required)")
	ociUninstallCmd.MarkFlagRequired("installation-key")
}

func runOCIUninstall(cmd *cobra.Command, args []string) {
	green := color.New(color.FgGreen, color.Bold)
	red := color.New(color.FgRed, color.Bold)
	yellow := color.New(color.FgYellow, color.Bold)
	cyan := color.New(color.FgCyan, color.Bold)
	bold := color.New(color.Bold)

	// Header
	fmt.Println()
	cyan.Println("+-----------------------------------------+")
	cyan.Println("|   Kloudlite OCI Uninstallation          |")
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
	fmt.Printf("  Installation Key: %s\n", ociUninstallInstallationKey)

	cfg, err := ociinternal.LoadOCIConfig(ctx, ociUninstallTenancy, ociUninstallUser, ociUninstallRegion, ociUninstallCompartment, ociUninstallFingerprint, ociUninstallKeyFile)
	if err != nil {
		red.Printf("x\n")
		yellow.Printf("  Error: %v\n\n", err)
		os.Exit(1)
	}
	green.Printf("  Tenancy:     %s\n", cfg.TenancyOCID)
	green.Printf("  Region:      %s\n", cfg.Region)
	green.Printf("  Compartment: %s\n", cfg.CompartmentOCID)
	fmt.Println()

	// Resource Cleanup
	bold.Println("Removing Resources")
	bold.Println("------------------")

	fmt.Printf("  o Cleaning up resources...\n")

	startTime := time.Now()

	// Phase 1: Delete Reserved IP and ALL instances (by tag) in parallel
	var instanceErr, ipErr error

	// Find ALL instances with the installation-id tag (includes workmachines and sub-resources)
	instanceIDs, findErr := ociinternal.FindAllInstancesByTag(ctx, cfg, ociUninstallInstallationKey)
	if findErr != nil {
		fmt.Printf("    [%s] Warning: Failed to discover instances by tag - %v\n", time.Now().Format("15:04:05"), findErr)
	} else {
		fmt.Printf("    [%s] Found %d instance(s) with installation-id=%s\n", time.Now().Format("15:04:05"), len(instanceIDs), ociUninstallInstallationKey)
	}

	var phase1Wg sync.WaitGroup
	phase1Wg.Add(2)
	go func() {
		defer phase1Wg.Done()
		fmt.Printf("    [%s] Starting: Reserved IP cleanup\n", time.Now().Format("15:04:05"))
		ipErr = ociinternal.DeleteReservedPublicIP(ctx, cfg, ociUninstallInstallationKey)
		if ipErr != nil {
			fmt.Printf("    [%s] Warning: Reserved IP cleanup - %v\n", time.Now().Format("15:04:05"), ipErr)
		} else {
			fmt.Printf("    [%s] Completed: Reserved IP cleanup\n", time.Now().Format("15:04:05"))
		}
	}()
	go func() {
		defer phase1Wg.Done()
		if findErr != nil || len(instanceIDs) == 0 {
			if findErr == nil {
				fmt.Printf("    [%s] No instances found to terminate\n", time.Now().Format("15:04:05"))
			}
			return
		}
		fmt.Printf("    [%s] Starting: Terminating %d instance(s)\n", time.Now().Format("15:04:05"), len(instanceIDs))
		instanceErr = ociinternal.TerminateAllInstances(ctx, cfg, instanceIDs)
		if instanceErr != nil {
			fmt.Printf("    [%s] Warning: Instance termination - %v\n", time.Now().Format("15:04:05"), instanceErr)
		} else {
			fmt.Printf("    [%s] Completed: %d instance(s) terminated\n", time.Now().Format("15:04:05"), len(instanceIDs))
		}
	}()
	phase1Wg.Wait()

	// Phase 2: Parallel cleanup of remaining resources
	var wg sync.WaitGroup
	var nsgErr, dgErr, policyErr, storageErr error

	bucketName := ociinternal.GetBucketName(ociUninstallInstallationKey)

	// Find VCN for NSG deletion
	vcnID, _, _ := ociinternal.GetDefaultVCN(ctx, cfg)

	// NSG (parallel)
	wg.Add(1)
	go func() {
		defer wg.Done()
		fmt.Printf("    [%s] Starting: NSG deletion\n", time.Now().Format("15:04:05"))
		nsgErr = ociinternal.DeleteNSG(ctx, cfg, vcnID, ociUninstallInstallationKey)
		if nsgErr != nil {
			fmt.Printf("    [%s] Failed: NSG - %v\n", time.Now().Format("15:04:05"), nsgErr)
		} else {
			fmt.Printf("    [%s] Completed: NSG\n", time.Now().Format("15:04:05"))
		}
	}()

	// Dynamic Group + Policy (parallel)
	wg.Add(1)
	go func() {
		defer wg.Done()
		fmt.Printf("    [%s] Starting: Dynamic Group + Policy deletion\n", time.Now().Format("15:04:05"))
		policyErr = ociinternal.DeletePolicy(ctx, cfg, ociUninstallInstallationKey)
		if policyErr != nil {
			fmt.Printf("    [%s] Failed: Policy - %v\n", time.Now().Format("15:04:05"), policyErr)
		}
		dgErr = ociinternal.DeleteDynamicGroup(ctx, cfg, ociUninstallInstallationKey)
		if dgErr != nil {
			fmt.Printf("    [%s] Failed: Dynamic Group - %v\n", time.Now().Format("15:04:05"), dgErr)
		}
		if policyErr == nil && dgErr == nil {
			fmt.Printf("    [%s] Completed: Dynamic Group + Policy\n", time.Now().Format("15:04:05"))
		}
	}()

	// Storage Bucket (parallel)
	wg.Add(1)
	go func() {
		defer wg.Done()
		fmt.Printf("    [%s] Starting: Storage Bucket deletion\n", time.Now().Format("15:04:05"))
		storageErr = ociinternal.DeleteStorageBucket(ctx, cfg, bucketName)
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
	hasErrors := nsgErr != nil || dgErr != nil || policyErr != nil || storageErr != nil || instanceErr != nil || ipErr != nil
	if hasErrors {
		red.Printf(" x\n")
		if instanceErr != nil {
			yellow.Printf("    Instances: %v\n", instanceErr)
		}
		if nsgErr != nil {
			yellow.Printf("    NSG: %v\n", nsgErr)
		}
		if dgErr != nil {
			yellow.Printf("    Dynamic Group: %v\n", dgErr)
		}
		if policyErr != nil {
			yellow.Printf("    Policy: %v\n", policyErr)
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
	if ipErr == nil {
		fmt.Printf("    Reserved IP:     Deleted\n")
	}
	if instanceErr == nil && findErr == nil && len(instanceIDs) > 0 {
		fmt.Printf("    Instances:       %d terminated\n", len(instanceIDs))
		for _, id := range instanceIDs {
			fmt.Printf("                     - %s\n", id)
		}
	}
	if nsgErr == nil {
		fmt.Printf("    NSG:             Deleted\n")
	}
	if dgErr == nil {
		fmt.Printf("    Dynamic Group:   Deleted\n")
	}
	if policyErr == nil {
		fmt.Printf("    Policy:          Deleted\n")
	}
	if storageErr == nil {
		fmt.Printf("    Storage Bucket:  %s\n", bucketName)
	}

	// Success Summary
	fmt.Println()
	green.Println("+-----------------------------------------+")
	green.Println("|   + Uninstallation Complete!            |")
	green.Println("+-----------------------------------------+")
	fmt.Println()
	fmt.Printf("All resources for installation key '%s' have been removed.\n", ociUninstallInstallationKey)
	fmt.Println()
}
