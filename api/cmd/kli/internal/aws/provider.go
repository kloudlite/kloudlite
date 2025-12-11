package aws

import (
	"context"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	ec2Types "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/kloudlite/kloudlite/api/cmd/kli/internal/provider"
)

// AWSProvider implements the CloudProvider interface for AWS
type AWSProvider struct {
	cfg aws.Config
}

// NewProvider creates a new AWS provider instance
func NewProvider(cfg aws.Config) *AWSProvider {
	return &AWSProvider{cfg: cfg}
}

// Name returns the provider name
func (p *AWSProvider) Name() string {
	return "aws"
}

// LoadBalancer operations

// CreateLoadBalancer creates an ALB with the given configuration
func (p *AWSProvider) CreateLoadBalancer(ctx context.Context, installationKey, vpcID string, subnetIDs []string, securityGroupID string) (*provider.LoadBalancerInfo, error) {
	albInfo, err := CreateALB(ctx, p.cfg, installationKey, vpcID, subnetIDs, securityGroupID)
	if err != nil {
		return nil, err
	}
	return &provider.LoadBalancerInfo{
		ARN:     albInfo.ARN,
		DNSName: albInfo.DNSName,
	}, nil
}

// CreateTargetGroup creates a target group for the ALB
func (p *AWSProvider) CreateTargetGroup(ctx context.Context, installationKey, vpcID string) (string, error) {
	return CreateTargetGroup(ctx, p.cfg, installationKey, vpcID)
}

// RegisterTargets registers EC2 instances with the target group
func (p *AWSProvider) RegisterTargets(ctx context.Context, targetGroupID string, instanceIDs ...string) error {
	return RegisterTargets(ctx, p.cfg, targetGroupID, instanceIDs...)
}

// CreateHTTPSListener creates an HTTPS listener with TLS termination
func (p *AWSProvider) CreateHTTPSListener(ctx context.Context, loadBalancerID, targetGroupID, certificateID string) (string, error) {
	return CreateHTTPSListener(ctx, p.cfg, loadBalancerID, targetGroupID, certificateID)
}

// CreateHTTPRedirectListener creates an HTTP listener that redirects to HTTPS
func (p *AWSProvider) CreateHTTPRedirectListener(ctx context.Context, loadBalancerID string) (string, error) {
	return CreateHTTPRedirectListener(ctx, p.cfg, loadBalancerID)
}

// WaitForLoadBalancerActive waits for the ALB to become active
func (p *AWSProvider) WaitForLoadBalancerActive(ctx context.Context, loadBalancerID string) error {
	return WaitForALBActive(ctx, p.cfg, loadBalancerID)
}

// DeleteLoadBalancer deletes an ALB by installation key
func (p *AWSProvider) DeleteLoadBalancer(ctx context.Context, installationKey string) error {
	return DeleteALB(ctx, p.cfg, installationKey)
}

// DeleteTargetGroup deletes a target group by installation key
func (p *AWSProvider) DeleteTargetGroup(ctx context.Context, installationKey string) error {
	return DeleteTargetGroup(ctx, p.cfg, installationKey)
}

// FindLoadBalancerByInstallationKey finds an ALB by installation key
func (p *AWSProvider) FindLoadBalancerByInstallationKey(ctx context.Context, installationKey string) (string, error) {
	return FindALBByInstallationKey(ctx, p.cfg, installationKey)
}

// FindTargetGroupByInstallationKey finds a target group by installation key
func (p *AWSProvider) FindTargetGroupByInstallationKey(ctx context.Context, installationKey string) (string, error) {
	return FindTargetGroupByInstallationKey(ctx, p.cfg, installationKey)
}

// GetLoadBalancerDNSName gets the DNS name of an ALB
func (p *AWSProvider) GetLoadBalancerDNSName(ctx context.Context, loadBalancerID string) (string, error) {
	return GetALBDNSName(ctx, p.cfg, loadBalancerID)
}

// TLS Certificate operations

// RequestCertificate requests a new ACM certificate
func (p *AWSProvider) RequestCertificate(ctx context.Context, domain string, installationKey string) (string, error) {
	return RequestCertificate(ctx, p.cfg, domain, installationKey)
}

