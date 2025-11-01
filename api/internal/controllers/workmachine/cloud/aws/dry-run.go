package aws

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/ec2"
	ec2types "github.com/aws/aws-sdk-go-v2/service/ec2/types"

	fn "github.com/kloudlite/kloudlite/api/pkg/operator-toolkit/functions"
)

const (
	dryRunAMI_ID        = "ami-06fa3f12191aa3337"
	dryRunInstance_ID   = "i-0f325ce07e90c9f42"
	dryRunSecurityGroup = "sg-08a3d6add69722a8c"
	dryRunVolumeID      = "vol-056ed4fd0c3697e3d"
)

func (p *provider) dryRunCreateInstance(ctx context.Context) error {
	_, err := p.ec2Client.RunInstances(ctx, &ec2.RunInstancesInput{
		DryRun:       fn.Ptr(true),
		ImageId:      fn.Ptr(dryRunAMI_ID),
		InstanceType: ec2types.InstanceTypeT3Micro,
		MinCount:     fn.Ptr[int32](1),
		MaxCount:     fn.Ptr[int32](1),
	})
	return handleDryRunError(err, "ec2:RunInstances")
}

func (p *provider) dryRunTerminateInstance(ctx context.Context) error {
	_, err := p.ec2Client.TerminateInstances(ctx, &ec2.TerminateInstancesInput{
		DryRun:      fn.Ptr(true),
		InstanceIds: []string{dryRunInstance_ID},
	})
	return handleDryRunError(err, "ec2:TerminateInstances")
}

func (p *provider) dryRunStartInstance(ctx context.Context) error {
	_, err := p.ec2Client.StartInstances(ctx, &ec2.StartInstancesInput{
		DryRun:      fn.Ptr(true),
		InstanceIds: []string{dryRunInstance_ID},
	})
	return handleDryRunError(err, "ec2:StartInstances")
}

func (p *provider) dryRunStopInstances(ctx context.Context) error {
	_, err := p.ec2Client.StopInstances(ctx, &ec2.StopInstancesInput{
		DryRun:      fn.Ptr(true),
		InstanceIds: []string{dryRunInstance_ID},
	})
	return handleDryRunError(err, "ec2:StopInstances")
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
		InstanceId:   fn.Ptr(dryRunInstance_ID),
		InstanceType: &ec2types.AttributeValue{Value: fn.Ptr("t3.micro")},
	})
	return handleDryRunError(err, "ec2:ModifyInstanceAttribute")
}

func (p *provider) dryRunCreateSecurityGroup(ctx context.Context) error {
	_, err := p.ec2Client.CreateSecurityGroup(ctx, &ec2.CreateSecurityGroupInput{
		DryRun:      fn.Ptr(true),
		GroupName:   fn.Ptr(dryRunSecurityGroup),
		Description: fn.Ptr("test"),
	})
	return handleDryRunError(err, "ec2:CreateSecurityGroup")
}

func (p *provider) dryRunDeleteSecurityGroup(ctx context.Context) error {
	_, err := p.ec2Client.DeleteSecurityGroup(ctx, &ec2.DeleteSecurityGroupInput{
		DryRun:  fn.Ptr(true),
		GroupId: fn.Ptr(dryRunSecurityGroup),
	})
	return handleDryRunError(err, "ec2:DeleteSecurityGroup")
}

func (p *provider) dryRunDescribeSecurityGroups(ctx context.Context) error {
	_, err := p.ec2Client.DescribeSecurityGroups(ctx, &ec2.DescribeSecurityGroupsInput{
		MaxResults: fn.Ptr(int32(5)),
	})
	return err
}

func (p *provider) dryRunAuthorizeSecurityGroupIngress(ctx context.Context) error {
	_, err := p.ec2Client.AuthorizeSecurityGroupIngress(ctx, &ec2.AuthorizeSecurityGroupIngressInput{
		DryRun:  fn.Ptr(true),
		GroupId: fn.Ptr(dryRunSecurityGroup),
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

func (p *provider) dryRunAuthorizeSecurityGroupEgress(ctx context.Context) error {
	_, err := p.ec2Client.AuthorizeSecurityGroupEgress(ctx, &ec2.AuthorizeSecurityGroupEgressInput{
		DryRun:  fn.Ptr(true),
		GroupId: fn.Ptr(dryRunSecurityGroup),
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

func (p *provider) dryRunCreateTags(ctx context.Context) error {
	_, err := p.ec2Client.CreateTags(ctx, &ec2.CreateTagsInput{
		DryRun:    fn.Ptr(true),
		Resources: []string{dryRunInstance_ID},
		Tags: []ec2types.Tag{
			{Key: fn.Ptr("test"), Value: fn.Ptr("test")},
		},
	})
	return handleDryRunError(err, "ec2:CreateTags")
}
