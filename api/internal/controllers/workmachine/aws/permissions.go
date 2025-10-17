package aws

import (
	"context"
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go-v2/service/ec2"
	ec2types "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/aws/aws-sdk-go-v2/service/iam"
	"github.com/aws/aws-sdk-go-v2/service/route53"
	route53types "github.com/aws/aws-sdk-go-v2/service/route53/types"
	"github.com/kloudlite/kloudlite/api/internal/controllers/workmachine/errors"
	fn "github.com/kloudlite/kloudlite/api/pkg/operator-toolkit/functions"
)

// ValidatePermissions checks if the AWS credentials have all required permissions
// It performs dry-run operations and describe calls to validate permissions
// Returns nil if all permissions are present, otherwise returns a detailed error
func ValidatePermissions(ctx context.Context, client *Client, vpcID, subnetID, hostedZoneID string) error {
	var missingPermissions []string
	var errs []string

	// Test 1: ec2:RunInstances (dry-run)
	if err := testRunInstances(ctx, client, subnetID); err != nil {
		missingPermissions = append(missingPermissions, "ec2:RunInstances")
		errs = append(errs, fmt.Sprintf("RunInstances: %v", err))
	}

	// Test 2: ec2:TerminateInstances (dry-run)
	if err := testTerminateInstances(ctx, client); err != nil {
		missingPermissions = append(missingPermissions, "ec2:TerminateInstances")
		errs = append(errs, fmt.Sprintf("TerminateInstances: %v", err))
	}

	// Test 3: ec2:StartInstances (dry-run)
	if err := testStartInstances(ctx, client); err != nil {
		missingPermissions = append(missingPermissions, "ec2:StartInstances")
		errs = append(errs, fmt.Sprintf("StartInstances: %v", err))
	}

	// Test 4: ec2:StopInstances (dry-run)
	if err := testStopInstances(ctx, client); err != nil {
		missingPermissions = append(missingPermissions, "ec2:StopInstances")
		errs = append(errs, fmt.Sprintf("StopInstances: %v", err))
	}

	// Test 5: ec2:DescribeInstances
	if err := testDescribeInstances(ctx, client); err != nil {
		missingPermissions = append(missingPermissions, "ec2:DescribeInstances")
		errs = append(errs, fmt.Sprintf("DescribeInstances: %v", err))
	}

	// Test 6: ec2:DescribeInstanceStatus
	if err := testDescribeInstanceStatus(ctx, client); err != nil {
		missingPermissions = append(missingPermissions, "ec2:DescribeInstanceStatus")
		errs = append(errs, fmt.Sprintf("DescribeInstanceStatus: %v", err))
	}

	// Test 7: ec2:CreateSecurityGroup (requires VPC ID)
	if vpcID != "" {
		if err := testCreateSecurityGroup(ctx, client, vpcID); err != nil {
			missingPermissions = append(missingPermissions, "ec2:CreateSecurityGroup")
			errs = append(errs, fmt.Sprintf("CreateSecurityGroup: %v", err))
		}
	}

	// Test 8: ec2:DeleteSecurityGroup (dry-run)
	if err := testDeleteSecurityGroup(ctx, client); err != nil {
		missingPermissions = append(missingPermissions, "ec2:DeleteSecurityGroup")
		errs = append(errs, fmt.Sprintf("DeleteSecurityGroup: %v", err))
	}

	// Test 9: ec2:AuthorizeSecurityGroupIngress (dry-run)
	if err := testAuthorizeSecurityGroupIngress(ctx, client); err != nil {
		missingPermissions = append(missingPermissions, "ec2:AuthorizeSecurityGroupIngress")
		errs = append(errs, fmt.Sprintf("AuthorizeSecurityGroupIngress: %v", err))
	}

	// Test 10: ec2:AuthorizeSecurityGroupEgress (dry-run)
	if err := testAuthorizeSecurityGroupEgress(ctx, client); err != nil {
		missingPermissions = append(missingPermissions, "ec2:AuthorizeSecurityGroupEgress")
		errs = append(errs, fmt.Sprintf("AuthorizeSecurityGroupEgress: %v", err))
	}

	// Test 11: ec2:CreateTags
	if err := testCreateTags(ctx, client); err != nil {
		missingPermissions = append(missingPermissions, "ec2:CreateTags")
		errs = append(errs, fmt.Sprintf("CreateTags: %v", err))
	}

	// Test 12: route53:ChangeResourceRecordSets (requires hosted zone ID)
	if hostedZoneID != "" {
		if err := testRoute53ChangeResourceRecordSets(ctx, client, hostedZoneID); err != nil {
			missingPermissions = append(missingPermissions, "route53:ChangeResourceRecordSets")
			errs = append(errs, fmt.Sprintf("ChangeResourceRecordSets: %v", err))
		}
	}

	// Test 13: route53:GetHostedZone
	if hostedZoneID != "" {
		if err := testRoute53GetHostedZone(ctx, client, hostedZoneID); err != nil {
			missingPermissions = append(missingPermissions, "route53:GetHostedZone")
			errs = append(errs, fmt.Sprintf("GetHostedZone: %v", err))
		}
	}

	// Test 14: iam:GetInstanceProfile (if IAM instance profile is used)
	if err := testIAMGetInstanceProfile(ctx, client); err != nil {
		missingPermissions = append(missingPermissions, "iam:GetInstanceProfile")
		errs = append(errs, fmt.Sprintf("GetInstanceProfile: %v", err))
	}

	if len(missingPermissions) > 0 {
		allRequiredPermissions := []string{
			"ec2:RunInstances",
			"ec2:TerminateInstances",
			"ec2:StartInstances",
			"ec2:StopInstances",
			"ec2:DescribeInstances",
			"ec2:DescribeInstanceStatus",
			"ec2:CreateSecurityGroup",
			"ec2:DeleteSecurityGroup",
			"ec2:AuthorizeSecurityGroupIngress",
			"ec2:AuthorizeSecurityGroupEgress",
			"ec2:CreateTags",
			"route53:ChangeResourceRecordSets",
			"route53:GetHostedZone",
			"iam:GetInstanceProfile",
		}

		return errors.NewPermissionDeniedError(
			"AWS permission validation",
			missingPermissions,
			allRequiredPermissions,
			fmt.Errorf("missing permissions: %s", strings.Join(errs, "; ")),
		)
	}

	return nil
}

