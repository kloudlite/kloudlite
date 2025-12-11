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
	"github.com/kloudlite/kloudlite/api/cmd/kli/internal/console"
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
  - Create IAM role 'kl-{installation-key}-role' with required permissions (including S3, ELB, ACM)
  - Create S3 bucket 'kl-{installation-key}-backups' for K3s database backups
  - Create security groups for EC2 and ALB
  - Launch t3.medium EC2 instance with 100GB storage
  - Configure instance in default VPC with public IP
  - Setup automated K3s SQLite backup to S3 every 30 minutes
  - Create Application Load Balancer with TLS termination
  - Request ACM certificate with DNS validation via Cloudflare
  - Configure custom domain using the subdomain reserved in console

NOTE: The subdomain must be reserved in the console (console.kloudlite.io)
before running this command. The installation will fail if no subdomain
has been configured for the installation key.`,
	Example: `  # Install using default AWS region from config
  kli aws install --installation-key prod

  # Install in a specific region
  kli aws install --installation-key staging --region us-west-2

  # Install without ALB (direct EC2 access only)
  kli aws install --installation-key dev --skip-alb`,
	Run: runAWSInstall,
}

var (
	region                      string
	installationKey             string
	enableTerminationProtection bool
	skipALB                     bool
)

func init() {
	awsInstallCmd.Flags().StringVar(&region, "region", "", "AWS region (uses default from AWS config if not specified)")
	awsInstallCmd.Flags().StringVar(&installationKey, "installation-key", "", "Installation key to identify this installation (required)")
	awsInstallCmd.Flags().BoolVar(&enableTerminationProtection, "enable-termination-protection", true, "Enable EC2 termination protection (default: true)")
	awsInstallCmd.Flags().BoolVar(&skipALB, "skip-alb", false, "Skip ALB and TLS setup (direct EC2 access only)")
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
	cyan.Println("+-----------------------------------------+")
	cyan.Println("|   Kloudlite AWS Installation            |")
	cyan.Println("+-----------------------------------------+")
	fmt.Println()

	ctx := context.Background()

	// Setup signal handling for cleanup on interruption
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	var createdResources struct {
		sync.Mutex
		instanceID string
		sgID       string
		masterSgID string
		albSgID    string
		iamCreated bool
		bucketName string
		albARN     string
		tgARN      string
		certARN    string
		vpcID      string
	}

	go func() {
		<-sigChan
		fmt.Println()
		yellow.Println("\nInstallation interrupted! Cleaning up resources...")

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
		if createdResources.albARN != "" {
			fmt.Printf("  Deleting ALB...\n")
			awsinternal.DeleteALB(context.Background(), cfg, installationKey)
		}
		if createdResources.tgARN != "" {
			fmt.Printf("  Deleting Target Group...\n")
			awsinternal.DeleteTargetGroup(context.Background(), cfg, installationKey)
		}
		if createdResources.certARN != "" {
			fmt.Printf("  Deleting ACM Certificate...\n")
			awsinternal.DeleteCertificate(context.Background(), cfg, createdResources.certARN)
		}
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
		if createdResources.masterSgID != "" && createdResources.vpcID != "" {
			fmt.Printf("  Deleting master security group...\n")
			awsinternal.DeleteSecurityGroupByName(context.Background(), cfg, createdResources.vpcID, fmt.Sprintf("kl-%s-master-sg", installationKey))
		}
		if createdResources.albSgID != "" && createdResources.vpcID != "" {
			fmt.Printf("  Deleting ALB security group...\n")
			awsinternal.DeleteSecurityGroupByName(context.Background(), cfg, createdResources.vpcID, fmt.Sprintf("kl-%s-alb-sg", installationKey))
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
	bold.Println("-------------")
	fmt.Printf("  Installation Key: %s\n", installationKey)
	fmt.Printf("  Region:          ")
	cfg, err := awsinternal.LoadAWSConfig(ctx, region)
	if err != nil {
		red.Printf("x\n")
		yellow.Printf("  Error: %v\n\n", err)
		os.Exit(1)
	}
	green.Printf("+ %s\n", cfg.Region)
	fmt.Println()

	// Console API client
	consoleClient := console.NewClient()

	// Verify Installation and get subdomain
	bold.Println("Verifying Installation")
	bold.Println("----------------------")

	fmt.Printf("  o Verifying installation key with registration API...")
	verifyResult, err := k8sinternal.VerifyInstallation(ctx, installationKey)
	if err != nil {
		red.Printf(" x\n")
		yellow.Printf("    Error: %v\n\n", err)
		os.Exit(1)
	}
	green.Printf(" +\n")
	fmt.Printf("    Secret key obtained successfully\n")

	secretKey := verifyResult.SecretKey
	var fullDomain string

	// Check if subdomain was configured in console (required for ALB)
	if !skipALB {
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

	// Find Ubuntu AMI
	fmt.Printf("  o Finding Ubuntu AMI...")
	amiID, err := awsinternal.FindUbuntuAMI(ctx, cfg)
	if err != nil {
		red.Printf(" x\n")
		yellow.Printf("    Error: %v\n\n", err)
		os.Exit(1)
	}
	green.Printf(" +\n")
	fmt.Printf("    %s\n", amiID)

	// Pace API calls to prevent rate limiting
	time.Sleep(1 * time.Second)

	// Network Resources
	fmt.Printf("  o Setting up network...")
	vpcID, vpcCIDR, err := awsinternal.GetDefaultVPC(ctx, cfg)
	if err != nil {
		red.Printf(" x\n")
		yellow.Printf("    Error: %v\n\n", err)
		os.Exit(1)
	}

	createdResources.Lock()
	createdResources.vpcID = vpcID
	createdResources.Unlock()

	subnetID, subnetAZ, err := awsinternal.GetDefaultSubnet(ctx, cfg, vpcID)
	if err != nil {
		red.Printf(" x\n")
		yellow.Printf("    Error: %v\n\n", err)
		os.Exit(1)
	}

	// Get all subnets for ALB (requires 2+ AZs)
	var allSubnets []string
	if !skipALB {
		subnets, err := awsinternal.GetAllDefaultSubnets(ctx, cfg, vpcID)
		if err != nil {
			red.Printf(" x\n")
			yellow.Printf("    Error getting subnets for ALB: %v\n\n", err)
			os.Exit(1)
		}
		for _, s := range subnets {
			allSubnets = append(allSubnets, s.ID)
		}
	}

	green.Printf(" +\n")
	fmt.Printf("    VPC: %s (%s)\n", vpcID, vpcCIDR)
	fmt.Printf("    Subnet: %s (AZ: %s)\n", subnetID, subnetAZ)
	if !skipALB {
		fmt.Printf("    ALB Subnets: %d across multiple AZs\n", len(allSubnets))
	}

	// Pace API calls to prevent rate limiting
	time.Sleep(1 * time.Second)

	// Parallel Resource Creation
	fmt.Printf("  o Creating resources in parallel...\n")

	var wg sync.WaitGroup
	var sgID, albSgID, bucketName string
	var sgErr, albSgErr, iamErr, s3Err error
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

	// ALB Security Group (parallel, only if not skipping ALB)
	if !skipALB {
		wg.Add(1)
		go func() {
			defer wg.Done()
			fmt.Printf("    [%s] Starting: ALB Security Group creation\n", time.Now().Format("15:04:05"))
			albSgID, albSgErr = awsinternal.CreateALBSecurityGroup(ctx, cfg, vpcID, installationKey)
			if albSgErr != nil {
				fmt.Printf("    [%s] Failed: ALB Security Group - %v\n", time.Now().Format("15:04:05"), albSgErr)
			} else {
				createdResources.Lock()
				createdResources.albSgID = albSgID
				createdResources.Unlock()
				fmt.Printf("    [%s] Completed: ALB Security Group\n", time.Now().Format("15:04:05"))
			}
		}()
	}

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
		red.Printf(" x\n")
		yellow.Printf("    Security Group Error: %v\n\n", sgErr)
		os.Exit(1)
	}
	if !skipALB && albSgErr != nil {
		red.Printf(" x\n")
		yellow.Printf("    ALB Security Group Error: %v\n\n", albSgErr)
		os.Exit(1)
	}
	if iamErr != nil {
		red.Printf(" x\n")
		yellow.Printf("    IAM Role Error: %v\n\n", iamErr)
		os.Exit(1)
	}
	if s3Err != nil {
		red.Printf(" x\n")
		yellow.Printf("    S3 Bucket Error: %v\n\n", s3Err)
		os.Exit(1)
	}

	green.Printf(" +\n")
	fmt.Printf("    Security Group: %s\n", sgName)
	if !skipALB {
		fmt.Printf("    ALB Security Group: kl-%s-alb-sg\n", installationKey)
	}
	fmt.Printf("    IAM Role:       %s\n", roleName)
	fmt.Printf("    S3 Bucket:      %s\n", bucketName)

	// Pace API calls to prevent rate limiting (longer delay after parallel operations)
	time.Sleep(2 * time.Second)

	// Create master security group (depends on ALB SG, so must be sequential)
	var masterSgID string
	if !skipALB {
		fmt.Printf("  o Creating master security group...")
		masterSgID, err = awsinternal.EnsureMasterSecurityGroup(ctx, cfg, vpcID, vpcCIDR, albSgID, installationKey)
		if err != nil {
			red.Printf(" x\n")
			yellow.Printf("    Error: %v\n\n", err)
			os.Exit(1)
		}
		createdResources.Lock()
		createdResources.masterSgID = masterSgID
		createdResources.Unlock()
		green.Printf(" +\n")
		fmt.Printf("    Master Security Group: kl-%s-master-sg\n", installationKey)
	}

	// Instance Profile (depends on IAM role)
	bold.Println("\nFinalizing IAM Setup")
	bold.Println("--------------------")
	fmt.Printf("  o Creating instance profile...")
	err = awsinternal.EnsureInstanceProfile(ctx, cfg, installationKey)
	if err != nil {
		red.Printf(" x\n")
		yellow.Printf("    Error: %v\n\n", err)
		os.Exit(1)
	}
	green.Printf(" +\n")

	// Pace API calls to prevent rate limiting
	time.Sleep(1 * time.Second)

	// Instance Launch
	bold.Println("\nInstance Deployment")
	bold.Println("-------------------")

	// Generate K3s agent token
	k3sToken, err := awsinternal.GenerateK3sToken()
	if err != nil {
		red.Printf(" x\n")
		yellow.Printf("    Error generating K3s token: %v\n\n", err)
		os.Exit(1)
	}

	// Use master security group for EC2 when ALB is enabled
	instanceSgID := sgID
	if !skipALB && masterSgID != "" {
		instanceSgID = masterSgID
	}

	fmt.Printf("  o Launching EC2 instance (t3.medium)...")
	instanceID, err := awsinternal.LaunchInstance(ctx, cfg, amiID, subnetID, instanceSgID, vpcID, secretKey, bucketName, k3sToken, installationKey, enableTerminationProtection, fullDomain)
	if err != nil {
		red.Printf(" x\n")
		yellow.Printf("    Error: %v\n\n", err)
		os.Exit(1)
	}
	createdResources.Lock()
	createdResources.instanceID = instanceID
	createdResources.Unlock()
	green.Printf(" +\n")
	fmt.Printf("    %s\n", instanceID)

	fmt.Printf("  o Waiting for instance to be ready...")
	publicIP, privateIP, err := awsinternal.WaitForInstance(ctx, cfg, instanceID)
	if err != nil {
		red.Printf(" x\n")
		yellow.Printf("    Error: %v\n\n", err)
		os.Exit(1)
	}
	green.Printf(" +\n")
	fmt.Printf("    Public IP: %s\n", publicIP)
	fmt.Printf("    Private IP: %s\n", privateIP)

	// ALB and TLS Setup (unless skipping)
	var albDNSName string
	if !skipALB {
		bold.Println("\nLoad Balancer Setup")
		bold.Println("-------------------")

		// Create Target Group
		fmt.Printf("  o Creating target group...")
		tgARN, err := awsinternal.CreateTargetGroup(ctx, cfg, installationKey, vpcID)
		if err != nil {
			red.Printf(" x\n")
			yellow.Printf("    Error: %v\n\n", err)
			os.Exit(1)
		}
		createdResources.Lock()
		createdResources.tgARN = tgARN
		createdResources.Unlock()
		green.Printf(" +\n")

		// Register EC2 instance with target group
		fmt.Printf("  o Registering instance with target group...")
		err = awsinternal.RegisterTargets(ctx, cfg, tgARN, instanceID)
		if err != nil {
			red.Printf(" x\n")
			yellow.Printf("    Error: %v\n\n", err)
			os.Exit(1)
		}
		green.Printf(" +\n")

		// Create ALB
		fmt.Printf("  o Creating Application Load Balancer...")
		albInfo, err := awsinternal.CreateALB(ctx, cfg, installationKey, vpcID, allSubnets, albSgID)
		if err != nil {
			red.Printf(" x\n")
			yellow.Printf("    Error: %v\n\n", err)
			os.Exit(1)
		}
		createdResources.Lock()
		createdResources.albARN = albInfo.ARN
		createdResources.Unlock()
		albDNSName = albInfo.DNSName
		green.Printf(" +\n")
		fmt.Printf("    ALB DNS: %s\n", albDNSName)

		// Wait for ALB to become active
		fmt.Printf("  o Waiting for ALB to become active...")
		err = awsinternal.WaitForALBActive(ctx, cfg, albInfo.ARN)
		if err != nil {
			red.Printf(" x\n")
			yellow.Printf("    Error: %v\n\n", err)
			os.Exit(1)
		}
		green.Printf(" +\n")

		bold.Println("\nTLS Certificate Setup")
		bold.Println("---------------------")

		// Request ACM certificate
		fmt.Printf("  o Requesting ACM certificate for %s...", fullDomain)
		certARN, err := awsinternal.RequestCertificate(ctx, cfg, fullDomain, installationKey)
		if err != nil {
			red.Printf(" x\n")
			yellow.Printf("    Error: %v\n\n", err)
			os.Exit(1)
		}
		createdResources.Lock()
		createdResources.certARN = certARN
		createdResources.Unlock()
		green.Printf(" +\n")

		// Get validation records
		fmt.Printf("  o Getting DNS validation records...")
		validationRecords, err := awsinternal.GetValidationRecords(ctx, cfg, certARN)
		if err != nil {
			red.Printf(" x\n")
			yellow.Printf("    Error: %v\n\n", err)
			os.Exit(1)
		}
		green.Printf(" +\n")
		fmt.Printf("    %d validation record(s) to create\n", len(validationRecords))

		// Send validation records to console for Cloudflare DNS creation
		fmt.Printf("  o Creating DNS validation records in Cloudflare...")
		var consoleRecords []console.ACMValidationRecord
		for _, r := range validationRecords {
			consoleRecords = append(consoleRecords, console.ACMValidationRecord{
				Name:  r.Name,
				Value: r.Value,
			})
		}
		_, err = consoleClient.CreateACMValidationRecords(ctx, installationKey, secretKey, consoleRecords)
		if err != nil {
			red.Printf(" x\n")
			yellow.Printf("    Error: %v\n\n", err)
			os.Exit(1)
		}
		green.Printf(" +\n")

		// Wait for certificate validation
		fmt.Printf("  o Waiting for certificate validation (this may take 2-5 minutes)...")
		err = awsinternal.WaitForValidation(ctx, cfg, certARN, 10*time.Minute)
		if err != nil {
			red.Printf(" x\n")
			yellow.Printf("    Error: %v\n\n", err)
			os.Exit(1)
		}
		green.Printf(" +\n")

		// Create HTTPS listener
		fmt.Printf("  o Creating HTTPS listener...")
		_, err = awsinternal.CreateHTTPSListener(ctx, cfg, albInfo.ARN, tgARN, certARN)
		if err != nil {
			red.Printf(" x\n")
			yellow.Printf("    Error: %v\n\n", err)
			os.Exit(1)
		}
		green.Printf(" +\n")

		// Create HTTP redirect listener
		fmt.Printf("  o Creating HTTP to HTTPS redirect...")
		_, err = awsinternal.CreateHTTPRedirectListener(ctx, cfg, albInfo.ARN)
		if err != nil {
			red.Printf(" x\n")
			yellow.Printf("    Error: %v\n\n", err)
			os.Exit(1)
		}
		green.Printf(" +\n")

		// Register ALB DNS with console for CNAME creation
		bold.Println("\nDNS Configuration")
		bold.Println("-----------------")
		fmt.Printf("  o Configuring DNS for %s...", fullDomain)
		_, err = consoleClient.ConfigureRootDNS(ctx, installationKey, secretKey, albDNSName, "cname")
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
	fmt.Printf("  Instance ID:    %s\n", instanceID)
	fmt.Printf("  Public IP:      %s\n", publicIP)
	fmt.Printf("  Private IP:     %s\n", privateIP)
	fmt.Printf("  Region:         %s\n", cfg.Region)
	fmt.Printf("  AZ:             %s\n", subnetAZ)

	if !skipALB {
		fmt.Println()
		bold.Println("Load Balancer Details")
		bold.Println("---------------------")
		fmt.Printf("  ALB DNS:        %s\n", albDNSName)
		fmt.Printf("  Custom Domain:  https://%s\n", fullDomain)
		fmt.Printf("  Wildcard:       https://*.%s\n", fullDomain)
	}

	fmt.Println()
	bold.Println("Instance Access")
	bold.Println("---------------")
	fmt.Println("  Via AWS Systems Manager:")
	cyan.Printf("    aws ssm start-session --target %s --region %s\n", instanceID, cfg.Region)

	if !skipALB {
		fmt.Println()
		bold.Println("Web Access")
		bold.Println("----------")
		cyan.Printf("    https://%s\n", fullDomain)
	}

	fmt.Println()
}
