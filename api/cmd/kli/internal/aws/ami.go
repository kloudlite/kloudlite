package aws

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
)

func FindUbuntuAMI(ctx context.Context, cfg aws.Config) (string, error) {
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
