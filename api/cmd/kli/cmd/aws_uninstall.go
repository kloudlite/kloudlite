package cmd

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"sync"
	"syscall"
	"time"

	awsinternal "github.com/kloudlite/kloudlite/api/cmd/kli/internal/aws"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/aws/aws-sdk-go-v2/service/iam"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	s3Types "github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

var awsUninstallCmd = &cobra.Command{
	Use:   "uninstall",
	Short: "Uninstall Kloudlite from AWS",
	Long: `Uninstall Kloudlite from AWS by removing all resources created during installation.

This command will:
  - Terminate EC2 instance 'kl-{installation-key}-instance'
  - Delete security group 'kl-{installation-key}-sg'
  - Delete SSH key pair 'kl-{installation-key}-key' and local key file
  - Delete IAM instance profile 'kl-{installation-key}-role'
  - Delete IAM role 'kl-{installation-key}-role'
  - Delete S3 bucket 'kl-{installation-key}-backups' and all backups

All resources are identified by the InstallationKey tag.`,
	Example: `  # Uninstall using default AWS region from config
  kli aws uninstall --installation-key prod

  # Uninstall from a specific region
  kli aws uninstall --installation-key staging --region us-west-2`,
	Run: runAWSUninstall,
}

var uninstallRegion string
var uninstallKey string

func init() {
	awsUninstallCmd.Flags().StringVar(&uninstallRegion, "region", "", "AWS region (uses default from AWS config if not specified)")
	awsUninstallCmd.Flags().StringVar(&uninstallKey, "installation-key", "", "Installation key to identify this installation (required)")
	awsUninstallCmd.MarkFlagRequired("installation-key")
}