// Individual permission test functions

func testRunInstances(ctx context.Context, client *Client, subnetID string) error {
	_, err := client.EC2.RunInstances(ctx, &ec2.RunInstancesInput{
		DryRun:       fn.Ptr(true),
		ImageId:      fn.Ptr("ami-test"),
		InstanceType: ec2types.InstanceTypeT3Micro,
		MinCount:     fn.Ptr(int32(1)),
		MaxCount:     fn.Ptr(int32(1)),
		SubnetId:     fn.Ptr(subnetID),
	})
	return handleDryRunError(err, "ec2:RunInstances")
}

func testTerminateInstances(ctx context.Context, client *Client) error {
	_, err := client.EC2.TerminateInstances(ctx, &ec2.TerminateInstancesInput{
		DryRun:      fn.Ptr(true),
		InstanceIds: []string{"i-test"},
	})
	return handleDryRunError(err, "ec2:TerminateInstances")
}

func testStartInstances(ctx context.Context, client *Client) error {
	_, err := client.EC2.StartInstances(ctx, &ec2.StartInstancesInput{
		DryRun:      fn.Ptr(true),
		InstanceIds: []string{"i-test"},
	})
	return handleDryRunError(err, "ec2:StartInstances")
}

func testStopInstances(ctx context.Context, client *Client) error {
	_, err := client.EC2.StopInstances(ctx, &ec2.StopInstancesInput{
		DryRun:      fn.Ptr(true),
		InstanceIds: []string{"i-test"},
	})
	return handleDryRunError(err, "ec2:StopInstances")
}

func testDescribeInstances(ctx context.Context, client *Client) error {
	_, err := client.EC2.DescribeInstances(ctx, &ec2.DescribeInstancesInput{
		MaxResults: fn.Ptr(int32(5)),
	})
	if err != nil && !isAccessDeniedError(err) {
		return nil // Other errors are acceptable for validation
	}
	return err
}

func testDescribeInstanceStatus(ctx context.Context, client *Client) error {
	_, err := client.EC2.DescribeInstanceStatus(ctx, &ec2.DescribeInstanceStatusInput{
		MaxResults: fn.Ptr(int32(5)),
	})
	if err != nil && !isAccessDeniedError(err) {
		return nil // Other errors are acceptable for validation
	}
	return err
}

func testCreateSecurityGroup(ctx context.Context, client *Client, vpcID string) error {
	_, err := client.EC2.CreateSecurityGroup(ctx, &ec2.CreateSecurityGroupInput{
		DryRun:      fn.Ptr(true),
		GroupName:   fn.Ptr("test-sg"),
		Description: fn.Ptr("test"),
		VpcId:       fn.Ptr(vpcID),
	})
	return handleDryRunError(err, "ec2:CreateSecurityGroup")
}

func testDeleteSecurityGroup(ctx context.Context, client *Client) error {
	_, err := client.EC2.DeleteSecurityGroup(ctx, &ec2.DeleteSecurityGroupInput{
		DryRun:  fn.Ptr(true),
		GroupId: fn.Ptr("sg-test"),
	})
	return handleDryRunError(err, "ec2:DeleteSecurityGroup")
}

