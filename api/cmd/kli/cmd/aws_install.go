package cmd

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	awsinternal "github.com/kloudlite/kloudlite/api/cmd/kli/internal/aws"
	k8sinternal "github.com/kloudlite/kloudlite/api/cmd/kli/internal/k8s"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

var awsInstallCmd = &cobra.Command{
	Use:   "install",
	Short: "Install Kloudlite on AWS",
	Long: `Install Kloudlite on AWS by creating all necessary resources.

This command will:
  - Find Ubuntu 24.04 LTS AMD64 AMI in the region
  - Create IAM role 'kl-{installation-key}-role' with required permissions (including S3)
  - Create S3 bucket 'kl-{installation-key}-backups' for K3s database backups
  - Create security group 'kl-{installation-key}-sg' with required ports
  - Create SSH key pair 'kl-{installation-key}-key' and save to ~/.kl/kl-{installation-key}-key.pem
  - Launch t3.medium EC2 instance with 100GB storage
  - Configure instance in default VPC with public IP
  - Setup automated K3s SQLite backup to S3 every 30 minutes`,
	Example: `  # Install using default AWS region from config
  kli aws install --installation-key prod

  # Install in a specific region
  kli aws install --installation-key staging --region us-west-2`,
	Run: runAWSInstall,
}

var (
	region                      string
	installationKey             string
	enableTerminationProtection bool
)

func init() {
	awsInstallCmd.Flags().StringVar(&region, "region", "", "AWS region (uses default from AWS config if not specified)")
	awsInstallCmd.Flags().StringVar(&installationKey, "installation-key", "", "Installation key to identify this installation (required)")
	awsInstallCmd.Flags().BoolVar(&enableTerminationProtection, "enable-termination-protection", true, "Enable EC2 termination protection (default: true)")
	awsInstallCmd.MarkFlagRequired("installation-key")
}

