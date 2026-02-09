package main

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	ociinternal "github.com/kloudlite/kloudlite/api/cmd/kli/internal/oci"
)

func runUninstall(ctx context.Context, cfg *Config) error {
	totalSteps := 4
	step := 0

	nextStep := func(desc string) {
		step++
		log.Printf("[STEP %d/%d] %s", step, totalSteps, desc)
	}

	// Load OCI config
	nextStep("Loading OCI configuration...")
	ociCfg, err := ociinternal.LoadOCIConfig(ctx, cfg.OCITenancy, cfg.OCIUser, cfg.OCIRegion, cfg.OCICompartment, cfg.OCIFingerprint, "")
	if err != nil {
		return fmt.Errorf("failed to load OCI config: %w", err)
	}
	log.Printf("  Tenancy:     %s", ociCfg.TenancyOCID)
	log.Printf("  Region:      %s", ociCfg.Region)
	log.Printf("  Compartment: %s", ociCfg.CompartmentOCID)

	// Phase 1: Find instances and delete NLB + instances in parallel
	nextStep("Phase 1: Deleting NLB and instances...")

	instanceIDs, findErr := ociinternal.FindAllInstancesByTag(ctx, ociCfg, cfg.InstallationKey)
	if findErr != nil {
		log.Printf("  Warning: Failed to discover instances by tag: %v", findErr)
	} else {
		log.Printf("  Found %d instance(s) with installation-id=%s", len(instanceIDs), cfg.InstallationKey)
	}

	var phase1Wg sync.WaitGroup
	var instanceErr, nlbErr error

	phase1Wg.Add(2)

	go func() {
		defer phase1Wg.Done()
		log.Printf("  [parallel] Starting NLB cleanup")
		nlbErr = ociinternal.DeleteNetworkLoadBalancer(ctx, ociCfg, cfg.InstallationKey)
		if nlbErr != nil {
			log.Printf("  [parallel] NLB cleanup warning: %v", nlbErr)
		} else {
			log.Printf("  [parallel] NLB cleanup completed")
		}
	}()

	go func() {
		defer phase1Wg.Done()
		if findErr != nil || len(instanceIDs) == 0 {
			if findErr == nil {
				log.Printf("  [parallel] No instances found to terminate")
			}
			return
		}
		log.Printf("  [parallel] Terminating %d instance(s)", len(instanceIDs))
		instanceErr = ociinternal.TerminateAllInstances(ctx, ociCfg, instanceIDs)
		if instanceErr != nil {
			log.Printf("  [parallel] Instance termination warning: %v", instanceErr)
		} else {
			log.Printf("  [parallel] %d instance(s) terminated", len(instanceIDs))
		}
	}()

	phase1Wg.Wait()

	// Phase 2: Parallel cleanup of remaining resources
	nextStep("Phase 2: Deleting NSG, IAM, and Storage...")

	var wg sync.WaitGroup
	var nsgErr, dgErr, policyErr, storageErr error

	bucketName := ociinternal.GetBucketName(cfg.InstallationKey)
	vcnID, _, _ := ociinternal.GetDefaultVCN(ctx, ociCfg)

	startTime := time.Now()

	wg.Add(3)

	// NSG
	go func() {
		defer wg.Done()
		log.Printf("  [parallel] Starting NSG deletion")
		nsgErr = ociinternal.DeleteNSG(ctx, ociCfg, vcnID, cfg.InstallationKey)
		if nsgErr != nil {
			log.Printf("  [parallel] NSG deletion failed: %v", nsgErr)
		} else {
			log.Printf("  [parallel] NSG deletion completed")
		}
	}()

	// Dynamic Group + Policy
	go func() {
		defer wg.Done()
		log.Printf("  [parallel] Starting Dynamic Group + Policy deletion")
		policyErr = ociinternal.DeletePolicy(ctx, ociCfg, cfg.InstallationKey)
		if policyErr != nil {
			log.Printf("  [parallel] Policy deletion failed: %v", policyErr)
		}
		dgErr = ociinternal.DeleteDynamicGroup(ctx, ociCfg, cfg.InstallationKey)
		if dgErr != nil {
			log.Printf("  [parallel] Dynamic Group deletion failed: %v", dgErr)
		}
		if policyErr == nil && dgErr == nil {
			log.Printf("  [parallel] Dynamic Group + Policy deletion completed")
		}
	}()

	// Storage Bucket
	go func() {
		defer wg.Done()
		log.Printf("  [parallel] Starting Storage Bucket deletion")
		storageErr = ociinternal.DeleteStorageBucket(ctx, ociCfg, bucketName)
		if storageErr != nil {
			log.Printf("  [parallel] Storage Bucket deletion failed: %v", storageErr)
		} else {
			log.Printf("  [parallel] Storage Bucket deletion completed")
		}
	}()

	wg.Wait()
	log.Printf("  Phase 2 completed in %.1fs", time.Since(startTime).Seconds())

	// Report results
	nextStep("Summarizing results...")

	var errors []string
	if instanceErr != nil {
		errors = append(errors, fmt.Sprintf("Instances: %v", instanceErr))
	}
	if nsgErr != nil {
		errors = append(errors, fmt.Sprintf("NSG: %v", nsgErr))
	}
	if dgErr != nil {
		errors = append(errors, fmt.Sprintf("Dynamic Group: %v", dgErr))
	}
	if policyErr != nil {
		errors = append(errors, fmt.Sprintf("Policy: %v", policyErr))
	}
	if storageErr != nil {
		errors = append(errors, fmt.Sprintf("Storage Bucket: %v", storageErr))
	}

	log.Printf("=== Uninstallation Summary ===")
	log.Printf("  NLB:             %s", statusStr(nlbErr))
	log.Printf("  Instances:       %d terminated", len(instanceIDs))
	log.Printf("  NSG:             %s", statusStr(nsgErr))
	log.Printf("  Dynamic Group:   %s", statusStr(dgErr))
	log.Printf("  Policy:          %s", statusStr(policyErr))
	log.Printf("  Storage Bucket:  %s", statusStr(storageErr))

	if len(errors) > 0 {
		return fmt.Errorf("uninstallation completed with errors: %v", errors)
	}

	return nil
}

func statusStr(err error) string {
	if err != nil {
		return fmt.Sprintf("FAILED (%v)", err)
	}
	return "Deleted"
}
