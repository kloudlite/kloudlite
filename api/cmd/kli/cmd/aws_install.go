package cmd

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/aws/aws-sdk-go-v2/service/iam"
	iamTypes "github.com/aws/aws-sdk-go-v2/service/iam/types"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	s3Types "github.com/aws/aws-sdk-go-v2/service/s3/types"
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

var region string
var installationKey string
var enableTerminationProtection bool

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
		instanceID  string
		sgID        string
		iamCreated  bool
		bucketName  string
	}

	go func() {
		<-sigChan
		fmt.Println()
		yellow.Println("\n⚠ Installation interrupted! Cleaning up resources...")

		// Load config for cleanup
		cfg, err := loadAWSConfig(context.Background(), region)
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
			deleteS3Bucket(context.Background(), cfg, createdResources.bucketName)
		}

		yellow.Println("Cleanup completed. Exiting...")
		os.Exit(130) // Standard exit code for SIGINT
	}()

	// Configuration
	bold.Println("Configuration")
	bold.Println("─────────────")
	fmt.Printf("  Installation Key: %s\n", installationKey)
	fmt.Printf("  Region:          ")
	cfg, err := loadAWSConfig(ctx, region)
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
	amiID, err := findUbuntuAMI(ctx, cfg)
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
	vpcID, vpcCIDR, err := getDefaultVPC(ctx, cfg)
	if err != nil {
		red.Printf(" ✗\n")
		yellow.Printf("    Error: %v\n\n", err)
		os.Exit(1)
	}
	subnetID, subnetAZ, err := getDefaultSubnet(ctx, cfg, vpcID)
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
		sgID, sgErr = ensureSecurityGroup(ctx, cfg, vpcID, vpcCIDR, installationKey)
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
		_, iamErr = ensureIAMRole(ctx, cfg, installationKey, bucketName)
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
		s3Err = ensureS3Bucket(ctx, cfg, bucketName, installationKey)
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
	err = ensureInstanceProfile(ctx, cfg, installationKey)
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
	secretKey, err := verifyInstallation(ctx, installationKey)
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

	fmt.Printf("  ○ Launching EC2 instance (t3.medium)...")
	instanceID, err := launchInstance(ctx, cfg, amiID, subnetID, sgID, vpcID, secretKey, bucketName, installationKey, enableTerminationProtection)
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
	publicIP, privateIP, err := waitForInstance(ctx, cfg, instanceID)
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

func loadAWSConfig(ctx context.Context, region string) (aws.Config, error) {
	var opts []func(*config.LoadOptions) error
	if region != "" {
		opts = append(opts, config.WithRegion(region))
	}
	return config.LoadDefaultConfig(ctx, opts...)
}

func findUbuntuAMI(ctx context.Context, cfg aws.Config) (string, error) {
	ec2Client := ec2.NewFromConfig(cfg)

	// Search for Ubuntu 24.04 LTS AMD64 AMI
	// Owner: Canonical (099720109477)
	input := &ec2.DescribeImagesInput{
		Owners: []string{"099720109477"}, // Canonical
		Filters: []types.Filter{
			{
				Name:   aws.String("name"),
				Values: []string{"ubuntu/images/hvm-ssd-gp3/ubuntu-noble-24.04-amd64-server-*"},
			},
			{
				Name:   aws.String("architecture"),
				Values: []string{"x86_64"},
			},
			{
				Name:   aws.String("root-device-type"),
				Values: []string{"ebs"},
			},
			{
				Name:   aws.String("virtualization-type"),
				Values: []string{"hvm"},
			},
			{
				Name:   aws.String("state"),
				Values: []string{"available"},
			},
		},
	}

	result, err := ec2Client.DescribeImages(ctx, input)
	if err != nil {
		return "", fmt.Errorf("failed to describe images: %w", err)
	}

	if len(result.Images) == 0 {
		return "", fmt.Errorf("no Ubuntu 24.04 LTS AMD64 AMI found in region %s", cfg.Region)
	}

	// Find the most recent AMI
	var newestAMI *types.Image
	for i := range result.Images {
		if newestAMI == nil || *result.Images[i].CreationDate > *newestAMI.CreationDate {
			newestAMI = &result.Images[i]
		}
	}

	return *newestAMI.ImageId, nil
}

