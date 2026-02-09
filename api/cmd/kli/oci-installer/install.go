package main

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/kloudlite/kloudlite/api/cmd/kli/internal/console"
	k8sinternal "github.com/kloudlite/kloudlite/api/cmd/kli/internal/k8s"
	ociinternal "github.com/kloudlite/kloudlite/api/cmd/kli/internal/oci"
)

func runInstall(ctx context.Context, cfg *Config) error {
	totalSteps := 9
	if cfg.SkipLB {
		totalSteps = 7
	}
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

	// Console API client
	consoleClient := console.NewClientWithBase(cfg.ConsoleBaseURL)

	// Verify installation
	nextStep("Verifying installation key...")
	verifyResult, err := k8sinternal.VerifyInstallation(ctx, cfg.InstallationKey, &k8sinternal.VerifyInstallationOptions{
		Provider: "oci",
		Region:   ociCfg.Region,
	})
	if err != nil {
		return fmt.Errorf("failed to verify installation: %w", err)
	}
	log.Printf("  Secret key obtained successfully")

	secretKey := verifyResult.SecretKey
	var fullDomain string

	if !cfg.SkipLB {
		if verifyResult.Subdomain == "" {
			return fmt.Errorf("no subdomain configured for this installation; configure a subdomain in the console before running install")
		}
		fullDomain = console.GetFullDomain(verifyResult.Subdomain)
		log.Printf("  Subdomain: %s", verifyResult.Subdomain)
		log.Printf("  URL: https://%s", fullDomain)
	}

	// Find Ubuntu image
	nextStep("Finding Ubuntu 24.04 image...")
	imageID, imageName, err := ociinternal.FindUbuntuImage(ctx, ociCfg)
	if err != nil {
		return fmt.Errorf("failed to find Ubuntu image: %w", err)
	}
	log.Printf("  Image: %s", imageName)

	// Network setup
	nextStep("Setting up network...")
	vcnID, vcnCIDR, err := ociinternal.GetDefaultVCN(ctx, ociCfg)
	if err != nil {
		return fmt.Errorf("failed to get default VCN: %w", err)
	}

	subnetID, subnetCIDR, err := ociinternal.GetDefaultSubnet(ctx, ociCfg, vcnID)
	if err != nil {
		return fmt.Errorf("failed to get default subnet: %w", err)
	}
	log.Printf("  VCN: %s", vcnID)
	log.Printf("  Subnet: %s (CIDR: %s)", subnetID, subnetCIDR)

	// Ensure subnet security list rules
	if err := ociinternal.EnsureSubnetSecurityListRules(ctx, ociCfg, subnetID); err != nil {
		return fmt.Errorf("failed to update subnet security list: %w", err)
	}
	log.Printf("  Subnet security list rules updated")

	// Parallel resource creation
	nextStep("Creating resources in parallel (NSG, Storage, IAM)...")
	var (
		wg                                   sync.WaitGroup
		nsgID                                string
		bucketName                           = ociinternal.GetBucketName(cfg.InstallationKey)
		nsgErr, storageErr, dgErr, policyErr error
	)

	startTime := time.Now()

	wg.Add(3)

	// NSG
	go func() {
		defer wg.Done()
		log.Printf("  [parallel] Starting NSG creation")
		nsgID, nsgErr = ociinternal.EnsureNSG(ctx, ociCfg, vcnID, vcnCIDR, cfg.InstallationKey)
		if nsgErr != nil {
			log.Printf("  [parallel] NSG failed: %v", nsgErr)
		} else {
			log.Printf("  [parallel] NSG completed")
		}
	}()

	// Storage Bucket
	go func() {
		defer wg.Done()
		log.Printf("  [parallel] Starting Storage Bucket creation")
		storageErr = ociinternal.EnsureStorageBucket(ctx, ociCfg, bucketName, cfg.InstallationKey)
		if storageErr != nil {
			log.Printf("  [parallel] Storage Bucket failed: %v", storageErr)
		} else {
			log.Printf("  [parallel] Storage Bucket completed")
		}
	}()

	// Dynamic Group + Policy
	go func() {
		defer wg.Done()
		log.Printf("  [parallel] Starting Dynamic Group + Policy creation")
		_, dgErr = ociinternal.EnsureDynamicGroup(ctx, ociCfg, cfg.InstallationKey)
		if dgErr != nil {
			log.Printf("  [parallel] Dynamic Group failed: %v", dgErr)
			return
		}
		_, policyErr = ociinternal.EnsurePolicy(ctx, ociCfg, cfg.InstallationKey)
		if policyErr != nil {
			log.Printf("  [parallel] Policy failed: %v", policyErr)
		} else {
			log.Printf("  [parallel] Dynamic Group + Policy completed")
		}
	}()

	wg.Wait()
	log.Printf("  Parallel operations completed in %.1fs", time.Since(startTime).Seconds())

	// Check for errors
	if nsgErr != nil {
		return fmt.Errorf("NSG creation failed: %w", nsgErr)
	}
	if storageErr != nil {
		return fmt.Errorf("Storage Bucket creation failed: %w", storageErr)
	}
	if dgErr != nil {
		return fmt.Errorf("Dynamic Group creation failed: %w", dgErr)
	}
	if policyErr != nil {
		return fmt.Errorf("Policy creation failed: %w", policyErr)
	}

	log.Printf("  NSG:            %s", nsgID)
	log.Printf("  Storage Bucket: %s", bucketName)
	log.Printf("  Dynamic Group:  Created")
	log.Printf("  Policy:         Created")

	// Instance launch
	nextStep("Launching OCI instance (VM.Standard.E4.Flex 1 OCPU / 8GB)...")
	k3sToken, err := ociinternal.GenerateK3sToken()
	if err != nil {
		return fmt.Errorf("failed to generate K3s token: %w", err)
	}

	// No SSH key injection in server-side mode (no dev mode)
	instanceID, err := ociinternal.LaunchInstance(ctx, ociCfg, imageID, subnetID, nsgID, secretKey, bucketName, k3sToken, cfg.InstallationKey, fullDomain, "", cfg.EnableDeletionProtection)
	if err != nil {
		return fmt.Errorf("failed to launch instance: %w", err)
	}
	log.Printf("  Instance ID: %s", instanceID)

	log.Printf("  Waiting for instance to be ready...")
	publicIP, privateIP, err := ociinternal.WaitForInstance(ctx, ociCfg, instanceID)
	if err != nil {
		return fmt.Errorf("instance failed to become ready: %w", err)
	}
	log.Printf("  Public IP:  %s", publicIP)
	log.Printf("  Private IP: %s", privateIP)

	// Load Balancer setup
	if !cfg.SkipLB {
		nextStep("Creating Network Load Balancer...")
		_, nlbIP, err := ociinternal.CreateNetworkLoadBalancer(ctx, ociCfg, subnetID, instanceID, privateIP, cfg.InstallationKey)
		if err != nil {
			return fmt.Errorf("failed to create NLB: %w", err)
		}
		log.Printf("  NLB IP: %s", nlbIP)

		// DNS Configuration
		nextStep("Configuring DNS (Cloudflare proxied)...")
		_, err = consoleClient.ConfigureRootDNS(ctx, cfg.InstallationKey, secretKey, nlbIP, "a", true)
		if err != nil {
			return fmt.Errorf("failed to configure DNS: %w", err)
		}
		log.Printf("  DNS configured for %s -> %s", fullDomain, nlbIP)
	}

	// Summary
	log.Printf("=== Installation Summary ===")
	log.Printf("  Instance ID:    %s", instanceID)
	log.Printf("  Public IP:      %s", publicIP)
	log.Printf("  Private IP:     %s", privateIP)
	log.Printf("  Region:         %s", ociCfg.Region)
	log.Printf("  Compartment:    %s", ociCfg.CompartmentOCID)
	if !cfg.SkipLB {
		log.Printf("  Custom Domain:  https://%s", fullDomain)
	}

	return nil
}
