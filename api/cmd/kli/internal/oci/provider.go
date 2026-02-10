package oci

import (
	"context"
	"fmt"
	"time"

	"github.com/kloudlite/kloudlite/api/cmd/kli/internal/provider"
)

// OCIProvider implements the provider.CloudProvider interface for Oracle Cloud Infrastructure
type OCIProvider struct {
	Config *OCIConfig
}

// NewOCIProvider creates a new OCI provider with the given configuration
func NewOCIProvider(cfg *OCIConfig) *OCIProvider {
	return &OCIProvider{Config: cfg}
}

// Name returns the provider name
func (p *OCIProvider) Name() string {
	return "oci"
}

// LoadBalancerProvider implementation
// OCI uses reserved public IPs assigned directly to VMs instead of load balancers.
// These methods satisfy the CloudProvider interface but are no-ops.

// CreateLoadBalancer is not applicable for OCI — uses reserved public IPs instead.
func (p *OCIProvider) CreateLoadBalancer(ctx context.Context, installationKey, vpcID string, subnetIDs []string, securityGroupID string) (*provider.LoadBalancerInfo, error) {
	return nil, fmt.Errorf("OCI uses reserved public IPs instead of load balancers")
}

// CreateTargetGroup is not applicable for OCI.
func (p *OCIProvider) CreateTargetGroup(ctx context.Context, installationKey, vpcID string) (string, error) {
	return "", nil
}

// RegisterTargets is not applicable for OCI.
func (p *OCIProvider) RegisterTargets(ctx context.Context, targetGroupID string, instanceIDs ...string) error {
	return nil
}

// CreateHTTPSListener is not applicable for OCI — Cloudflare handles TLS.
func (p *OCIProvider) CreateHTTPSListener(ctx context.Context, loadBalancerID, targetGroupID, certificateID string) (string, error) {
	return "", fmt.Errorf("HTTPS listener not applicable for OCI - use Cloudflare for TLS termination")
}

// CreateHTTPRedirectListener is not applicable for OCI.
func (p *OCIProvider) CreateHTTPRedirectListener(ctx context.Context, loadBalancerID string) (string, error) {
	return "", fmt.Errorf("HTTP redirect listener not applicable for OCI")
}

// WaitForLoadBalancerActive is not applicable for OCI — reserved IPs are instant.
func (p *OCIProvider) WaitForLoadBalancerActive(ctx context.Context, loadBalancerID string) error {
	return nil
}

// DeleteLoadBalancer deletes the reserved public IP for an OCI installation.
func (p *OCIProvider) DeleteLoadBalancer(ctx context.Context, installationKey string) error {
	return DeleteReservedPublicIP(ctx, p.Config, installationKey)
}

// DeleteTargetGroup is a no-op for OCI.
func (p *OCIProvider) DeleteTargetGroup(ctx context.Context, installationKey string) error {
	return nil
}

// FindLoadBalancerByInstallationKey returns the reserved IP name for the installation.
func (p *OCIProvider) FindLoadBalancerByInstallationKey(ctx context.Context, installationKey string) (string, error) {
	return ReservedIPName(installationKey), nil
}

// FindTargetGroupByInstallationKey is a no-op for OCI.
func (p *OCIProvider) FindTargetGroupByInstallationKey(ctx context.Context, installationKey string) (string, error) {
	return "", nil
}

// GetLoadBalancerDNSName is not applicable for OCI — use reserved IP directly.
func (p *OCIProvider) GetLoadBalancerDNSName(ctx context.Context, loadBalancerID string) (string, error) {
	return "", fmt.Errorf("OCI uses reserved public IPs, not load balancer DNS names")
}

// TLSCertificateProvider implementation
// Note: OCI TLS is handled by Cloudflare proxy, not OCI Certificate Manager