func runAWSUninstall(cmd *cobra.Command, args []string) {
	green := color.New(color.FgGreen, color.Bold)
	red := color.New(color.FgRed, color.Bold)
	yellow := color.New(color.FgYellow, color.Bold)
	cyan := color.New(color.FgCyan, color.Bold)
	bold := color.New(color.Bold)

	// Header
	fmt.Println()
	cyan.Println("╭─────────────────────────────────────────╮")
	cyan.Println("│   Kloudlite AWS Uninstallation          │")
	cyan.Println("╰─────────────────────────────────────────╯")
	fmt.Println()

	ctx := context.Background()

	// Setup signal handling - for uninstallation, we don't want to abort mid-cleanup
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-sigChan
		fmt.Println()
		yellow.Println("\n⚠ Interrupt received. Uninstallation will continue to completion...")
		yellow.Println("   (Aborting now may leave orphaned resources)")
		// Don't exit - let uninstallation complete
	}()

	// Configuration
	bold.Println("Configuration")
	bold.Println("─────────────")
	fmt.Printf("  Installation Key: %s\n", uninstallKey)
	fmt.Printf("  Region:          ")
	cfg, err := awsinternal.LoadAWSConfig(ctx, uninstallRegion)
	if err != nil {
		red.Printf("✗\n")
		yellow.Printf("  Error: %v\n\n", err)
		os.Exit(1)
	}
	green.Printf("✓ %s\n", cfg.Region)
	fmt.Println()

	// Resource Cleanup
	bold.Println("Removing Resources")
	bold.Println("──────────────────")

	fmt.Printf("  ○ Cleaning up resources in parallel...\n")

	var wg sync.WaitGroup
	var instanceCount int
	var sgErr, keyErr, iamErr, s3Err error
	sgName := fmt.Sprintf("kl-%s-sg", uninstallKey)
	keyName := fmt.Sprintf("kl-%s-key", uninstallKey)
	roleName := fmt.Sprintf("kl-%s-role", uninstallKey)
	bucketName := fmt.Sprintf("kl-%s-backups", uninstallKey)

	startTime := time.Now()

	// Terminate instances and delete security group (parallel, SG has retry logic)
	wg.Add(1)
	go func() {
		defer wg.Done()
		fmt.Printf("    [%s] Starting: Finding and terminating instances\n", time.Now().Format("15:04:05"))
		instanceIDs, err := findInstancesByKey(ctx, cfg, uninstallKey)
		if err == nil && len(instanceIDs) > 0 {
			instanceCount = len(instanceIDs)
			fmt.Printf("    [%s] Terminating %d instance(s)\n", time.Now().Format("15:04:05"), instanceCount)
			_ = terminateInstances(ctx, cfg, instanceIDs)
			fmt.Printf("    [%s] Instances termination initiated\n", time.Now().Format("15:04:05"))
		} else {
			fmt.Printf("    [%s] No instances found\n", time.Now().Format("15:04:05"))
		}
		// Delete SG with retry logic to wait for instances
		fmt.Printf("    [%s] Starting: Security Group deletion (with retries)\n", time.Now().Format("15:04:05"))
		sgErr = deleteSecurityGroup(ctx, cfg, uninstallKey)
		if sgErr != nil {
			fmt.Printf("    [%s] Failed: Security Group - %v\n", time.Now().Format("15:04:05"), sgErr)
		} else {
			fmt.Printf("    [%s] Completed: Security Group\n", time.Now().Format("15:04:05"))
		}
	}()

	// Delete SSH key pair (parallel, completely independent)
	wg.Add(1)
	go func() {
		defer wg.Done()
		fmt.Printf("    [%s] Starting: SSH Key Pair deletion\n", time.Now().Format("15:04:05"))
		keyErr = deleteKeyPair(ctx, cfg, uninstallKey)
		if keyErr != nil {
			fmt.Printf("    [%s] Failed: SSH Key - %v\n", time.Now().Format("15:04:05"), keyErr)
		} else {
			fmt.Printf("    [%s] Completed: SSH Key\n", time.Now().Format("15:04:05"))
		}
	}()

	// Delete IAM resources (parallel, independent of instances/sg/keys)
	wg.Add(1)
	go func() {
		defer wg.Done()
		fmt.Printf("    [%s] Starting: IAM cleanup\n", time.Now().Format("15:04:05"))
		// Delete instance profile first
		fmt.Printf("    [%s] Deleting instance profile\n", time.Now().Format("15:04:05"))
		_ = deleteInstanceProfile(ctx, cfg, uninstallKey)
		// Then delete IAM role
		fmt.Printf("    [%s] Deleting IAM role\n", time.Now().Format("15:04:05"))
		iamErr = deleteIAMRole(ctx, cfg, uninstallKey)
		if iamErr != nil {
			fmt.Printf("    [%s] Failed: IAM - %v\n", time.Now().Format("15:04:05"), iamErr)
		} else {
			fmt.Printf("    [%s] Completed: IAM cleanup\n", time.Now().Format("15:04:05"))
		}
	}()

	// Delete S3 bucket (parallel, independent of other resources)
	wg.Add(1)
	go func() {
		defer wg.Done()
		fmt.Printf("    [%s] Starting: S3 Bucket deletion\n", time.Now().Format("15:04:05"))
		s3Err = deleteS3BucketWithBackups(ctx, cfg, bucketName)
		if s3Err != nil {
			fmt.Printf("    [%s] Failed: S3 Bucket - %v\n", time.Now().Format("15:04:05"), s3Err)
		} else {
			fmt.Printf("    [%s] Completed: S3 Bucket\n", time.Now().Format("15:04:05"))
		}
	}()

	wg.Wait()
	elapsed := time.Since(startTime)
	fmt.Printf("    Parallel operations completed in %.1fs\n", elapsed.Seconds())

	// Report results
	if sgErr != nil || keyErr != nil || iamErr != nil || s3Err != nil {
		red.Printf(" ✗\n")
		if sgErr != nil {
			yellow.Printf("    Security Group: %v\n", sgErr)
		}
		if keyErr != nil {
			yellow.Printf("    SSH Key: %v\n", keyErr)
		}
		if iamErr != nil {
			yellow.Printf("    IAM: %v\n", iamErr)
		}
		if s3Err != nil {
			yellow.Printf("    S3 Bucket: %v\n", s3Err)
		}
	} else {
		green.Printf(" ✓\n")
	}

	// Summary of what was deleted
	if instanceCount > 0 {
		fmt.Printf("    Instances:      %d terminated\n", instanceCount)
	}
	if sgErr == nil {
		fmt.Printf("    Security Group: %s\n", sgName)
	}
	if keyErr == nil {
		fmt.Printf("    SSH Key:        %s\n", keyName)
	}
	if iamErr == nil {
		fmt.Printf("    IAM Role:       %s\n", roleName)
	}
	if s3Err == nil {
		fmt.Printf("    S3 Bucket:      %s\n", bucketName)
	}

	// Success Summary
	fmt.Println()
	green.Println("╭─────────────────────────────────────────╮")
	green.Println("│   ✓ Uninstallation Complete!           │")
	green.Println("╰─────────────────────────────────────────╯")
	fmt.Println()
	fmt.Printf("All resources for installation key '%s' have been removed.\n", uninstallKey)
	fmt.Println()
}