func ensureIAMRole(ctx context.Context, cfg aws.Config, installationKey, bucketName string) (string, error) {
	iamClient := iam.NewFromConfig(cfg)
	roleName := fmt.Sprintf("kl-%s-role", installationKey)

	// Check if role exists
	getResult, err := iamClient.GetRole(ctx, &iam.GetRoleInput{
		RoleName: aws.String(roleName),
	})
	if err == nil {
		// Role exists, return ARN
		return *getResult.Role.Arn, nil
	}

	// Create trust policy for EC2
	trustPolicy := map[string]interface{}{
		"Version": "2012-10-17",
		"Statement": []map[string]interface{}{
			{
				"Effect": "Allow",
				"Principal": map[string]string{
					"Service": "ec2.amazonaws.com",
				},
				"Action": "sts:AssumeRole",
			},
		},
	}
	trustPolicyJSON, err := json.Marshal(trustPolicy)
	if err != nil {
		return "", fmt.Errorf("failed to marshal trust policy: %w", err)
	}

	// Create role
	createResult, err := iamClient.CreateRole(ctx, &iam.CreateRoleInput{
		RoleName:                 aws.String(roleName),
		AssumeRolePolicyDocument: aws.String(string(trustPolicyJSON)),
		Description:              aws.String("Kloudlite runtime IAM role for EC2 instances"),
		Tags: []iamTypes.Tag{
			{Key: aws.String("Name"), Value: aws.String(roleName)},
			{Key: aws.String("ManagedBy"), Value: aws.String("kloudlite")},
			{Key: aws.String("Project"), Value: aws.String("kloudlite")},
			{Key: aws.String("Purpose"), Value: aws.String("kloudlite-installation")},
			{Key: aws.String("InstallationKey"), Value: aws.String(installationKey)},
		},
	})
	if err != nil {
		return "", fmt.Errorf("failed to create IAM role: %w", err)
	}

	// Construct S3 ARN using bucket name
	bucketArn := fmt.Sprintf("arn:aws:s3:::%s", bucketName)
	bucketObjectArn := fmt.Sprintf("arn:aws:s3:::%s/*", bucketName)

	// Create inline policy with required permissions
	policy := map[string]interface{}{
		"Version": "2012-10-17",
		"Statement": []map[string]interface{}{
			{
				"Effect": "Allow",
				"Action": []string{
					"ec2:RunInstances",
					"ec2:TerminateInstances",
					"ec2:DescribeInstances",
					"ec2:ModifyInstanceAttribute",
					"ec2:DescribeInstanceTypes",
					"ec2:DescribeImages",
					"ec2:DescribeVolumes",
					"ec2:CreateTags",
				},
				"Resource": "*",
			},
			{
				"Effect": "Allow",
				"Action": []string{
					"s3:ListBucket",
					"s3:GetBucketLocation",
				},
				"Resource": bucketArn,
			},
			{
				"Effect": "Allow",
				"Action": []string{
					"s3:PutObject",
					"s3:GetObject",
					"s3:DeleteObject",
				},
				"Resource": bucketObjectArn,
			},
		},
	}
	policyJSON, err := json.Marshal(policy)
	if err != nil {
		return "", fmt.Errorf("failed to marshal policy: %w", err)
	}

	_, err = iamClient.PutRolePolicy(ctx, &iam.PutRolePolicyInput{
		RoleName:       aws.String(roleName),
		PolicyName:     aws.String("kl-ec2-policy"),
		PolicyDocument: aws.String(string(policyJSON)),
	})
	if err != nil {
		return "", fmt.Errorf("failed to put role policy: %w", err)
	}

	// Attach AWS managed policy for SSM
	_, err = iamClient.AttachRolePolicy(ctx, &iam.AttachRolePolicyInput{
		RoleName:  aws.String(roleName),
		PolicyArn: aws.String("arn:aws:iam::aws:policy/AmazonSSMManagedInstanceCore"),
	})
	if err != nil {
		return "", fmt.Errorf("failed to attach SSM policy: %w", err)
	}

	return *createResult.Role.Arn, nil
}

