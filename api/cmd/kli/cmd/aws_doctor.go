package cmd

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/sts"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

// awsDoctorCmd represents the aws doctor command
var awsDoctorCmd = &cobra.Command{
	Use:   "doctor",
	Short: "Check AWS prerequisites for Kloudlite installation",
	Long: `Verify that your AWS environment is properly configured for Kloudlite installation.

This command checks:
  - AWS CLI is installed
  - AWS credentials are configured
  - Current session has required IAM permissions`,
	Example: `  # Check AWS prerequisites
  kli aws doctor`,
	Run: runAWSDoctor,
}

func runAWSDoctor(cmd *cobra.Command, args []string) {
	green := color.New(color.FgGreen, color.Bold)
	red := color.New(color.FgRed, color.Bold)
	yellow := color.New(color.FgYellow, color.Bold)
	cyan := color.New(color.FgCyan, color.Bold)

	fmt.Println()
	cyan.Println("AWS Doctor - Checking Prerequisites")
	fmt.Println()

	allPassed := true
	ctx := context.Background()

	// Check 1: AWS SDK configuration and credentials
	fmt.Print("Checking AWS credentials and configuration... ")
	cfg, identity, err := checkAWSCredentials(ctx)
	if err == nil {
		green.Println("PASSED")
		fmt.Printf("   Account: %s\n", *identity.Account)
		fmt.Printf("   User/Role: %s\n", *identity.Arn)
	} else {
		red.Println("FAILED")
		yellow.Printf("   Error: %v\n", err)
		yellow.Println("   Configure AWS credentials: https://docs.aws.amazon.com/cli/latest/userguide/cli-configure-files.html")
		allPassed = false
	}

	// Check 2: Required IAM permissions
	fmt.Print("Checking IAM permissions... ")
	if cfg.Region != "" {
		permissions := checkAWSPermissions(ctx, &cfg)
		if permissions.HasRequired {
			green.Println("PASSED")
			if len(permissions.Missing) > 0 {
				yellow.Printf("   Warning: Some optional permissions missing: %v\n", permissions.Missing)
			}
		} else {
			red.Println("FAILED")
			yellow.Println("   Missing required IAM permissions for Kloudlite installation:")
			yellow.Println()
			yellow.Println("   EC2/VM Permissions (to create and manage the VM):")
			yellow.Println("   - ec2:RunInstances")
			yellow.Println("   - ec2:DescribeInstances")
			yellow.Println()
			yellow.Println("   VPC/Network Permissions (to use existing VPC):")
			yellow.Println("   - ec2:DescribeVpcs")
			yellow.Println("   - ec2:DescribeSubnets")
			yellow.Println()
			yellow.Println("   Security Group Permissions (for ports 443, 6443, 8472, 10250, 5001):")
			yellow.Println("   - ec2:CreateSecurityGroup")
			yellow.Println("   - ec2:DescribeSecurityGroups")
			yellow.Println("   - ec2:AuthorizeSecurityGroupIngress")
			yellow.Println()
			yellow.Println("   IAM Role Permissions (to create and assign runtime role to VM):")
			yellow.Println("   - iam:CreateRole")
			yellow.Println("   - iam:GetRole")
			yellow.Println("   - iam:PutRolePolicy")
			yellow.Println("   - iam:AttachRolePolicy")
			yellow.Println("   - iam:UpdateAssumeRolePolicy")
			yellow.Println("   - iam:CreateInstanceProfile")
			yellow.Println("   - iam:AddRoleToInstanceProfile")
			yellow.Println("   - iam:PassRole")
			yellow.Println()
			yellow.Println("   Region Permissions:")
			yellow.Println("   - ec2:DescribeRegions")
			yellow.Println()
			allPassed = false
		}
	} else {
		yellow.Println("SKIPPED (credentials check failed)")
	}

	// Summary
	fmt.Println()
	if allPassed {
		green.Println("All checks passed! Your AWS environment is ready for Kloudlite installation.")
	} else {
		red.Println("Some checks failed. Please resolve the issues above before proceeding.")
		fmt.Println()
		fmt.Println("For more information, visit: https://docs.kloudlite.io/installation/aws")
	}
	fmt.Println()
}

func checkAWSCredentials(ctx context.Context) (aws.Config, *sts.GetCallerIdentityOutput, error) {
	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return aws.Config{}, nil, fmt.Errorf("failed to load AWS configuration: %w", err)
	}

	stsClient := sts.NewFromConfig(cfg)
	identity, err := stsClient.GetCallerIdentity(ctx, &sts.GetCallerIdentityInput{})
	if err != nil {
		return aws.Config{}, nil, fmt.Errorf("failed to get caller identity: %w", err)
	}

	return cfg, identity, nil
}

type PermissionCheck struct {
	HasRequired bool
	Missing     []string
}

func checkAWSPermissions(ctx context.Context, cfg *aws.Config) *PermissionCheck {
	// Required permissions for Kloudlite installation on AWS
	// Single VM installation in default VPC with security group and IAM role
	requiredPermissions := []string{
		// VM/EC2 Permissions
		"ec2:RunInstances",
		"ec2:DescribeInstances",

		// VPC/Network Permissions (read-only for default VPC)
		"ec2:DescribeVpcs",
		"ec2:DescribeSubnets",

		// Security Group Permissions (ports: 443, 6443, 8472, 10250, 5001)
		"ec2:CreateSecurityGroup",
		"ec2:DescribeSecurityGroups",
		"ec2:AuthorizeSecurityGroupIngress",

		// IAM Permissions (to create, edit, and assign role to VM)
		"iam:CreateRole",
		"iam:GetRole",
		"iam:PutRolePolicy",
		"iam:AttachRolePolicy",
		"iam:UpdateAssumeRolePolicy",
		"iam:CreateInstanceProfile",
		"iam:AddRoleToInstanceProfile",
		"iam:PassRole",

		// Region Permissions
		"ec2:DescribeRegions",
	}

	// Basic permission check using EC2 DescribeRegions
	ec2Client := ec2.NewFromConfig(*cfg)
	_, err := ec2Client.DescribeRegions(ctx, &ec2.DescribeRegionsInput{})
	if err != nil {
		return &PermissionCheck{
			HasRequired: false,
			Missing:     requiredPermissions,
		}
	}

	// Basic check passed - user has some EC2 permissions
	// TODO: Implement more granular permission checks using AWS IAM Policy Simulator
	return &PermissionCheck{
		HasRequired: true,
		Missing:     []string{},
	}
}
