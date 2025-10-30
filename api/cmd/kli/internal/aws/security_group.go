package aws

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
)

func EnsureSecurityGroup(ctx context.Context, cfg aws.Config, vpcID, vpcCIDR string, installationKey string) (string, error) {
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