func ensureInstanceProfile(ctx context.Context, cfg aws.Config, installationKey string) error {
	iamClient := iam.NewFromConfig(cfg)
	profileName := fmt.Sprintf("kl-%s-role", installationKey)
	roleName := fmt.Sprintf("kl-%s-role", installationKey)

	// Check if instance profile exists
	_, err := iamClient.GetInstanceProfile(ctx, &iam.GetInstanceProfileInput{
		InstanceProfileName: aws.String(profileName),
	})
	if err == nil {
		// Instance profile exists
		return nil
	}

	// Create instance profile
	_, err = iamClient.CreateInstanceProfile(ctx, &iam.CreateInstanceProfileInput{
		InstanceProfileName: aws.String(profileName),
		Tags: []iamTypes.Tag{
			{Key: aws.String("Name"), Value: aws.String(profileName)},
			{Key: aws.String("ManagedBy"), Value: aws.String("kloudlite")},
			{Key: aws.String("Project"), Value: aws.String("kloudlite")},
			{Key: aws.String("Purpose"), Value: aws.String("kloudlite-installation")},
			{Key: aws.String("InstallationKey"), Value: aws.String(installationKey)},
		},
	})
	if err != nil {
		return fmt.Errorf("failed to create instance profile: %w", err)
	}

	// Add role to instance profile
	_, err = iamClient.AddRoleToInstanceProfile(ctx, &iam.AddRoleToInstanceProfileInput{
		InstanceProfileName: aws.String(profileName),
		RoleName:            aws.String(roleName),
	})
	if err != nil {
		return fmt.Errorf("failed to add role to instance profile: %w", err)
	}

	// Wait a bit for IAM to propagate
	time.Sleep(10 * time.Second)

	return nil
}

func getDefaultVPC(ctx context.Context, cfg aws.Config) (string, string, error) {
	ec2Client := ec2.NewFromConfig(cfg)

	result, err := ec2Client.DescribeVpcs(ctx, &ec2.DescribeVpcsInput{
		Filters: []types.Filter{
			{
				Name:   aws.String("isDefault"),
				Values: []string{"true"},
			},
		},
	})
	if err != nil {
		return "", "", fmt.Errorf("failed to describe VPCs: %w", err)
	}

	if len(result.Vpcs) == 0 {
		return "", "", fmt.Errorf("no default VPC found in region %s", cfg.Region)
	}

	vpc := result.Vpcs[0]
	return *vpc.VpcId, *vpc.CidrBlock, nil
}

func getDefaultSubnet(ctx context.Context, cfg aws.Config, vpcID string) (string, string, error) {
	ec2Client := ec2.NewFromConfig(cfg)

	result, err := ec2Client.DescribeSubnets(ctx, &ec2.DescribeSubnetsInput{
		Filters: []types.Filter{
			{
				Name:   aws.String("vpc-id"),
				Values: []string{vpcID},
			},
			{
				Name:   aws.String("default-for-az"),
				Values: []string{"true"},
			},
		},
	})
	if err != nil {
		return "", "", fmt.Errorf("failed to describe subnets: %w", err)
	}

	if len(result.Subnets) == 0 {
		return "", "", fmt.Errorf("no default subnet found in VPC %s", vpcID)
	}

	// Prefer subnets NOT in us-east-1e (which often doesn't support t3 instances)
	// Try to find a subnet in a different AZ first
	for _, subnet := range result.Subnets {
		if subnet.AvailabilityZone != nil && *subnet.AvailabilityZone != "us-east-1e" {
			az := ""
			if subnet.AvailabilityZone != nil {
				az = *subnet.AvailabilityZone
			}
			return *subnet.SubnetId, az, nil
		}
	}

	// If all subnets are in us-east-1e or we couldn't find a better one, use the first
	az := ""
	if result.Subnets[0].AvailabilityZone != nil {
		az = *result.Subnets[0].AvailabilityZone
	}
	return *result.Subnets[0].SubnetId, az, nil
}

