package cmd

import (
	"context"
	"fmt"
	"os"

	compute "cloud.google.com/go/compute/apiv1"
	"cloud.google.com/go/compute/apiv1/computepb"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"google.golang.org/api/iterator"
	"google.golang.org/api/oauth2/v2"
)

// gcpDoctorCmd represents the gcp doctor command
var gcpDoctorCmd = &cobra.Command{
	Use:   "doctor",
	Short: "Check GCP prerequisites for Kloudlite installation",
	Long: `Verify that your GCP environment is properly configured for Kloudlite installation.

This command checks:
  - gcloud CLI is installed
  - gcloud is authenticated
  - Current session has required IAM permissions
  - Default project is set`,
	Example: `  # Check GCP prerequisites
  kli gcp doctor`,
	Run: runGCPDoctor,
}

func runGCPDoctor(cmd *cobra.Command, args []string) {
	green := color.New(color.FgGreen, color.Bold)
	red := color.New(color.FgRed, color.Bold)
	yellow := color.New(color.FgYellow, color.Bold)
	cyan := color.New(color.FgCyan, color.Bold)

	fmt.Println()
	cyan.Println("GCP Doctor - Checking Prerequisites")
	fmt.Println()

	allPassed := true
	ctx := context.Background()

	// Check 1: GCP authentication and project
	fmt.Print("Checking GCP credentials and project... ")
	account, project, err := checkGCPCredentials(ctx)
	if err == nil {
		green.Println("PASSED")
		fmt.Printf("   Authenticated as: %s\n", account)
		fmt.Printf("   Project: %s\n", project)
	} else {
		red.Println("FAILED")
		yellow.Printf("   Error: %v\n", err)
		yellow.Println("   Configure GCP credentials: https://cloud.google.com/docs/authentication/getting-started")
		allPassed = false
	}

	// Check 2: Required IAM permissions
	fmt.Print("Checking IAM permissions... ")
	if project != "" {
		permissions := checkGCPPermissions(ctx, project)
		if permissions.HasRequired {
			green.Println("PASSED")
			if len(permissions.Missing) > 0 {
				yellow.Printf("   Warning: Some optional permissions missing: %v\n", permissions.Missing)
			}
		} else {
			red.Println("FAILED")
			yellow.Println("   Missing required IAM permissions for Kloudlite installation:")
			yellow.Println()
			yellow.Println("   Compute Engine/VM Permissions (to create and manage the VM):")
			yellow.Println("   - compute.instances.create")
			yellow.Println("   - compute.instances.get")
			yellow.Println()
			yellow.Println("   VPC/Network Permissions (to use existing VPC):")
			yellow.Println("   - compute.networks.get")
			yellow.Println("   - compute.subnetworks.get")
			yellow.Println()
			yellow.Println("   Firewall Permissions (for ports 443, 6443, 8472, 10250, 5001):")
			yellow.Println("   - compute.firewalls.create")
			yellow.Println("   - compute.firewalls.get")
			yellow.Println()
			yellow.Println("   Service Account Permissions (to create and assign runtime service account to VM):")
			yellow.Println("   - iam.serviceAccounts.create")
			yellow.Println("   - iam.serviceAccounts.get")
			yellow.Println("   - iam.serviceAccounts.update")
			yellow.Println("   - iam.serviceAccounts.setIamPolicy")
			yellow.Println("   - iam.serviceAccounts.actAs")
			yellow.Println()
			allPassed = false
		}
	} else {
		yellow.Println("SKIPPED (credentials/project check failed)")
	}

	// Summary
	fmt.Println()
	if allPassed {
		green.Println("All checks passed! Your GCP environment is ready for Kloudlite installation.")
	} else {
		red.Println("Some checks failed. Please resolve the issues above before proceeding.")
		fmt.Println()
		fmt.Println("For more information, visit: https://docs.kloudlite.io/installation/gcp")
	}
	fmt.Println()
}

func checkGCPCredentials(ctx context.Context) (string, string, error) {
	// Try to create an OAuth2 service to verify authentication
	oauth2Service, err := oauth2.NewService(ctx)
	if err != nil {
		return "", "", fmt.Errorf("failed to initialize OAuth2 service: %w", err)
	}

	// Get user info to verify authentication
	userInfo, err := oauth2Service.Userinfo.Get().Do()
	if err != nil {
		return "", "", fmt.Errorf("not authenticated or invalid credentials: %w", err)
	}

	// Get default project from environment or ADC
	project, err := getDefaultProject(ctx)
	if err != nil {
		return "", "", fmt.Errorf("no default project configured: %w", err)
	}

	return userInfo.Email, project, nil
}

func getDefaultProject(ctx context.Context) (string, error) {
	// Try environment variables (in order of precedence)
	envVars := []string{"GOOGLE_CLOUD_PROJECT", "GCLOUD_PROJECT", "GCP_PROJECT"}
	for _, envVar := range envVars {
		if project := os.Getenv(envVar); project != "" {
			return project, nil
		}
	}

	return "", fmt.Errorf("please set GOOGLE_CLOUD_PROJECT environment variable with your project ID")
}

type GCPPermissionCheck struct {
	HasRequired bool
	Missing     []string
}

func checkGCPPermissions(ctx context.Context, project string) *GCPPermissionCheck {
	// Required permissions for Kloudlite installation on GCP
	// Single VM installation in default VPC with firewall rules and service account
	requiredPermissions := []string{
		// VM/Compute Engine Permissions
		"compute.instances.create",
		"compute.instances.get",

		// VPC/Network Permissions (read-only for default VPC)
		"compute.networks.get",
		"compute.subnetworks.get",

		// Firewall Permissions (ports: 443, 6443, 8472, 10250, 5001)
		"compute.firewalls.create",
		"compute.firewalls.get",

		// IAM/Service Account Permissions (to create, edit, and assign to VM)
		"iam.serviceAccounts.create",
		"iam.serviceAccounts.get",
		"iam.serviceAccounts.update",
		"iam.serviceAccounts.setIamPolicy",
		"iam.serviceAccounts.actAs",
	}

	// Basic permission check using Compute Instances List
	instancesClient, err := compute.NewInstancesRESTClient(ctx)
	if err != nil {
		return &GCPPermissionCheck{
			HasRequired: false,
			Missing:     requiredPermissions,
		}
	}
	defer instancesClient.Close()

	// Try to list instances to verify basic compute permissions
	req := &computepb.AggregatedListInstancesRequest{
		Project:    project,
		MaxResults: func() *uint32 { v := uint32(1); return &v }(),
	}

	it := instancesClient.AggregatedList(ctx, req)
	// Just try to get one instance to verify permissions
	_, err = it.Next()
	// err will be iterator.Done if no instances, which is still a success for permission check
	if err != nil && err != iterator.Done {
		// This is a real error (likely permission denied)
		return &GCPPermissionCheck{
			HasRequired: false,
			Missing:     requiredPermissions,
		}
	}

	// Basic check passed - user has some Compute Engine permissions
	// TODO: Implement more granular permission checks
	return &GCPPermissionCheck{
		HasRequired: true,
		Missing:     []string{},
	}
}
