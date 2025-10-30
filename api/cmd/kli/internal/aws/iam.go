package aws

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/iam"
	iamTypes "github.com/aws/aws-sdk-go-v2/service/iam/types"
)

func EnsureIAMRole(ctx context.Context, cfg aws.Config, installationKey, bucketName string) (string, error) {
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

func EnsureInstanceProfile(ctx context.Context, cfg aws.Config, installationKey string) error {
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