func ensureSecurityGroup(ctx context.Context, cfg aws.Config, vpcID, vpcCIDR string, installationKey string) (string, error) {
	ec2Client := ec2.NewFromConfig(cfg)
	sgName := fmt.Sprintf("kl-%s-sg", installationKey)

	// Check if security group exists
	descResult, err := ec2Client.DescribeSecurityGroups(ctx, &ec2.DescribeSecurityGroupsInput{
		Filters: []types.Filter{
			{
				Name:   aws.String("group-name"),
				Values: []string{sgName},
			},
			{
				Name:   aws.String("vpc-id"),
				Values: []string{vpcID},
			},
		},
	})
	if err == nil && len(descResult.SecurityGroups) > 0 {
		// Security group exists
		return *descResult.SecurityGroups[0].GroupId, nil
	}

	// Create security group
	createResult, err := ec2Client.CreateSecurityGroup(ctx, &ec2.CreateSecurityGroupInput{
		GroupName:   aws.String(sgName),
		Description: aws.String("Kloudlite security group"),
		VpcId:       aws.String(vpcID),
		TagSpecifications: []types.TagSpecification{
			{
				ResourceType: types.ResourceTypeSecurityGroup,
				Tags: []types.Tag{
					{Key: aws.String("Name"), Value: aws.String(sgName)},
					{Key: aws.String("ManagedBy"), Value: aws.String("kloudlite")},
					{Key: aws.String("Project"), Value: aws.String("kloudlite")},
					{Key: aws.String("Purpose"), Value: aws.String("kloudlite-installation")},
					{Key: aws.String("InstallationKey"), Value: aws.String(installationKey)},
				},
			},
		},
	})
	if err != nil {
		return "", fmt.Errorf("failed to create security group: %w", err)
	}

	sgID := *createResult.GroupId

	// Add ingress rules
	// Port 80 and 443 from anywhere for web access
	_, err = ec2Client.AuthorizeSecurityGroupIngress(ctx, &ec2.AuthorizeSecurityGroupIngressInput{
		GroupId: aws.String(sgID),
		IpPermissions: []types.IpPermission{
			{
				IpProtocol: aws.String("tcp"),
				FromPort:   aws.Int32(80),
				ToPort:     aws.Int32(80),
				IpRanges: []types.IpRange{
					{CidrIp: aws.String("0.0.0.0/0")},
				},
			},
			{
				IpProtocol: aws.String("tcp"),
				FromPort:   aws.Int32(443),
				ToPort:     aws.Int32(443),
				IpRanges: []types.IpRange{
					{CidrIp: aws.String("0.0.0.0/0")},
				},
			},
			// Internal ports from VPC CIDR
			{
				IpProtocol: aws.String("tcp"),
				FromPort:   aws.Int32(6443),
				ToPort:     aws.Int32(6443),
				IpRanges: []types.IpRange{
					{CidrIp: aws.String(vpcCIDR)},
				},
			},
			{
				IpProtocol: aws.String("udp"),
				FromPort:   aws.Int32(8472),
				ToPort:     aws.Int32(8472),
				IpRanges: []types.IpRange{
					{CidrIp: aws.String(vpcCIDR)},
				},
			},
			{
				IpProtocol: aws.String("tcp"),
				FromPort:   aws.Int32(10250),
				ToPort:     aws.Int32(10250),
				IpRanges: []types.IpRange{
					{CidrIp: aws.String(vpcCIDR)},
				},
			},
			{
				IpProtocol: aws.String("tcp"),
				FromPort:   aws.Int32(5001),
				ToPort:     aws.Int32(5001),
				IpRanges: []types.IpRange{
					{CidrIp: aws.String(vpcCIDR)},
				},
			},
		},
	})
	if err != nil {
		return "", fmt.Errorf("failed to authorize security group ingress: %w", err)
	}

	return sgID, nil
}

type verifyInstallationRequest struct {
	InstallationKey string `json:"installationKey"`
}

type verifyInstallationResponse struct {
	Success    bool   `json:"success"`
	SecretKey  string `json:"secretKey"`
	Subdomain  string `json:"subdomain,omitempty"`
	Error      string `json:"error,omitempty"`
}

func verifyInstallation(ctx context.Context, installationKey string) (string, error) {
	// TODO: Make this configurable via flag or environment variable
	registrationAPIURL := "https://console.kloudlite.io/api/installations/verify-key"

	// Create request payload
	reqPayload := verifyInstallationRequest{
		InstallationKey: installationKey,
	}
	reqBody, err := json.Marshal(reqPayload)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	// Create HTTP request with timeout
	httpClient := &http.Client{
		Timeout: 30 * time.Second,
	}

	req, err := http.NewRequestWithContext(ctx, "POST", registrationAPIURL, bytes.NewReader(reqBody))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	// Execute request
	resp, err := httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	// Read response
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	// Check HTTP status
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("API returned status %d: %s", resp.StatusCode, string(respBody))
	}

	// Parse response
	var verifyResp verifyInstallationResponse
	if err := json.Unmarshal(respBody, &verifyResp); err != nil {
		return "", fmt.Errorf("failed to parse response: %w", err)
	}

	// Check for API error
	if verifyResp.Error != "" {
		return "", fmt.Errorf("API error: %s", verifyResp.Error)
	}

	// Validate secret key
	if verifyResp.SecretKey == "" {
		return "", fmt.Errorf("no secret key returned from API")
	}

	return verifyResp.SecretKey, nil
}

