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

// CreateLoadBalancer creates a Standard Load Balancer
func (p *AzureProvider) CreateLoadBalancer(ctx context.Context, installationKey, vpcID string, subnetIDs []string, securityGroupID string) (*provider.LoadBalancerInfo, error) {
	lbInfo, err := CreateLoadBalancer(ctx, p.cfg, installationKey)
	if err != nil {
		return nil, err
	}
	return &provider.LoadBalancerInfo{
		ARN:     lbInfo.ID,
		DNSName: lbInfo.PublicIP,
	}, nil
}

// CreateTargetGroup is a no-op — backend pool is created with the LB
func (p *AzureProvider) CreateTargetGroup(ctx context.Context, installationKey, vpcID string) (string, error) {
	return "", nil
}

// RegisterTargets adds VM NIC to the LB backend pool
func (p *AzureProvider) RegisterTargets(ctx context.Context, targetGroupID string, instanceIDs ...string) error {
	// targetGroupID is unused for Azure LB; we use installation key from NIC naming
	// This is called from the generic provider interface but Azure install uses AddNICToBackendPool directly
	return nil
}

// CreateHTTPSListener creates an HTTPS listener (no-op since Cloudflare handles TLS)
func (p *AzureProvider) CreateHTTPSListener(ctx context.Context, loadBalancerID, targetGroupID, certificateID string) (string, error) {
	return "", nil
}

// CreateHTTPRedirectListener creates an HTTP redirect rule (no-op since Cloudflare handles this)
func (p *AzureProvider) CreateHTTPRedirectListener(ctx context.Context, loadBalancerID string) (string, error) {
	return "", nil
}

// WaitForLoadBalancerActive is a no-op — LB is ready after creation
func (p *AzureProvider) WaitForLoadBalancerActive(ctx context.Context, loadBalancerID string) error {
	return nil
}

// DeleteLoadBalancer deletes the Load Balancer by installation key
func (p *AzureProvider) DeleteLoadBalancer(ctx context.Context, installationKey string) error {
	return DeleteLoadBalancer(ctx, p.cfg, installationKey)
}

// DeleteTargetGroup is a no-op — backend pool is deleted with the LB
func (p *AzureProvider) DeleteTargetGroup(ctx context.Context, installationKey string) error {
	return nil
}

// FindLoadBalancerByInstallationKey finds a Load Balancer by installation key
func (p *AzureProvider) FindLoadBalancerByInstallationKey(ctx context.Context, installationKey string) (string, error) {
	return FindLoadBalancerByInstallationKey(ctx, p.cfg, installationKey)
}

// FindTargetGroupByInstallationKey is a no-op — backend pool is part of the LB
func (p *AzureProvider) FindTargetGroupByInstallationKey(ctx context.Context, installationKey string) (string, error) {
	return "", nil
}

// GetLoadBalancerDNSName gets the public IP of the Load Balancer
func (p *AzureProvider) GetLoadBalancerDNSName(ctx context.Context, loadBalancerID string) (string, error) {
	// Extract installation key from LB name — but we store the full ID
	// Use the PIP lookup approach instead
	installationKey := extractInstallationKey(extractResourceName(loadBalancerID))
	return GetLBPublicIP(ctx, p.cfg, installationKey)
}

// TLS Certificate operations (no-op since Cloudflare handles TLS)

func (p *AzureProvider) RequestCertificate(ctx context.Context, domain string, installationKey string) (string, error) {
	return "", nil
}

func (p *AzureProvider) GetValidationRecords(ctx context.Context, certificateID string) ([]provider.ValidationRecord, error) {
	return []provider.ValidationRecord{}, nil
}

func (p *AzureProvider) WaitForValidation(ctx context.Context, certificateID string, timeout time.Duration) error {
	return nil
}

func (p *AzureProvider) GetCertificateStatus(ctx context.Context, certificateID string) (string, error) {
	return "CLOUDFLARE_MANAGED", nil
}

func (p *AzureProvider) DeleteCertificate(ctx context.Context, certificateID string) error {
	return nil
}

func (p *AzureProvider) FindCertificateByInstallationKey(ctx context.Context, installationKey string) (string, error) {
	return "", nil
}

// Network operations

func (p *AzureProvider) GetDefaultVPC(ctx context.Context) (string, string, error) {
	return GetDefaultVNet(ctx, p.cfg)
}

func (p *AzureProvider) GetDefaultSubnet(ctx context.Context, vpcID string) (string, string, error) {
	return GetDefaultSubnet(ctx, p.cfg, vpcID)
}

func (p *AzureProvider) GetAllDefaultSubnets(ctx context.Context, vpcID string) ([]provider.SubnetInfo, error) {
	return GetAllSubnets(ctx, p.cfg, vpcID)
}

func (p *AzureProvider) CreateSecurityGroup(ctx context.Context, vpcID, vpcCIDR, installationKey string) (string, error) {
	return EnsureNetworkSecurityGroup(ctx, p.cfg, vpcCIDR, installationKey)
}

// CreateLoadBalancerSecurityGroup is a no-op — Standard LB doesn't need a separate NSG
func (p *AzureProvider) CreateLoadBalancerSecurityGroup(ctx context.Context, vpcID, installationKey string) (string, error) {
	return "", nil
}

func (p *AzureProvider) DeleteSecurityGroup(ctx context.Context, securityGroupID string) error {
	return DeleteNSG(ctx, p.cfg, securityGroupID)
}

func (p *AzureProvider) FindSecurityGroupByInstallationKey(ctx context.Context, installationKey string, isLoadBalancer bool) (string, error) {
	return FindNSGByInstallationKey(ctx, p.cfg, installationKey, isLoadBalancer)
}

// Verify AzureProvider implements CloudProvider interface
var _ provider.CloudProvider = (*AzureProvider)(nil)