func testAuthorizeSecurityGroupIngress(ctx context.Context, client *Client) error {
	_, err := client.EC2.AuthorizeSecurityGroupIngress(ctx, &ec2.AuthorizeSecurityGroupIngressInput{
		DryRun:  fn.Ptr(true),
		GroupId: fn.Ptr("sg-test"),
		IpPermissions: []ec2types.IpPermission{
			{
				IpProtocol: fn.Ptr("tcp"),
				FromPort:   fn.Ptr(int32(443)),
				ToPort:     fn.Ptr(int32(443)),
				IpRanges: []ec2types.IpRange{
					{CidrIp: fn.Ptr("0.0.0.0/0")},
				},
			},
		},
	})
	return handleDryRunError(err, "ec2:AuthorizeSecurityGroupIngress")
}

func testAuthorizeSecurityGroupEgress(ctx context.Context, client *Client) error {
	_, err := client.EC2.AuthorizeSecurityGroupEgress(ctx, &ec2.AuthorizeSecurityGroupEgressInput{
		DryRun:  fn.Ptr(true),
		GroupId: fn.Ptr("sg-test"),
		IpPermissions: []ec2types.IpPermission{
			{
				IpProtocol: fn.Ptr("-1"),
				IpRanges: []ec2types.IpRange{
					{CidrIp: fn.Ptr("0.0.0.0/0")},
				},
			},
		},
	})
	return handleDryRunError(err, "ec2:AuthorizeSecurityGroupEgress")
}

func testCreateTags(ctx context.Context, client *Client) error {
	_, err := client.EC2.CreateTags(ctx, &ec2.CreateTagsInput{
		DryRun:    fn.Ptr(true),
		Resources: []string{"i-test"},
		Tags: []ec2types.Tag{
			{Key: fn.Ptr("test"), Value: fn.Ptr("test")},
		},
	})
	return handleDryRunError(err, "ec2:CreateTags")
}

func testRoute53ChangeResourceRecordSets(ctx context.Context, client *Client, hostedZoneID string) error {
	_, err := client.Route53.ChangeResourceRecordSets(ctx, &route53.ChangeResourceRecordSetsInput{
		HostedZoneId: fn.Ptr(hostedZoneID),
		ChangeBatch: &route53types.ChangeBatch{
			Changes: []route53types.Change{
				{
					Action: route53types.ChangeActionUpsert,
					ResourceRecordSet: &route53types.ResourceRecordSet{
						Name: fn.Ptr("test.example.com"),
						Type: route53types.RRTypeA,
						TTL:  fn.Ptr(int64(300)),
						ResourceRecords: []route53types.ResourceRecord{
							{Value: fn.Ptr("1.2.3.4")},
						},
					},
				},
			},
		},
	})
	if err != nil && !isAccessDeniedError(err) {
		return nil // Other errors (like InvalidHostedZoneId) are acceptable
	}
	return err
}

func testRoute53GetHostedZone(ctx context.Context, client *Client, hostedZoneID string) error {
	_, err := client.Route53.GetHostedZone(ctx, &route53.GetHostedZoneInput{
		Id: fn.Ptr(hostedZoneID),
	})
	if err != nil && !isAccessDeniedError(err) {
		return nil // Other errors are acceptable
	}
	return err
}

func testIAMGetInstanceProfile(ctx context.Context, client *Client) error {
	_, err := client.IAM.GetInstanceProfile(ctx, &iam.GetInstanceProfileInput{
		InstanceProfileName: fn.Ptr("test-profile"),
	})
	if err != nil && !isAccessDeniedError(err) {
		return nil // Other errors are acceptable
	}
	return err
}

// Helper functions

func handleDryRunError(err error, action string) error {
	if err == nil {
		return fmt.Errorf("dry-run for %s should have failed with DryRunOperation", action)
	}

	// Check if it's a DryRunOperation error (indicates permission is granted)
	if strings.Contains(err.Error(), "DryRunOperation") || strings.Contains(err.Error(), "Request would have succeeded") {
		return nil
	}

	// Check if it's an access denied error
	if isAccessDeniedError(err) {
		return err
	}

	// Other errors might indicate the permission exists but parameters are invalid
	// This is acceptable for validation purposes
	return nil
}

func isAccessDeniedError(err error) bool {
	if err == nil {
		return false
	}
	errStr := err.Error()
	return strings.Contains(errStr, "UnauthorizedOperation") ||
		strings.Contains(errStr, "AccessDenied") ||
		strings.Contains(errStr, "Forbidden") ||
		strings.Contains(errStr, "not authorized")
}

