package azure

import (
	"context"
	"time"

	"github.com/kloudlite/kloudlite/api/cmd/kli/internal/provider"
)

// AzureProvider implements the CloudProvider interface for Azure
type AzureProvider struct {
	cfg *AzureConfig
}

// NewProvider creates a new Azure provider instance
func NewProvider(cfg *AzureConfig) *AzureProvider {
	return &AzureProvider{cfg: cfg}
}

// Name returns the provider name
func (p *AzureProvider) Name() string {
	return "azure"
}

// LoadBalancer operations

// CreateLoadBalancer creates an Application Gateway with the given configuration
func (p *AzureProvider) CreateLoadBalancer(ctx context.Context, installationKey, vpcID string, subnetIDs []string, securityGroupID string) (*provider.LoadBalancerInfo, error) {
	appGwInfo, err := CreateApplicationGateway(ctx, p.cfg, installationKey, vpcID, subnetIDs)
	if err != nil {
		return nil, err
	}
	return &provider.LoadBalancerInfo{
		ARN:     appGwInfo.ID,
		DNSName: appGwInfo.PublicIP,
	}, nil
}

// CreateTargetGroup creates a backend pool for the Application Gateway
func (p *AzureProvider) CreateTargetGroup(ctx context.Context, installationKey, vpcID string) (string, error) {
	return CreateBackendPool(ctx, p.cfg, installationKey)
}

// RegisterTargets registers VMs with the backend pool
func (p *AzureProvider) RegisterTargets(ctx context.Context, targetGroupID string, instanceIDs ...string) error {
	return RegisterBackendTargets(ctx, p.cfg, targetGroupID, instanceIDs...)
}

// CreateHTTPSListener creates an HTTPS listener (no-op since Cloudflare handles TLS)
func (p *AzureProvider) CreateHTTPSListener(ctx context.Context, loadBalancerID, targetGroupID, certificateID string) (string, error) {
	// No-op: Cloudflare handles TLS termination
	return "", nil
}

// CreateHTTPRedirectListener creates an HTTP redirect rule (no-op since Cloudflare handles this)
func (p *AzureProvider) CreateHTTPRedirectListener(ctx context.Context, loadBalancerID string) (string, error) {
	// No-op: Cloudflare handles HTTP to HTTPS redirect
	return "", nil
}

// WaitForLoadBalancerActive waits for the Application Gateway to become active
func (p *AzureProvider) WaitForLoadBalancerActive(ctx context.Context, loadBalancerID string) error {
	return WaitForAppGatewayActive(ctx, p.cfg, loadBalancerID)
}

// DeleteLoadBalancer deletes an Application Gateway by installation key
func (p *AzureProvider) DeleteLoadBalancer(ctx context.Context, installationKey string) error {
	return DeleteApplicationGateway(ctx, p.cfg, installationKey)
}

// DeleteTargetGroup deletes a backend pool by installation key
func (p *AzureProvider) DeleteTargetGroup(ctx context.Context, installationKey string) error {
	// Backend pool is part of App Gateway, deleted with it
	return nil
}

// FindLoadBalancerByInstallationKey finds an Application Gateway by installation key
func (p *AzureProvider) FindLoadBalancerByInstallationKey(ctx context.Context, installationKey string) (string, error) {
	return FindAppGatewayByInstallationKey(ctx, p.cfg, installationKey)
}

// FindTargetGroupByInstallationKey finds a backend pool by installation key
func (p *AzureProvider) FindTargetGroupByInstallationKey(ctx context.Context, installationKey string) (string, error) {
	return FindBackendPoolByInstallationKey(ctx, p.cfg, installationKey)
}

// GetLoadBalancerDNSName gets the public IP of an Application Gateway
func (p *AzureProvider) GetLoadBalancerDNSName(ctx context.Context, loadBalancerID string) (string, error) {
	return GetAppGatewayPublicIP(ctx, p.cfg, loadBalancerID)
}

// TLS Certificate operations (no-op since Cloudflare handles TLS)

// RequestCertificate requests a new TLS certificate (no-op)
func (p *AzureProvider) RequestCertificate(ctx context.Context, domain string, installationKey string) (string, error) {
	// No-op: Cloudflare handles TLS termination
	return "", nil
}

// GetValidationRecords retrieves DNS validation records (no-op)
func (p *AzureProvider) GetValidationRecords(ctx context.Context, certificateID string) ([]provider.ValidationRecord, error) {
	// No-op: Cloudflare handles TLS termination
	return []provider.ValidationRecord{}, nil
}

// WaitForValidation waits for the certificate to be validated (no-op)
func (p *AzureProvider) WaitForValidation(ctx context.Context, certificateID string, timeout time.Duration) error {
	// No-op: Cloudflare handles TLS termination
	return nil
}

// GetCertificateStatus returns the current status of a certificate (no-op)
func (p *AzureProvider) GetCertificateStatus(ctx context.Context, certificateID string) (string, error) {
	// No-op: Cloudflare handles TLS termination
	return "CLOUDFLARE_MANAGED", nil
}

// DeleteCertificate deletes a TLS certificate (no-op)
func (p *AzureProvider) DeleteCertificate(ctx context.Context, certificateID string) error {
	// No-op: Cloudflare handles TLS termination
	return nil
}

// FindCertificateByInstallationKey finds a certificate by installation key (no-op)
func (p *AzureProvider) FindCertificateByInstallationKey(ctx context.Context, installationKey string) (string, error) {
	// No-op: Cloudflare handles TLS termination
	return "", nil
}

// Network operations

// GetDefaultVPC returns the default VNet ID and CIDR
func (p *AzureProvider) GetDefaultVPC(ctx context.Context) (string, string, error) {
	return GetDefaultVNet(ctx, p.cfg)
}

// GetDefaultSubnet returns a default subnet in the VNet
func (p *AzureProvider) GetDefaultSubnet(ctx context.Context, vpcID string) (string, string, error) {
	return GetDefaultSubnet(ctx, p.cfg, vpcID)
}

// GetAllDefaultSubnets returns all default subnets in the VNet
func (p *AzureProvider) GetAllDefaultSubnets(ctx context.Context, vpcID string) ([]provider.SubnetInfo, error) {
	return GetAllSubnets(ctx, p.cfg, vpcID)
}

// CreateSecurityGroup creates a Network Security Group for the VM
func (p *AzureProvider) CreateSecurityGroup(ctx context.Context, vpcID, vpcCIDR, installationKey string) (string, error) {
	return EnsureNetworkSecurityGroup(ctx, p.cfg, vpcCIDR, installationKey)
}

// CreateLoadBalancerSecurityGroup creates a Network Security Group for the Application Gateway
func (p *AzureProvider) CreateLoadBalancerSecurityGroup(ctx context.Context, vpcID, installationKey string) (string, error) {
	return CreateAppGatewayNSG(ctx, p.cfg, installationKey)
}

// DeleteSecurityGroup deletes a Network Security Group
func (p *AzureProvider) DeleteSecurityGroup(ctx context.Context, securityGroupID string) error {
	return DeleteNSG(ctx, p.cfg, securityGroupID)
}

// FindSecurityGroupByInstallationKey finds a Network Security Group by installation key
func (p *AzureProvider) FindSecurityGroupByInstallationKey(ctx context.Context, installationKey string, isLoadBalancer bool) (string, error) {
	return FindNSGByInstallationKey(ctx, p.cfg, installationKey, isLoadBalancer)
}

// Verify AzureProvider implements CloudProvider interface
var _ provider.CloudProvider = (*AzureProvider)(nil)