func launchInstance(ctx context.Context, cfg aws.Config, amiID, subnetID, sgID, vpcID, secretKey, bucketName string, installationKey string, enableProtection bool) (string, error) {
	ec2Client := ec2.NewFromConfig(cfg)
	instanceName := fmt.Sprintf("kl-%s-instance", installationKey)
	profileName := fmt.Sprintf("kl-%s-role", installationKey)
	region := cfg.Region

	// Create cloud-init script to install K3s on startup
	userData := fmt.Sprintf(`#!/bin/bash
set -euo pipefail

# Log output to file
exec > >(tee -a /var/log/kloudlite-init.log)
exec 2>&1

echo "Starting Kloudlite installation at $(date)"

# Update system
apt-get update -y
apt-get upgrade -y

# Install required packages
apt-get install -y curl wget git

# Install K3s server
echo "Installing K3s server..."
K3S_VERSION="v1.31.1+k3s1"
curl -sfL https://get.k3s.io | INSTALL_K3S_VERSION="$K3S_VERSION" sh -s - server \
  --disable traefik \
  --write-kubeconfig-mode 644

# Wait for K3s to be ready
echo "Waiting for K3s to be ready..."
until kubectl get nodes 2>/dev/null; do
  sleep 2
done

echo "K3s installation completed at $(date)"

# Extract K3s token and server URL
echo "Extracting K3s configuration..."
K3S_AGENT_TOKEN=$(cat /var/lib/rancher/k3s/server/node-token)
K3S_SERVER_URL="https://127.0.0.1:6443"

# Fetch instance IPs from EC2 metadata service
echo "Fetching instance metadata..."
TOKEN=$(curl -X PUT "http://169.254.169.254/latest/api/token" -H "X-aws-ec2-metadata-token-ttl-seconds: 21600" 2>/dev/null)
PUBLIC_IP=$(curl -H "X-aws-ec2-metadata-token: $TOKEN" http://169.254.169.254/latest/meta-data/public-ipv4 2>/dev/null || echo "")
PRIVATE_IP=$(curl -H "X-aws-ec2-metadata-token: $TOKEN" http://169.254.169.254/latest/meta-data/local-ipv4 2>/dev/null || echo "")

echo "Instance IPs - Public: $PUBLIC_IP, Private: $PRIVATE_IP"
echo "K3s Version: $K3S_VERSION"
echo "K3s Server URL: $K3S_SERVER_URL"

# Install Kloudlite components
echo "Installing Kloudlite API Server and Frontend..."

# Create namespace
kubectl create namespace kloudlite || true

# Create templated API Server manifest with environment variables
echo "Creating API Server manifest..."
cat <<'EOF' | kubectl apply -f -
apiVersion: v1
kind: Service
metadata:
  name: api-server
  namespace: kloudlite
spec:
  selector:
    app: api-server
  ports:
    - name: http
      port: 8080
      targetPort: 8080
  clusterIP: None
---
apiVersion: v1
kind: Service
metadata:
  name: api-server-lb
  namespace: kloudlite
spec:
  type: LoadBalancer
  selector:
    app: api-server
  ports:
    - name: http
      port: 80
      targetPort: 8080
---
apiVersion: apps/v1
kind: StatefulSet
metadata:
  name: api-server
  namespace: kloudlite
spec:
  serviceName: api-server
  replicas: 1
  selector:
    matchLabels:
      app: api-server
  template:
    metadata:
      labels:
        app: api-server
    spec:
      containers:
        - name: api-server
          image: ghcr.io/kloudlite/kloudlite/api-server:latest
          ports:
            - containerPort: 8080
              name: http
          env:
            - name: PORT
              value: "8080"
            - name: INSTALLATION_ID
              value: "%s"
            - name: INSTALLATION_SECRET
              value: "%s"
            - name: AWS_VPC_ID
              value: "%s"
            - name: AWS_SECURITY_GROUP_ID
              value: "%s"
            - name: AWS_REGION
              value: "%s"
            - name: AWS_AMI_ID
              value: "%s"
            - name: AWS_PRIVATE_IP
              value: "$PRIVATE_IP"
            - name: K3S_VERSION
              value: "$K3S_VERSION"
            - name: K3S_AGENT_TOKEN
              value: "$K3S_AGENT_TOKEN"
            - name: K3S_SERVER_URL
              value: "$K3S_SERVER_URL"
          resources:
            requests:
              memory: "256Mi"
              cpu: "100m"
            limits:
              memory: "512Mi"
              cpu: "500m"
          livenessProbe:
            httpGet:
              path: /health
              port: 8080
            initialDelaySeconds: 30
            periodSeconds: 10
          readinessProbe:
            httpGet:
              path: /health
              port: 8080
            initialDelaySeconds: 5
            periodSeconds: 5
  volumeClaimTemplates:
    - metadata:
        name: data
      spec:
        accessModes: ["ReadWriteOnce"]
        resources:
          requests:
            storage: 10Gi
EOF

# Wait for API Server to be ready
echo "Waiting for API Server to be ready..."
kubectl wait --for=condition=ready pod -l app=api-server -n kloudlite --timeout=300s || true

# Apply Frontend Deployment
echo "Deploying Frontend..."
MANIFEST_BASE_URL="https://raw.githubusercontent.com/kloudlite/kloudlite/master/api/manifests/install"
kubectl apply -f ${MANIFEST_BASE_URL}/frontend.yaml

# Wait for Frontend to be ready
echo "Waiting for Frontend to be ready..."
kubectl wait --for=condition=ready pod -l app=frontend -n kloudlite --timeout=300s || true

echo "Getting service endpoints..."
kubectl get svc -n kloudlite

# Setup K3s SQLite backup to S3
echo "Setting up K3s backup CronJob..."
cat <<'BACKUP_EOF' | kubectl apply -f -
apiVersion: v1
kind: ServiceAccount
metadata:
  name: k3s-backup
  namespace: kloudlite
---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: k3s-backup
rules:
  - apiGroups: [""]
    resources: ["nodes"]
    verbs: ["get", "list"]
---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  name: k3s-backup
subjects:
  - kind: ServiceAccount
    name: k3s-backup
    namespace: kloudlite
roleRef:
  kind: ClusterRole
  name: k3s-backup
  apiGroup: rbac.authorization.k8s.io
---
apiVersion: batch/v1
kind: CronJob
metadata:
  name: k3s-backup
  namespace: kloudlite
spec:
  schedule: "*/30 * * * *"  # Every 30 minutes
  successfulJobsHistoryLimit: 3
  failedJobsHistoryLimit: 3
  jobTemplate:
    spec:
      template:
        metadata:
          labels:
            app: k3s-backup
        spec:
          serviceAccountName: k3s-backup
          hostNetwork: true
          hostPID: true
          restartPolicy: OnFailure
          containers:
            - name: backup
              image: amazon/aws-cli:latest
              command:
                - /bin/bash
                - -c
                - |
                  set -euo pipefail

                  echo "Starting K3s backup at $(date)"

                  # K3s database location
                  DB_PATH="/var/lib/rancher/k3s/server/db/state.db"
                  BACKUP_FILE="/tmp/k3s-backup-$(date +%%Y%%m%%d-%%H%%M%%S).db"
                  S3_BUCKET="%s"
                  S3_KEY="backups/k3s-backup-$(date +%%Y%%m%%d-%%H%%M%%S).db"

                  # Check if database exists
                  if [ ! -f "$DB_PATH" ]; then
                    echo "ERROR: K3s database not found at $DB_PATH"
                    exit 1
                  fi

                  # Copy database (SQLite backup)
                  echo "Backing up database..."
                  cp "$DB_PATH" "$BACKUP_FILE"

                  # Compress backup
                  echo "Compressing backup..."
                  gzip "$BACKUP_FILE"
                  BACKUP_FILE="${BACKUP_FILE}.gz"

                  # Upload to S3
                  echo "Uploading to S3: s3://${S3_BUCKET}/${S3_KEY}.gz"
                  aws s3 cp "$BACKUP_FILE" "s3://${S3_BUCKET}/${S3_KEY}.gz" --region %s

                  # Cleanup local backup
                  rm -f "$BACKUP_FILE"

                  echo "Backup completed successfully at $(date)"
              env:
                - name: AWS_REGION
                  value: "%s"
              volumeMounts:
                - name: k3s-data
                  mountPath: /var/lib/rancher/k3s
                  readOnly: true
              resources:
                requests:
                  memory: "128Mi"
                  cpu: "100m"
                limits:
                  memory: "256Mi"
                  cpu: "200m"
          volumes:
            - name: k3s-data
              hostPath:
                path: /var/lib/rancher/k3s
                type: Directory
BACKUP_EOF

echo "K3s backup CronJob created successfully"

echo "Kloudlite installation completed successfully at $(date)!"
`, installationKey, secretKey, vpcID, sgID, region, amiID, bucketName, region, region)

	// Base64 encode the user data
	userDataEncoded := base64.StdEncoding.EncodeToString([]byte(userData))

	result, err := ec2Client.RunInstances(ctx, &ec2.RunInstancesInput{
		ImageId:      aws.String(amiID),
		InstanceType: types.InstanceTypeT3Medium,
		MinCount:     aws.Int32(1),
		MaxCount:     aws.Int32(1),
		UserData:     aws.String(userDataEncoded),
		NetworkInterfaces: []types.InstanceNetworkInterfaceSpecification{
			{
				DeviceIndex:              aws.Int32(0),
				SubnetId:                 aws.String(subnetID),
				Groups:                   []string{sgID},
				AssociatePublicIpAddress: aws.Bool(true),
			},
		},
		BlockDeviceMappings: []types.BlockDeviceMapping{
			{
				DeviceName: aws.String("/dev/sda1"),
				Ebs: &types.EbsBlockDevice{
					VolumeSize:          aws.Int32(100),
					VolumeType:          types.VolumeTypeGp3,
					DeleteOnTermination: aws.Bool(true),
				},
			},
		},
		IamInstanceProfile: &types.IamInstanceProfileSpecification{
			Name: aws.String(profileName),
		},
		TagSpecifications: []types.TagSpecification{
			{
				ResourceType: types.ResourceTypeInstance,
				Tags: []types.Tag{
					{Key: aws.String("Name"), Value: aws.String(instanceName)},
					{Key: aws.String("ManagedBy"), Value: aws.String("kloudlite")},
					{Key: aws.String("Project"), Value: aws.String("kloudlite")},
					{Key: aws.String("Purpose"), Value: aws.String("kloudlite-installation")},
					{Key: aws.String("InstallationKey"), Value: aws.String(installationKey)},
				},
			},
			{
				ResourceType: types.ResourceTypeVolume,
				Tags: []types.Tag{
					{Key: aws.String("Name"), Value: aws.String(fmt.Sprintf("%s-volume", instanceName))},
					{Key: aws.String("ManagedBy"), Value: aws.String("kloudlite")},
					{Key: aws.String("Project"), Value: aws.String("kloudlite")},
					{Key: aws.String("Purpose"), Value: aws.String("kloudlite-installation")},
					{Key: aws.String("InstallationKey"), Value: aws.String(installationKey)},
				},
			},
		},
	})
	if err != nil {
		return "", fmt.Errorf("failed to run instance: %w", err)
	}

	instanceID := *result.Instances[0].InstanceId

	// Enable termination protection if requested
	if enableProtection {
		_, err = ec2Client.ModifyInstanceAttribute(ctx, &ec2.ModifyInstanceAttributeInput{
			InstanceId: aws.String(instanceID),
			DisableApiTermination: &types.AttributeBooleanValue{
				Value: aws.Bool(true),
			},
		})
		if err != nil {
			return instanceID, fmt.Errorf("failed to enable termination protection: %w", err)
		}
	}

	return instanceID, nil
}