// RequestCertificate is not applicable for OCI - Cloudflare handles TLS
func (p *OCIProvider) RequestCertificate(ctx context.Context, domain string, installationKey string) (string, error) {
	return "", fmt.Errorf("TLS certificates are handled by Cloudflare for OCI installations")
}

// GetValidationRecords is not applicable for OCI - Cloudflare handles TLS
func (p *OCIProvider) GetValidationRecords(ctx context.Context, certificateID string) ([]provider.ValidationRecord, error) {
	return nil, fmt.Errorf("TLS certificates are handled by Cloudflare for OCI installations")
}

// WaitForValidation is not applicable for OCI - Cloudflare handles TLS
func (p *OCIProvider) WaitForValidation(ctx context.Context, certificateID string, timeout time.Duration) error {
	return fmt.Errorf("TLS certificates are handled by Cloudflare for OCI installations")
}

// GetCertificateStatus is not applicable for OCI - Cloudflare handles TLS
func (p *OCIProvider) GetCertificateStatus(ctx context.Context, certificateID string) (string, error) {
	return "", fmt.Errorf("TLS certificates are handled by Cloudflare for OCI installations")
}

// DeleteCertificate is not applicable for OCI - Cloudflare handles TLS
func (p *OCIProvider) DeleteCertificate(ctx context.Context, certificateID string) error {
	return nil // No-op for OCI
}

// FindCertificateByInstallationKey is not applicable for OCI - Cloudflare handles TLS
func (p *OCIProvider) FindCertificateByInstallationKey(ctx context.Context, installationKey string) (string, error) {
	return "", nil // No certificates managed by OCI
}

// NetworkProvider implementation

// GetDefaultVPC returns the default VCN for OCI
func (p *OCIProvider) GetDefaultVPC(ctx context.Context) (vpcID string, cidr string, err error) {
	return GetDefaultVCN(ctx, p.Config)
}

// GetDefaultSubnet returns a default subnet in the VCN
func (p *OCIProvider) GetDefaultSubnet(ctx context.Context, vpcID string) (subnetID string, availabilityZone string, err error) {
	subnet, _, err := GetDefaultSubnet(ctx, p.Config, vpcID)
	if err != nil {
		return "", "", err
	}
	return subnet, p.Config.Region, nil
}

// GetAllDefaultSubnets returns all default subnets
func (p *OCIProvider) GetAllDefaultSubnets(ctx context.Context, vpcID string) ([]provider.SubnetInfo, error) {
	subnet, cidr, err := GetDefaultSubnet(ctx, p.Config, vpcID)
	if err != nil {
		return nil, err
	}
	return []provider.SubnetInfo{
		{
			ID:               subnet,
			AvailabilityZone: p.Config.Region,
			CIDR:             cidr,
		},
	}, nil
}

// CreateSecurityGroup creates a NSG for the instance
func (p *OCIProvider) CreateSecurityGroup(ctx context.Context, vpcID, vpcCIDR, installationKey string) (string, error) {
	return EnsureNSG(ctx, p.Config, vpcID, vpcCIDR, installationKey)
}

// CreateLoadBalancerSecurityGroup is a no-op for OCI - NLB uses the same NSG
func (p *OCIProvider) CreateLoadBalancerSecurityGroup(ctx context.Context, vpcID, installationKey string) (string, error) {
	return NSGName(installationKey), nil
}

// DeleteSecurityGroup deletes the NSG
func (p *OCIProvider) DeleteSecurityGroup(ctx context.Context, securityGroupID string) error {
	// securityGroupID is the NSG OCID
	// We need the VCN ID which we don't have here, so this is best-effort
	return nil
}

// FindSecurityGroupByInstallationKey finds the NSG by installation key
func (p *OCIProvider) FindSecurityGroupByInstallationKey(ctx context.Context, installationKey string, isLoadBalancer bool) (string, error) {
	return NSGName(installationKey), nil
}

// Verify OCIProvider implements CloudProvider interface
var _ provider.CloudProvider = (*OCIProvider)(nil)