func findInstancesByKey(ctx context.Context, cfg aws.Config, installationKey string) ([]string, error) {
	ec2Client := ec2.NewFromConfig(cfg)

	result, err := ec2Client.DescribeInstances(ctx, &ec2.DescribeInstancesInput{
		Filters: []types.Filter{
			{
				Name:   aws.String("tag:kloudlite.io/installation-id"),
				Values: []string{installationKey},
			},
			{
				Name: aws.String("instance-state-name"),
				Values: []string{
					string(types.InstanceStateNameRunning),
					string(types.InstanceStateNamePending),
					string(types.InstanceStateNameStopping),
					string(types.InstanceStateNameStopped),
				},
			},
		},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to describe instances: %w", err)
	}

	var instanceIDs []string
	for _, reservation := range result.Reservations {
		for _, instance := range reservation.Instances {
			if instance.InstanceId != nil {
				instanceIDs = append(instanceIDs, *instance.InstanceId)
			}
		}
	}

	return instanceIDs, nil
}

func terminateInstances(ctx context.Context, cfg aws.Config, instanceIDs []string) error {
	ec2Client := ec2.NewFromConfig(cfg)

	// Disable termination protection for each instance before terminating
	for _, instanceID := range instanceIDs {
		_, err := ec2Client.ModifyInstanceAttribute(ctx, &ec2.ModifyInstanceAttributeInput{
			InstanceId: aws.String(instanceID),
			DisableApiTermination: &types.AttributeBooleanValue{
				Value: aws.Bool(false),
			},
		})
		if err != nil {
			// Log the error but continue - instance might not have protection enabled
			fmt.Printf("    [%s] Warning: Failed to disable termination protection for %s: %v\n",
				time.Now().Format("15:04:05"), instanceID, err)
		}
	}

	_, err := ec2Client.TerminateInstances(ctx, &ec2.TerminateInstancesInput{
		InstanceIds: instanceIDs,
	})
	if err != nil {
		return fmt.Errorf("failed to terminate instances: %w", err)
	}

	return nil
}

func deleteSecurityGroup(ctx context.Context, cfg aws.Config, installationKey string) error {
	ec2Client := ec2.NewFromConfig(cfg)
	sgName := fmt.Sprintf("kl-%s-sg", installationKey)

	// Find security group by name and installation ID tag
	descResult, err := ec2Client.DescribeSecurityGroups(ctx, &ec2.DescribeSecurityGroupsInput{
		Filters: []types.Filter{
			{
				Name:   aws.String("group-name"),
				Values: []string{sgName},
			},
			{
				Name:   aws.String("tag:kloudlite.io/installation-id"),
				Values: []string{installationKey},
			},
		},
	})
	if err != nil {
		return fmt.Errorf("failed to describe security groups: %w", err)
	}

	if len(descResult.SecurityGroups) == 0 {
		return fmt.Errorf("security group not found")
	}

	sgID := *descResult.SecurityGroups[0].GroupId

	// Retry deletion with exponential backoff for dependency violations
	// Increased retries and wait time to handle network interface detachment
	maxRetries := 12
	for i := 0; i < maxRetries; i++ {
		if i > 0 {
			// Progressive backoff: 10s, 15s, 20s, 25s, 30s, then 30s for remaining
			waitTime := time.Duration(10+min(i*5, 20)) * time.Second
			fmt.Printf("    [%s] Security Group retry %d/%d, waiting %ds...\n",
				time.Now().Format("15:04:05"), i, maxRetries-1, int(waitTime.Seconds()))
			time.Sleep(waitTime)
		}

		// Before attempting deletion, check for and detach any network interfaces
		if i > 0 && i%3 == 0 {
			// Every 3rd retry, actively check for and detach network interfaces
			fmt.Printf("    [%s] Checking for attached network interfaces...\n", time.Now().Format("15:04:05"))
			descNIResult, err := ec2Client.DescribeNetworkInterfaces(ctx, &ec2.DescribeNetworkInterfacesInput{
				Filters: []types.Filter{
					{
						Name:   aws.String("group-id"),
						Values: []string{sgID},
					},
				},
			})
			if err == nil && len(descNIResult.NetworkInterfaces) > 0 {
				fmt.Printf("    [%s] Found %d network interface(s) still attached, waiting for detachment...\n",
					time.Now().Format("15:04:05"), len(descNIResult.NetworkInterfaces))
				for _, ni := range descNIResult.NetworkInterfaces {
					if ni.Attachment != nil && ni.Attachment.AttachmentId != nil {
						fmt.Printf("    [%s] Network interface %s is still attached\n",
							time.Now().Format("15:04:05"), *ni.NetworkInterfaceId)
					}
				}
			}
		}

		_, err = ec2Client.DeleteSecurityGroup(ctx, &ec2.DeleteSecurityGroupInput{
			GroupId: aws.String(sgID),
		})
		if err == nil {
			fmt.Printf("    [%s] Security Group deleted successfully\n", time.Now().Format("15:04:05"))
			return nil
		}

		// Check if it's a dependency violation
		errMsg := err.Error()
		if i < maxRetries-1 && strings.Contains(errMsg, "DependencyViolation") {
			fmt.Printf("    [%s] Security Group has dependencies, will retry...\n", time.Now().Format("15:04:05"))
			// Retry - network interfaces may still be detaching
			continue
		}

		// Other error or final retry - return error
		fmt.Printf("    [%s] Security Group deletion failed: %v\n", time.Now().Format("15:04:05"), err)
		return fmt.Errorf("failed to delete security group: %w", err)
	}

	return nil
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func deleteKeyPair(ctx context.Context, cfg aws.Config, installationKey string) error {
	ec2Client := ec2.NewFromConfig(cfg)
	keyName := fmt.Sprintf("kl-%s-key", installationKey)

	// Delete key pair from AWS
	_, err := ec2Client.DeleteKeyPair(ctx, &ec2.DeleteKeyPairInput{
		KeyName: aws.String(keyName),
	})
	if err != nil {
		return fmt.Errorf("failed to delete key pair: %w", err)
	}

	// Delete local key file
	keyPath := filepath.Join(os.Getenv("HOME"), ".kl", fmt.Sprintf("kl-%s-key.pem", installationKey))
	if _, err := os.Stat(keyPath); err == nil {
		os.Remove(keyPath)
	}

	return nil
}

func deleteInstanceProfile(ctx context.Context, cfg aws.Config, installationKey string) error {
	iamClient := iam.NewFromConfig(cfg)
	profileName := fmt.Sprintf("kl-%s-role", installationKey)
	roleName := fmt.Sprintf("kl-%s-role", installationKey)

	// Get instance profile to check if it exists
	_, err := iamClient.GetInstanceProfile(ctx, &iam.GetInstanceProfileInput{
		InstanceProfileName: aws.String(profileName),
	})
	if err != nil {
		// Profile doesn't exist, skip
		return nil
	}

	// Remove role from instance profile
	_, err = iamClient.RemoveRoleFromInstanceProfile(ctx, &iam.RemoveRoleFromInstanceProfileInput{
		InstanceProfileName: aws.String(profileName),
		RoleName:            aws.String(roleName),
	})
	if err != nil {
		// Ignore error if role is not in profile
	}

	// Delete instance profile
	_, err = iamClient.DeleteInstanceProfile(ctx, &iam.DeleteInstanceProfileInput{
		InstanceProfileName: aws.String(profileName),
	})
	if err != nil {
		return fmt.Errorf("failed to delete instance profile: %w", err)
	}

	return nil
}

func deleteIAMRole(ctx context.Context, cfg aws.Config, installationKey string) error {
	iamClient := iam.NewFromConfig(cfg)
	roleName := fmt.Sprintf("kl-%s-role", installationKey)

	// Check if role exists
	_, err := iamClient.GetRole(ctx, &iam.GetRoleInput{
		RoleName: aws.String(roleName),
	})
	if err != nil {
		// Role doesn't exist, skip
		return nil
	}

	// Delete inline policies
	listPoliciesResult, err := iamClient.ListRolePolicies(ctx, &iam.ListRolePoliciesInput{
		RoleName: aws.String(roleName),
	})
	if err == nil {
		for _, policyName := range listPoliciesResult.PolicyNames {
			_, _ = iamClient.DeleteRolePolicy(ctx, &iam.DeleteRolePolicyInput{
				RoleName:   aws.String(roleName),
				PolicyName: aws.String(policyName),
			})
		}
	}

	// Detach managed policies
	listAttachedResult, err := iamClient.ListAttachedRolePolicies(ctx, &iam.ListAttachedRolePoliciesInput{
		RoleName: aws.String(roleName),
	})
	if err == nil {
		for _, policy := range listAttachedResult.AttachedPolicies {
			_, _ = iamClient.DetachRolePolicy(ctx, &iam.DetachRolePolicyInput{
				RoleName:  aws.String(roleName),
				PolicyArn: policy.PolicyArn,
			})
		}
	}

	// Delete role
	_, err = iamClient.DeleteRole(ctx, &iam.DeleteRoleInput{
		RoleName: aws.String(roleName),
	})
	if err != nil {
		return fmt.Errorf("failed to delete IAM role: %w", err)
	}

	return nil
}

func deleteS3BucketWithBackups(ctx context.Context, cfg aws.Config, bucketName string) error {
	s3Client := s3.NewFromConfig(cfg)

	// Check if bucket exists
	_, err := s3Client.HeadBucket(ctx, &s3.HeadBucketInput{
		Bucket: aws.String(bucketName),
	})
	if err != nil {
		// Bucket doesn't exist, skip
		return nil
	}

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
	_, err = s3Client.DeleteBucket(ctx, &s3.DeleteBucketInput{
		Bucket: aws.String(bucketName),
	})
	if err != nil {
		return fmt.Errorf("failed to delete bucket: %w", err)
	}

	return nil
}