func runAWSInstall(cmd *cobra.Command, args []string) {
	green := color.New(color.FgGreen, color.Bold)
	red := color.New(color.FgRed, color.Bold)
	yellow := color.New(color.FgYellow, color.Bold)
	cyan := color.New(color.FgCyan, color.Bold)
	bold := color.New(color.Bold)

	// Header
	fmt.Println()
	cyan.Println("╭─────────────────────────────────────────╮")
	cyan.Println("│   Kloudlite AWS Installation            │")
	cyan.Println("╰─────────────────────────────────────────╯")
	fmt.Println()

	ctx := context.Background()

	// Setup signal handling for cleanup on interruption
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	var createdResources struct {
		sync.Mutex
		instanceID string
		sgID       string
		iamCreated bool
		bucketName string
	}

	go func() {
		<-sigChan
		fmt.Println()
		yellow.Println("\n⚠ Installation interrupted! Cleaning up resources...")

		// Load config for cleanup
		cfg, err := awsinternal.LoadAWSConfig(context.Background(), region)
		if err != nil {
			red.Printf("Failed to load AWS config for cleanup: %v\n", err)
			os.Exit(1)
		}

		createdResources.Lock()
		defer createdResources.Unlock()

		ec2Client := ec2.NewFromConfig(cfg)

		// Cleanup in reverse order
		if createdResources.instanceID != "" {
			fmt.Printf("  Terminating instance %s...\n", createdResources.instanceID)
			// Disable termination protection first
			_, _ = ec2Client.ModifyInstanceAttribute(context.Background(), &ec2.ModifyInstanceAttributeInput{
				InstanceId: aws.String(createdResources.instanceID),
				DisableApiTermination: &types.AttributeBooleanValue{
					Value: aws.Bool(false),
				},
			})
			// Then terminate
			_, _ = ec2Client.TerminateInstances(context.Background(), &ec2.TerminateInstancesInput{
				InstanceIds: []string{createdResources.instanceID},
			})
		}
		if createdResources.sgID != "" {
			fmt.Printf("  Deleting security group...\n")
			deleteSecurityGroup(context.Background(), cfg, installationKey)
		}
		if createdResources.iamCreated {
			fmt.Printf("  Deleting IAM resources...\n")
			deleteInstanceProfile(context.Background(), cfg, installationKey)
			deleteIAMRole(context.Background(), cfg, installationKey)
		}
		if createdResources.bucketName != "" {
			fmt.Printf("  Deleting S3 bucket...\n")
			awsinternal.DeleteS3Bucket(context.Background(), cfg, createdResources.bucketName)
		}

		yellow.Println("Cleanup completed. Exiting...")
		os.Exit(130) // Standard exit code for SIGINT
	}()

	// Configuration
	bold.Println("Configuration")
	bold.Println("─────────────")
	fmt.Printf("  Installation Key: %s\n", installationKey)
	fmt.Printf("  Region:          ")
	cfg, err := awsinternal.LoadAWSConfig(ctx, region)
	if err != nil {
		red.Printf("✗\n")
		yellow.Printf("  Error: %v\n\n", err)
		os.Exit(1)
	}
	green.Printf("✓ %s\n", cfg.Region)
	fmt.Println()

	// Infrastructure Setup
	bold.Println("Infrastructure Setup")
	bold.Println("────────────────────")

	// Find Ubuntu AMI
	fmt.Printf("  ○ Finding Ubuntu AMI...")
	amiID, err := awsinternal.FindUbuntuAMI(ctx, cfg)
	if err != nil {
		red.Printf(" ✗\n")
		yellow.Printf("    Error: %v\n\n", err)
		os.Exit(1)
	}
	green.Printf(" ✓\n")
	fmt.Printf("    %s\n", amiID)

	// Pace API calls to prevent rate limiting
	time.Sleep(1 * time.Second)

	// Network Resources
	fmt.Printf("  ○ Setting up network...")
	vpcID, vpcCIDR, err := awsinternal.GetDefaultVPC(ctx, cfg)
	if err != nil {
		red.Printf(" ✗\n")
		yellow.Printf("    Error: %v\n\n", err)
		os.Exit(1)
	}
	subnetID, subnetAZ, err := awsinternal.GetDefaultSubnet(ctx, cfg, vpcID)
	if err != nil {
		red.Printf(" ✗\n")
		yellow.Printf("    Error: %v\n\n", err)
		os.Exit(1)
	}
	green.Printf(" ✓\n")
	fmt.Printf("    VPC: %s (%s)\n", vpcID, vpcCIDR)
	fmt.Printf("    Subnet: %s (AZ: %s)\n", subnetID, subnetAZ)

	// Pace API calls to prevent rate limiting
	time.Sleep(1 * time.Second)

	// Parallel Resource Creation
	fmt.Printf("  ○ Creating resources in parallel...\n")

	var wg sync.WaitGroup
	var sgID, bucketName string
	var sgErr, iamErr, s3Err error
	sgName := fmt.Sprintf("kl-%s-sg", installationKey)
	roleName := fmt.Sprintf("kl-%s-role", installationKey)
	bucketName = fmt.Sprintf("kl-%s-backups", installationKey)

	startTime := time.Now()

	// Security Group (parallel)
	wg.Add(1)
	go func() {
		defer wg.Done()
		fmt.Printf("    [%s] Starting: Security Group creation\n", time.Now().Format("15:04:05"))
		sgID, sgErr = awsinternal.EnsureSecurityGroup(ctx, cfg, vpcID, vpcCIDR, installationKey)
		if sgErr != nil {
			fmt.Printf("    [%s] Failed: Security Group - %v\n", time.Now().Format("15:04:05"), sgErr)
		} else {
			createdResources.Lock()
			createdResources.sgID = sgID
			createdResources.Unlock()
			fmt.Printf("    [%s] Completed: Security Group\n", time.Now().Format("15:04:05"))
		}
	}()

	// IAM Role (parallel)
	wg.Add(1)
	go func() {
		defer wg.Done()
		fmt.Printf("    [%s] Starting: IAM Role creation\n", time.Now().Format("15:04:05"))
		_, iamErr = awsinternal.EnsureIAMRole(ctx, cfg, installationKey, bucketName)
		if iamErr != nil {
			fmt.Printf("    [%s] Failed: IAM Role - %v\n", time.Now().Format("15:04:05"), iamErr)
		} else {
			createdResources.Lock()
			createdResources.iamCreated = true
			createdResources.Unlock()
			fmt.Printf("    [%s] Completed: IAM Role\n", time.Now().Format("15:04:05"))
		}
	}()

	// S3 Bucket (parallel)
	wg.Add(1)
	go func() {
		defer wg.Done()
		fmt.Printf("    [%s] Starting: S3 Bucket creation\n", time.Now().Format("15:04:05"))
		s3Err = awsinternal.EnsureS3Bucket(ctx, cfg, bucketName, installationKey)
		if s3Err != nil {
			fmt.Printf("    [%s] Failed: S3 Bucket - %v\n", time.Now().Format("15:04:05"), s3Err)
		} else {
			createdResources.Lock()
			createdResources.bucketName = bucketName
			createdResources.Unlock()
			fmt.Printf("    [%s] Completed: S3 Bucket\n", time.Now().Format("15:04:05"))
		}
	}()

	wg.Wait()
	elapsed := time.Since(startTime)
	fmt.Printf("    Parallel operations completed in %.1fs\n", elapsed.Seconds())

	// Check for errors
	if sgErr != nil {
		red.Printf(" ✗\n")
		yellow.Printf("    Security Group Error: %v\n\n", sgErr)
		os.Exit(1)
	}
	if iamErr != nil {
		red.Printf(" ✗\n")
		yellow.Printf("    IAM Role Error: %v\n\n", iamErr)
		os.Exit(1)
	}
	if s3Err != nil {
		red.Printf(" ✗\n")
		yellow.Printf("    S3 Bucket Error: %v\n\n", s3Err)
		os.Exit(1)
	}

	green.Printf(" ✓\n")
	fmt.Printf("    Security Group: %s\n", sgName)
	fmt.Printf("    IAM Role:       %s\n", roleName)
	fmt.Printf("    S3 Bucket:      %s\n", bucketName)

	// Pace API calls to prevent rate limiting (longer delay after parallel operations)
	time.Sleep(2 * time.Second)

	// Instance Profile (depends on IAM role)
	bold.Println("\nFinalizing IAM Setup")
	bold.Println("────────────────────")
	fmt.Printf("  ○ Creating instance profile...")
	err = awsinternal.EnsureInstanceProfile(ctx, cfg, installationKey)
	if err != nil {
		red.Printf(" ✗\n")
		yellow.Printf("    Error: %v\n\n", err)
		os.Exit(1)
	}
	green.Printf(" ✓\n")

	// Pace API calls to prevent rate limiting
	time.Sleep(1 * time.Second)

	// Verify Installation
	bold.Println("\nVerifying Installation")
	bold.Println("──────────────────────")

	fmt.Printf("  ○ Verifying installation key with registration API...")
	secretKey, err := k8sinternal.VerifyInstallation(ctx, installationKey)
	if err != nil {
		red.Printf(" ✗\n")
		yellow.Printf("    Error: %v\n\n", err)
		os.Exit(1)
	}
	green.Printf(" ✓\n")
	fmt.Printf("    Secret key obtained successfully\n")

	// Instance Launch
	bold.Println("\nInstance Deployment")
	bold.Println("───────────────────")

	// Generate K3s agent token
	k3sToken, err := awsinternal.GenerateK3sToken()
	if err != nil {
		red.Printf(" ✗\n")
		yellow.Printf("    Error generating K3s token: %v\n\n", err)
		os.Exit(1)
	}

	fmt.Printf("  ○ Launching EC2 instance (t3.medium)...")
	instanceID, err := awsinternal.LaunchInstance(ctx, cfg, amiID, subnetID, sgID, vpcID, secretKey, bucketName, k3sToken, installationKey, enableTerminationProtection)
	if err != nil {
		red.Printf(" ✗\n")
		yellow.Printf("    Error: %v\n\n", err)
		os.Exit(1)
	}
	createdResources.Lock()
	createdResources.instanceID = instanceID
	createdResources.Unlock()
	green.Printf(" ✓\n")
	fmt.Printf("    %s\n", instanceID)

	fmt.Printf("  ○ Waiting for instance to be ready...")
	publicIP, privateIP, err := awsinternal.WaitForInstance(ctx, cfg, instanceID)
	if err != nil {
		red.Printf(" ✗\n")
		yellow.Printf("    Error: %v\n\n", err)
		os.Exit(1)
	}
	green.Printf(" ✓\n")

	// Success Summary
	fmt.Println()
	green.Println("╭─────────────────────────────────────────╮")
	green.Println("│   ✓ Installation Complete!             │")
	green.Println("╰─────────────────────────────────────────╯")
	fmt.Println()

	bold.Println("Instance Details")
	bold.Println("────────────────")
	fmt.Printf("  Instance ID:    %s\n", instanceID)
	fmt.Printf("  Public IP:      %s\n", publicIP)
	fmt.Printf("  Private IP:     %s\n", privateIP)
	fmt.Printf("  Region:         %s\n", cfg.Region)
	fmt.Printf("  AZ:             %s\n", subnetAZ)

	fmt.Println()
	bold.Println("Instance Access")
	bold.Println("───────────────")
	fmt.Println("  Via AWS Systems Manager:")
	cyan.Printf("    aws ssm start-session --target %s --region %s\n", instanceID, cfg.Region)
	fmt.Println()
}