func waitForInstance(ctx context.Context, cfg aws.Config, instanceID string) (string, string, error) {
	ec2Client := ec2.NewFromConfig(cfg)

	waiter := ec2.NewInstanceRunningWaiter(ec2Client)
	err := waiter.Wait(ctx, &ec2.DescribeInstancesInput{
		InstanceIds: []string{instanceID},
	}, 5*time.Minute)
	if err != nil {
		return "", "", fmt.Errorf("failed waiting for instance to be running: %w", err)
	}

	// Get instance details
	result, err := ec2Client.DescribeInstances(ctx, &ec2.DescribeInstancesInput{
		InstanceIds: []string{instanceID},
	})
	if err != nil {
		return "", "", fmt.Errorf("failed to describe instance: %w", err)
	}

	if len(result.Reservations) == 0 || len(result.Reservations[0].Instances) == 0 {
		return "", "", fmt.Errorf("instance not found")
	}

	instance := result.Reservations[0].Instances[0]
	publicIP := ""
	privateIP := ""

	if instance.PublicIpAddress != nil {
		publicIP = *instance.PublicIpAddress
	}
	if instance.PrivateIpAddress != nil {
		privateIP = *instance.PrivateIpAddress
	}

	return publicIP, privateIP, nil
}

func ensureS3Bucket(ctx context.Context, cfg aws.Config, bucketName, installationKey string) error {
	s3Client := s3.NewFromConfig(cfg)

	// Check if bucket exists
	_, err := s3Client.HeadBucket(ctx, &s3.HeadBucketInput{
		Bucket: aws.String(bucketName),
	})
	if err == nil {
		// Bucket exists
		return nil
	}

	// Create bucket
	createInput := &s3.CreateBucketInput{
		Bucket: aws.String(bucketName),
	}

	// For regions other than us-east-1, we need to specify LocationConstraint
	if cfg.Region != "us-east-1" {
		createInput.CreateBucketConfiguration = &s3Types.CreateBucketConfiguration{
			LocationConstraint: s3Types.BucketLocationConstraint(cfg.Region),
		}
	}

	_, err = s3Client.CreateBucket(ctx, createInput)
	if err != nil {
		return fmt.Errorf("failed to create S3 bucket: %w", err)
	}

	// Enable versioning for backup safety
	_, err = s3Client.PutBucketVersioning(ctx, &s3.PutBucketVersioningInput{
		Bucket: aws.String(bucketName),
		VersioningConfiguration: &s3Types.VersioningConfiguration{
			Status: s3Types.BucketVersioningStatusEnabled,
		},
	})
	if err != nil {
		return fmt.Errorf("failed to enable bucket versioning: %w", err)
	}

	// Add lifecycle policy to expire old backups after 30 days
	_, err = s3Client.PutBucketLifecycleConfiguration(ctx, &s3.PutBucketLifecycleConfigurationInput{
		Bucket: aws.String(bucketName),
		LifecycleConfiguration: &s3Types.BucketLifecycleConfiguration{
			Rules: []s3Types.LifecycleRule{
				{
					ID:     aws.String("expire-old-backups"),
					Status: s3Types.ExpirationStatusEnabled,
					Expiration: &s3Types.LifecycleExpiration{
						Days: aws.Int32(30),
					},
					Filter: &s3Types.LifecycleRuleFilter{
						Prefix: aws.String(""),
					},
				},
			},
		},
	})
	if err != nil {
		return fmt.Errorf("failed to set lifecycle policy: %w", err)
	}

	// Add tags
	_, err = s3Client.PutBucketTagging(ctx, &s3.PutBucketTaggingInput{
		Bucket: aws.String(bucketName),
		Tagging: &s3Types.Tagging{
			TagSet: []s3Types.Tag{
				{Key: aws.String("Name"), Value: aws.String(bucketName)},
				{Key: aws.String("ManagedBy"), Value: aws.String("kloudlite")},
				{Key: aws.String("Project"), Value: aws.String("kloudlite")},
				{Key: aws.String("Purpose"), Value: aws.String("k3s-backups")},
				{Key: aws.String("InstallationKey"), Value: aws.String(installationKey)},
			},
		},
	})
	if err != nil {
		return fmt.Errorf("failed to tag bucket: %w", err)
	}

	return nil
}

