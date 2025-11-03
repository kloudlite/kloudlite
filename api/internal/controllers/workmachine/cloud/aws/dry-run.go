package aws

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/ec2"
	ec2types "github.com/aws/aws-sdk-go-v2/service/ec2/types"

	fn "github.com/kloudlite/kloudlite/api/pkg/operator-toolkit/functions"
)

// Dry-run Permission Validation Approach
//
// These dry-run checks validate IAM permissions without affecting actual resources.
// We use dummy resource IDs that follow AWS format requirements but don't reference real resources.
//
// AWS Resource ID Formats (must be strictly followed):
// - Instance IDs: i-{17 hex chars} (e.g., i-0123456789abcdef0)
// - Volume IDs: vol-{17 hex chars} (e.g., vol-0123456789abcdef0)
//
// AWS Dry-Run Behavior:
//  1. If IAM permissions are missing: Returns "UnauthorizedOperation" error
//  2. If IAM permissions are granted: Returns one of:
//     a) "DryRunOperation" error (permission check passed, would have succeeded)
//     b) "InvalidInstanceID.NotFound" error (permission check passed, but resource doesn't exist)
//     c) "InvalidVolumeID.NotFound" error (permission check passed, but resource doesn't exist)
//
// For permission validation, both (a) and (b)/(c) are acceptable outcomes - they confirm
// the IAM role has the required permissions. The handleDryRunError function treats both as success.
//
// Why dummy IDs work:
// - AWS validates IAM permissions BEFORE checking if resources exist
// - Permission errors (UnauthorizedOperation) are returned immediately
// - Resource-not-found errors only occur AFTER permission validation passes
const (
	// Using valid format instance ID that is extremely unlikely to exist
	// Format: i-{17 hex chars} - using pattern with zeros and 'd' (hex digit)
	dryRunInstanceID = "i-0000000000000d111" // Exactly 17 hex chars after 'i-'

	// Using valid format volume ID that is extremely unlikely to exist
	// Format: vol-{17 hex chars}
	dryRunVolumeID = "vol-0000000000000d111" // Exactly 17 hex chars after 'vol-'
)

func (p *provider) dryRunCreateInstance(ctx context.Context) error {
	_, err := p.ec2Client.RunInstances(ctx, &ec2.RunInstancesInput{
		DryRun:       fn.Ptr(true),
		ImageId:      fn.Ptr(p.AMI), // Use configured AMI from provider
		InstanceType: ec2types.InstanceTypeT3Micro,
		MinCount:     fn.Ptr[int32](1),
		MaxCount:     fn.Ptr[int32](1),
	})
	return handleDryRunError(err, "ec2:RunInstances")
}

func (p *provider) dryRunTerminateInstance(ctx context.Context) error {
	_, err := p.ec2Client.TerminateInstances(ctx, &ec2.TerminateInstancesInput{
		DryRun:      fn.Ptr(true),
		InstanceIds: []string{dryRunInstanceID},
	})
	return handleDryRunError(err, "ec2:TerminateInstances")
}

func (p *provider) dryRunStartInstance(ctx context.Context) error {
	_, err := p.ec2Client.StartInstances(ctx, &ec2.StartInstancesInput{
		DryRun:      fn.Ptr(true),
		InstanceIds: []string{dryRunInstanceID},
	})
	return handleDryRunError(err, "ec2:StartInstances")
}

func (p *provider) dryRunStopInstances(ctx context.Context) error {
	_, err := p.ec2Client.StopInstances(ctx, &ec2.StopInstancesInput{
		DryRun:      fn.Ptr(true),
		InstanceIds: []string{dryRunInstanceID},
	})
	return handleDryRunError(err, "ec2:StopInstances")
}

func (p *provider) dryRunRebootInstances(ctx context.Context) error {
	_, err := p.ec2Client.RebootInstances(ctx, &ec2.RebootInstancesInput{
		DryRun:      fn.Ptr(true),
		InstanceIds: []string{dryRunInstanceID},
	})
	return handleDryRunError(err, "ec2:RebootInstances")
}

func (p *provider) dryRunDescribeInstances(ctx context.Context) error {
	_, err := p.ec2Client.DescribeInstances(ctx, &ec2.DescribeInstancesInput{
		MaxResults: fn.Ptr(int32(5)),
	})
	// DescribeInstances doesn't support DryRun, so we just check for access denied
	return err
}

func (p *provider) dryRunDescribeInstanceStatus(ctx context.Context) error {
	_, err := p.ec2Client.DescribeInstanceStatus(ctx, &ec2.DescribeInstanceStatusInput{
		MaxResults: fn.Ptr(int32(5)),
	})
	return err
}

func (p *provider) dryRunDescribeVolumes(ctx context.Context) error {
	_, err := p.ec2Client.DescribeVolumes(ctx, &ec2.DescribeVolumesInput{
		MaxResults: fn.Ptr(int32(5)),
	})
	return err
}

func (p *provider) dryRunModifyVolume(ctx context.Context) error {
	_, err := p.ec2Client.ModifyVolume(ctx, &ec2.ModifyVolumeInput{
		DryRun:   fn.Ptr(true),
		VolumeId: fn.Ptr(dryRunVolumeID),
		Size:     fn.Ptr(int32(100)),
	})
	return handleDryRunError(err, "ec2:ModifyVolume")
}

func (p *provider) dryRunModifyInstanceAttribute(ctx context.Context) error {
	_, err := p.ec2Client.ModifyInstanceAttribute(ctx, &ec2.ModifyInstanceAttributeInput{
		DryRun:       fn.Ptr(true),
		InstanceId:   fn.Ptr(dryRunInstanceID),
		InstanceType: &ec2types.AttributeValue{Value: fn.Ptr("t3.micro")},
	})
	return handleDryRunError(err, "ec2:ModifyInstanceAttribute")
}

func (p *provider) dryRunDescribeSecurityGroups(ctx context.Context) error {
	_, err := p.ec2Client.DescribeSecurityGroups(ctx, &ec2.DescribeSecurityGroupsInput{
		MaxResults: fn.Ptr(int32(5)),
	})
	return err
}