// GetValidationRecords retrieves DNS validation records for an ACM certificate
func (p *AWSProvider) GetValidationRecords(ctx context.Context, certificateID string) ([]provider.ValidationRecord, error) {
	records, err := GetValidationRecords(ctx, p.cfg, certificateID)
	if err != nil {
		return nil, err
	}
	result := make([]provider.ValidationRecord, len(records))
	for i, r := range records {
		result[i] = provider.ValidationRecord{
			Name:  r.Name,
			Value: r.Value,
			Type:  r.Type,
		}
	}
	return result, nil
}

// WaitForValidation waits for the ACM certificate to be validated
func (p *AWSProvider) WaitForValidation(ctx context.Context, certificateID string, timeout time.Duration) error {
	return WaitForValidation(ctx, p.cfg, certificateID, timeout)
}

// GetCertificateStatus returns the current status of an ACM certificate
func (p *AWSProvider) GetCertificateStatus(ctx context.Context, certificateID string) (string, error) {
	return GetCertificateStatus(ctx, p.cfg, certificateID)
}

// DeleteCertificate deletes an ACM certificate
func (p *AWSProvider) DeleteCertificate(ctx context.Context, certificateID string) error {
	return DeleteCertificate(ctx, p.cfg, certificateID)
}

// FindCertificateByInstallationKey finds an ACM certificate by installation key
func (p *AWSProvider) FindCertificateByInstallationKey(ctx context.Context, installationKey string) (string, error) {
	return FindCertificateByInstallationKey(ctx, p.cfg, installationKey)
}

// Network operations

// GetDefaultVPC returns the default VPC ID and CIDR
func (p *AWSProvider) GetDefaultVPC(ctx context.Context) (string, string, error) {
	return GetDefaultVPC(ctx, p.cfg)
}

// GetDefaultSubnet returns a default subnet in the VPC
func (p *AWSProvider) GetDefaultSubnet(ctx context.Context, vpcID string) (string, string, error) {
	return GetDefaultSubnet(ctx, p.cfg, vpcID)
}

// GetAllDefaultSubnets returns all default subnets in the VPC
func (p *AWSProvider) GetAllDefaultSubnets(ctx context.Context, vpcID string) ([]provider.SubnetInfo, error) {
	return GetAllDefaultSubnets(ctx, p.cfg, vpcID)
}

// CreateSecurityGroup creates a security group for the EC2 instance
func (p *AWSProvider) CreateSecurityGroup(ctx context.Context, vpcID, vpcCIDR, installationKey string) (string, error) {
	return EnsureSecurityGroup(ctx, p.cfg, vpcID, vpcCIDR, installationKey)
}

// CreateLoadBalancerSecurityGroup creates a security group for the ALB
func (p *AWSProvider) CreateLoadBalancerSecurityGroup(ctx context.Context, vpcID, installationKey string) (string, error) {
	return CreateALBSecurityGroup(ctx, p.cfg, vpcID, installationKey)
}

// DeleteSecurityGroup deletes a security group
func (p *AWSProvider) DeleteSecurityGroup(ctx context.Context, securityGroupID string) error {
	ec2Client := ec2.NewFromConfig(p.cfg)
	_, err := ec2Client.DeleteSecurityGroup(ctx, &ec2.DeleteSecurityGroupInput{
		GroupId: aws.String(securityGroupID),
	})
	if err != nil {
		return fmt.Errorf("failed to delete security group: %w", err)
	}
	return nil
}

// FindSecurityGroupByInstallationKey finds a security group by installation key
func (p *AWSProvider) FindSecurityGroupByInstallationKey(ctx context.Context, installationKey string, isLoadBalancer bool) (string, error) {
	ec2Client := ec2.NewFromConfig(p.cfg)

	suffix := "-sg"
	if isLoadBalancer {
		suffix = "-alb-sg"
	}
	sgName := fmt.Sprintf("kl-%s%s", installationKey, suffix)

	result, err := ec2Client.DescribeSecurityGroups(ctx, &ec2.DescribeSecurityGroupsInput{
		Filters: []ec2Types.Filter{
			{
				Name:   aws.String("group-name"),
				Values: []string{sgName},
			},
		},
	})
	if err != nil {
		return "", fmt.Errorf("failed to describe security groups: %w", err)
	}

	if len(result.SecurityGroups) == 0 {
		return "", nil
	}

	return *result.SecurityGroups[0].GroupId, nil
}

// Verify AWSProvider implements CloudProvider interface
var _ provider.CloudProvider = (*AWSProvider)(nil)
