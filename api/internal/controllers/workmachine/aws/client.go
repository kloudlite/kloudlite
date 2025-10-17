package aws

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/iam"
	"github.com/aws/aws-sdk-go-v2/service/route53"
	"github.com/kloudlite/kloudlite/api/internal/controllers/workmachine/errors"
)

// Client wraps the AWS SDK v2 clients
type Client struct {
	cfg     aws.Config
	EC2     *ec2.Client
	Route53 *route53.Client
	IAM     *iam.Client
	Region  string
}

// NewClient creates a new AWS client with the given region
// It uses the default credential chain (environment variables, shared credentials file, EC2 instance role)
func NewClient(ctx context.Context, region string) (*Client, error) {
	if region == "" {
		return nil, errors.NewInvalidConfigurationError("region", "region cannot be empty")
	}

	// Load AWS configuration
	cfg, err := config.LoadDefaultConfig(ctx, config.WithRegion(region))
	if err != nil {
		return nil, fmt.Errorf("failed to load AWS config: %w", err)
	}

	return &Client{
		cfg:     cfg,
		EC2:     ec2.NewFromConfig(cfg),
		Route53: route53.NewFromConfig(cfg),
		IAM:     iam.NewFromConfig(cfg),
		Region:  region,
	}, nil
}

// GetConfig returns the AWS configuration
func (c *Client) GetConfig() aws.Config {
	return c.cfg
}