func deleteS3Bucket(ctx context.Context, cfg aws.Config, bucketName string) error {
	s3Client := s3.NewFromConfig(cfg)

	// List and delete all objects (including versions)
	listInput := &s3.ListObjectVersionsInput{
		Bucket: aws.String(bucketName),
	}

	for {
		listOutput, err := s3Client.ListObjectVersions(ctx, listInput)
		if err != nil {
			return fmt.Errorf("failed to list objects: %w", err)
		}

		// Delete versions
		if len(listOutput.Versions) > 0 {
			var objects []s3Types.ObjectIdentifier
			for _, version := range listOutput.Versions {
				objects = append(objects, s3Types.ObjectIdentifier{
					Key:       version.Key,
					VersionId: version.VersionId,
				})
			}

			_, err = s3Client.DeleteObjects(ctx, &s3.DeleteObjectsInput{
				Bucket: aws.String(bucketName),
				Delete: &s3Types.Delete{
					Objects: objects,
				},
			})
			if err != nil {
				return fmt.Errorf("failed to delete object versions: %w", err)
			}
		}

		// Delete delete markers
		if len(listOutput.DeleteMarkers) > 0 {
			var objects []s3Types.ObjectIdentifier
			for _, marker := range listOutput.DeleteMarkers {
				objects = append(objects, s3Types.ObjectIdentifier{
					Key:       marker.Key,
					VersionId: marker.VersionId,
				})
			}

			_, err = s3Client.DeleteObjects(ctx, &s3.DeleteObjectsInput{
				Bucket: aws.String(bucketName),
				Delete: &s3Types.Delete{
					Objects: objects,
				},
			})
			if err != nil {
				return fmt.Errorf("failed to delete markers: %w", err)
			}
		}

		// Check if there are more objects
		if !aws.ToBool(listOutput.IsTruncated) {
			break
		}
		listInput.KeyMarker = listOutput.NextKeyMarker
		listInput.VersionIdMarker = listOutput.NextVersionIdMarker
	}

	// Delete the bucket
	_, err := s3Client.DeleteBucket(ctx, &s3.DeleteBucketInput{
		Bucket: aws.String(bucketName),
	})
	if err != nil {
		return fmt.Errorf("failed to delete bucket: %w", err)
	}

	return nil
}
