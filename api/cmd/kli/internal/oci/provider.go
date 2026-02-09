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

// CreateLoadBalancer creates an OCI Network Load Balancer
func (p *OCIProvider) CreateLoadBalancer(ctx context.Context, installationKey, vpcID string, subnetIDs []string, securityGroupID string) (*provider.LoadBalancerInfo, error) {
	// For OCI, we need instance IP to create the NLB backend
	// This is handled in the install command directly
	return nil, fmt.Errorf("CreateLoadBalancer should be called via OCI-specific functions")
}

// CreateTargetGroup is not applicable for OCI NLB - uses backend sets
func (p *OCIProvider) CreateTargetGroup(ctx context.Context, installationKey, vpcID string) (string, error) {
	return BackendSetName, nil
}

// RegisterTargets is handled during NLB creation for OCI
func (p *OCIProvider) RegisterTargets(ctx context.Context, targetGroupID string, instanceIDs ...string) error {
	return nil
}

// CreateHTTPSListener is not applicable for OCI - Cloudflare handles TLS
func (p *OCIProvider) CreateHTTPSListener(ctx context.Context, loadBalancerID, targetGroupID, certificateID string) (string, error) {
	return "", fmt.Errorf("HTTPS listener not applicable for OCI - use Cloudflare for TLS termination")
}

// CreateHTTPRedirectListener is not applicable for OCI NLB (Layer 4)
func (p *OCIProvider) CreateHTTPRedirectListener(ctx context.Context, loadBalancerID string) (string, error) {
	return "", fmt.Errorf("HTTP redirect listener not applicable for OCI NLB (Layer 4)")
}

// WaitForLoadBalancerActive waits for the NLB to become active
func (p *OCIProvider) WaitForLoadBalancerActive(ctx context.Context, loadBalancerID string) error {
	_, err := WaitForNLBActive(ctx, p.Config, loadBalancerID)
	return err
}

// DeleteLoadBalancer deletes the OCI NLB
func (p *OCIProvider) DeleteLoadBalancer(ctx context.Context, installationKey string) error {
	return DeleteNetworkLoadBalancer(ctx, p.Config, installationKey)
}

// DeleteTargetGroup is a no-op for OCI - backend sets are deleted with the NLB
func (p *OCIProvider) DeleteTargetGroup(ctx context.Context, installationKey string) error {
	return nil
}

// FindLoadBalancerByInstallationKey finds the NLB by installation key
func (p *OCIProvider) FindLoadBalancerByInstallationKey(ctx context.Context, installationKey string) (string, error) {
	ip, err := GetNLBIP(ctx, p.Config, installationKey)
	if err != nil {
		return "", err
	}
	if ip == "" {
		return "", nil
	}
	return NLBName(installationKey), nil
}

// FindTargetGroupByInstallationKey returns the backend set name
func (p *OCIProvider) FindTargetGroupByInstallationKey(ctx context.Context, installationKey string) (string, error) {
	return BackendSetName, nil
}

// GetLoadBalancerDNSName returns the IP address of the NLB
func (p *OCIProvider) GetLoadBalancerDNSName(ctx context.Context, loadBalancerID string) (string, error) {
	// loadBalancerID here is the NLB name, extract installation key
	return "", fmt.Errorf("use GetNLBIP instead")
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
