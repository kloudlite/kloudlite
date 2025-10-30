package aws

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
)

func GetDefaultVPC(ctx context.Context, cfg aws.Config) (string, string, error) {
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

func GetDefaultSubnet(ctx context.Context, cfg aws.Config, vpcID string) (string, string, error) {
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
