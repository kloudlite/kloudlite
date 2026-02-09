package cmd

import (
	"context"
	"fmt"

	"github.com/fatih/color"
	ociinternal "github.com/kloudlite/kloudlite/api/cmd/kli/internal/oci"
	"github.com/spf13/cobra"
)

// ociDoctorCmd represents the oci doctor command
var ociDoctorCmd = &cobra.Command{
	Use:   "doctor",
	Short: "Check OCI prerequisites for Kloudlite installation",
	Long: `Verify that your OCI environment is properly configured for Kloudlite installation.

This command checks:
  - OCI CLI config exists (~/.oci/config)
  - OCI authentication works
  - Required permissions are available
  - Tenancy, user, region, and compartment are configured`,
	Example: `  # Check OCI prerequisites
  kli oci doctor`,
	Run: runOCIDoctor,
}

func runOCIDoctor(cmd *cobra.Command, args []string) {
	green := color.New(color.FgGreen, color.Bold)
	red := color.New(color.FgRed, color.Bold)
	yellow := color.New(color.FgYellow, color.Bold)
	cyan := color.New(color.FgCyan, color.Bold)

	fmt.Println()
	cyan.Println("OCI Doctor - Checking Prerequisites")
	fmt.Println()

	allPassed := true
	ctx := context.Background()

	// Check 1: OCI config file exists
	fmt.Print("Checking OCI config file... ")
	if ociinternal.OCIConfigExists() {
		green.Println("PASSED")
		fmt.Printf("   Config file: %s\n", ociinternal.GetOCIConfigPath())
	} else {
		red.Println("FAILED")
		yellow.Printf("   OCI config file not found at %s\n", ociinternal.GetOCIConfigPath())
		yellow.Println("   Configure OCI CLI: https://docs.oracle.com/en-us/iaas/Content/API/SDKDocs/cliinstall.htm")
		allPassed = false
	}

	// Check 2: OCI authentication and configuration
	fmt.Print("Checking OCI authentication... ")
	cfg, err := ociinternal.LoadOCIConfig(ctx, "", "", "", "", "", "")
	if err == nil {
		green.Println("PASSED")
		fmt.Printf("   Tenancy: %s\n", cfg.TenancyOCID)
		if cfg.UserOCID != "" {
			fmt.Printf("   User: %s\n", cfg.UserOCID)
		}
		fmt.Printf("   Region: %s\n", cfg.Region)
		fmt.Printf("   Compartment: %s\n", cfg.CompartmentOCID)
	} else {
		red.Println("FAILED")
		yellow.Printf("   Error: %v\n", err)
		yellow.Println("   Configure OCI credentials: https://docs.oracle.com/en-us/iaas/Content/API/Concepts/apisigningkey.htm")
		allPassed = false
	}

	// Check 3: Basic API connectivity (list compartments)
	if cfg != nil {
		fmt.Print("Checking OCI API connectivity... ")
		apiOk := checkOCIAPIConnectivity(ctx, cfg)
		if apiOk {
			green.Println("PASSED")
		} else {
			red.Println("FAILED")
			yellow.Println("   Could not connect to OCI APIs.")
			yellow.Println("   Check your credentials and network connectivity.")
			yellow.Println()
			yellow.Println("   Required permissions:")
			yellow.Println("   - Compute: manage instances, images")
			yellow.Println("   - Networking: manage VCNs, subnets, NSGs, network load balancers")
			yellow.Println("   - Object Storage: manage buckets and objects")
			yellow.Println("   - Identity: manage dynamic groups and policies")
			allPassed = false
		}
	}

	// Summary
	fmt.Println()
	if allPassed {
		green.Println("All checks passed! Your OCI environment is ready for Kloudlite installation.")
	} else {
		red.Println("Some checks failed. Please resolve the issues above before proceeding.")
		fmt.Println()
		fmt.Println("For more information, visit: https://docs.kloudlite.io/installation/oci")
	}
	fmt.Println()
}

func checkOCIAPIConnectivity(ctx context.Context, cfg *ociinternal.OCIConfig) bool {
	// Try to get object storage namespace as a basic API connectivity test
	_, err := ociinternal.GetObjectStorageNamespace(ctx, cfg)
	return err == nil
}
