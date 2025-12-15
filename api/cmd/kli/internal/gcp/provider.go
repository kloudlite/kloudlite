package gcp

import (
	"context"
	"fmt"
	"time"

	"github.com/kloudlite/kloudlite/api/cmd/kli/internal/provider"
)

// GCPProvider implements the provider.CloudProvider interface for Google Cloud Platform
type GCPProvider struct {
	Config *GCPConfig
}

// NewGCPProvider creates a new GCP provider with the given configuration
func NewGCPProvider(cfg *GCPConfig) *GCPProvider {
	return &GCPProvider{Config: cfg}
}

// Name returns the provider name
func (p *GCPProvider) Name() string {
	return "gcp"
}

// LoadBalancerProvider implementation

// CreateLoadBalancer creates a GCP HTTP(S) Load Balancer
// Note: GCP LB doesn't use VPC/subnets/security groups in the same way as AWS
// The subnetIDs and securityGroupID parameters are ignored for GCP
func (p *GCPProvider) CreateLoadBalancer(ctx context.Context, installationKey, vpcID string, subnetIDs []string, securityGroupID string) (*provider.LoadBalancerInfo, error) {
	// Reserve external IP
	ip, err := ReserveExternalIP(ctx, p.Config, installationKey)
	if err != nil {
		return nil, fmt.Errorf("failed to reserve external IP: %w", err)
	}

	// Create health check
	hcURL, err := CreateHealthCheck(ctx, p.Config, installationKey)
	if err != nil {
		return nil, fmt.Errorf("failed to create health check: %w", err)
	}

	// Create instance group (assumes VM already exists)
	instanceName := GetInstanceName(installationKey)
	igURL, err := CreateUnmanagedInstanceGroup(ctx, p.Config, instanceName, installationKey)
	if err != nil {
		return nil, fmt.Errorf("failed to create instance group: %w", err)
	}

	// Create backend service
	bsURL, err := CreateBackendService(ctx, p.Config, igURL, hcURL, installationKey)
	if err != nil {
		return nil, fmt.Errorf("failed to create backend service: %w", err)
	}

	// Create URL map
	urlMapURL, err := CreateURLMap(ctx, p.Config, bsURL, installationKey)
	if err != nil {
		return nil, fmt.Errorf("failed to create URL map: %w", err)
	}

	// Create target HTTP proxy
	proxyURL, err := CreateTargetHTTPProxy(ctx, p.Config, urlMapURL, installationKey)
	if err != nil {
		return nil, fmt.Errorf("failed to create target HTTP proxy: %w", err)
	}

	// Create global forwarding rule
	err = CreateGlobalForwardingRule(ctx, p.Config, ip, proxyURL, installationKey)
	if err != nil {
		return nil, fmt.Errorf("failed to create global forwarding rule: %w", err)
	}

	return &provider.LoadBalancerInfo{
		ARN:     fmt.Sprintf("projects/%s/global/forwardingRules/kl-%s-fwd", p.Config.Project, shortKey(installationKey)),
		DNSName: ip, // GCP returns IP address, not DNS name
	}, nil
}

// CreateTargetGroup is not applicable for GCP - uses Instance Groups instead
// Returns the instance group URL as the "target group ID"
func (p *GCPProvider) CreateTargetGroup(ctx context.Context, installationKey, vpcID string) (string, error) {
	instanceName := GetInstanceName(installationKey)
	igURL, err := CreateUnmanagedInstanceGroup(ctx, p.Config, instanceName, installationKey)
	if err != nil {
		return "", err
	}
	return igURL, nil
}

// RegisterTargets is not applicable for GCP in the same way as AWS
// GCP instance groups are managed differently - instances are added during creation
func (p *GCPProvider) RegisterTargets(ctx context.Context, targetGroupID string, instanceIDs ...string) error {
	// GCP instance groups have instances added during creation
	// This is a no-op for GCP
	return nil
}

// CreateHTTPSListener is not applicable for GCP
// GCP handles HTTPS through Target HTTPS Proxy which requires managed certificates
// For our implementation, Cloudflare handles TLS termination
func (p *GCPProvider) CreateHTTPSListener(ctx context.Context, loadBalancerID, targetGroupID, certificateID string) (string, error) {
	// GCP uses Target HTTPS Proxy for HTTPS, but we use Cloudflare for TLS
	return "", fmt.Errorf("HTTPS listener not applicable for GCP - use Cloudflare for TLS termination")
}

// CreateHTTPRedirectListener is not applicable for GCP
// GCP doesn't have listeners in the same way as AWS ALB
func (p *GCPProvider) CreateHTTPRedirectListener(ctx context.Context, loadBalancerID string) (string, error) {
	// GCP handles HTTP through Target HTTP Proxy
	return "", fmt.Errorf("HTTP redirect listener not applicable for GCP architecture")
}

// WaitForLoadBalancerActive waits for the GCP load balancer to become active
func (p *GCPProvider) WaitForLoadBalancerActive(ctx context.Context, loadBalancerID string) error {
	// Extract installation key from the load balancer ID
	// Format: projects/{project}/global/forwardingRules/kl-{key}-fwd
	installationKey := extractInstallationKeyFromLB(loadBalancerID)
	if installationKey == "" {
		return fmt.Errorf("invalid load balancer ID format")
	}
	return WaitForLoadBalancerActive(ctx, p.Config, installationKey)
}

// DeleteLoadBalancer deletes the GCP HTTP(S) Load Balancer
func (p *GCPProvider) DeleteLoadBalancer(ctx context.Context, installationKey string) error {
	return DeleteLoadBalancer(ctx, p.Config, installationKey)
}

// DeleteTargetGroup deletes the instance group (GCP equivalent of target group)
func (p *GCPProvider) DeleteTargetGroup(ctx context.Context, installationKey string) error {
	// Instance group is deleted as part of DeleteLoadBalancer
	return nil
}

// FindLoadBalancerByInstallationKey finds a GCP load balancer by installation key
func (p *GCPProvider) FindLoadBalancerByInstallationKey(ctx context.Context, installationKey string) (string, error) {
	ip, err := GetLoadBalancerIP(ctx, p.Config, installationKey)
	if err != nil {
		return "", err
	}
	if ip == "" {
		return "", nil
	}
	return fmt.Sprintf("projects/%s/global/forwardingRules/kl-%s-fwd", p.Config.Project, shortKey(installationKey)), nil
}

// FindTargetGroupByInstallationKey finds the instance group by installation key
func (p *GCPProvider) FindTargetGroupByInstallationKey(ctx context.Context, installationKey string) (string, error) {
	igName := fmt.Sprintf("kl-%s-ig", shortKey(installationKey))
	return igName, nil
}

// GetLoadBalancerDNSName returns the IP address of the GCP load balancer
// GCP uses static IP addresses, not DNS names
func (p *GCPProvider) GetLoadBalancerDNSName(ctx context.Context, loadBalancerID string) (string, error) {
	installationKey := extractInstallationKeyFromLB(loadBalancerID)
	if installationKey == "" {
		return "", fmt.Errorf("invalid load balancer ID format")
	}
	return GetLoadBalancerIP(ctx, p.Config, installationKey)
}

// TLSCertificateProvider implementation
// Note: GCP TLS is handled by Cloudflare proxy, not GCP Certificate Manager

// RequestCertificate is not applicable for GCP - Cloudflare handles TLS
func (p *GCPProvider) RequestCertificate(ctx context.Context, domain string, installationKey string) (string, error) {
	return "", fmt.Errorf("TLS certificates are handled by Cloudflare for GCP installations")
}

// GetValidationRecords is not applicable for GCP - Cloudflare handles TLS
func (p *GCPProvider) GetValidationRecords(ctx context.Context, certificateID string) ([]provider.ValidationRecord, error) {
	return nil, fmt.Errorf("TLS certificates are handled by Cloudflare for GCP installations")
}

// WaitForValidation is not applicable for GCP - Cloudflare handles TLS
func (p *GCPProvider) WaitForValidation(ctx context.Context, certificateID string, timeout time.Duration) error {
	return fmt.Errorf("TLS certificates are handled by Cloudflare for GCP installations")
}

// GetCertificateStatus is not applicable for GCP - Cloudflare handles TLS
func (p *GCPProvider) GetCertificateStatus(ctx context.Context, certificateID string) (string, error) {
	return "", fmt.Errorf("TLS certificates are handled by Cloudflare for GCP installations")
}

// DeleteCertificate is not applicable for GCP - Cloudflare handles TLS
func (p *GCPProvider) DeleteCertificate(ctx context.Context, certificateID string) error {
	return nil // No-op for GCP
}

// FindCertificateByInstallationKey is not applicable for GCP - Cloudflare handles TLS
func (p *GCPProvider) FindCertificateByInstallationKey(ctx context.Context, installationKey string) (string, error) {
	return "", nil // No certificates managed by GCP
}

// NetworkProvider implementation

// GetDefaultVPC returns the default VPC for GCP
func (p *GCPProvider) GetDefaultVPC(ctx context.Context) (vpcID string, cidr string, err error) {
	vpcID, _, err = GetDefaultVPC(ctx, p.Config)
	if err != nil {
		return "", "", err
	}
	cidr, err = GetVPCCIDR(ctx, p.Config)
	if err != nil {
		return "", "", err
	}
	return vpcID, cidr, nil
}

// GetDefaultSubnet returns a default subnet in the VPC
func (p *GCPProvider) GetDefaultSubnet(ctx context.Context, vpcID string) (subnetID string, availabilityZone string, err error) {
	subnet, _, err := GetDefaultSubnet(ctx, p.Config, vpcID)
	if err != nil {
		return "", "", err
	}
	return subnet, p.Config.Zone, nil
}

// GetAllDefaultSubnets returns all default subnets
// GCP doesn't have the same subnet-per-AZ model as AWS
func (p *GCPProvider) GetAllDefaultSubnets(ctx context.Context, vpcID string) ([]provider.SubnetInfo, error) {
	subnet, cidr, err := GetDefaultSubnet(ctx, p.Config, vpcID)
	if err != nil {
		return nil, err
	}
	return []provider.SubnetInfo{
		{
			ID:               subnet,
			AvailabilityZone: p.Config.Zone,
			CIDR:             cidr,
		},
	}, nil
}

// CreateSecurityGroup creates firewall rules for the instance
// GCP uses firewall rules with network tags instead of security groups
func (p *GCPProvider) CreateSecurityGroup(ctx context.Context, vpcID, vpcCIDR, installationKey string) (string, error) {
	err := EnsureFirewallRules(ctx, p.Config, vpcCIDR, installationKey)
	if err != nil {
		return "", err
	}
	// Return the network tag as the "security group ID"
	return NetworkTag(installationKey), nil
}

// CreateLoadBalancerSecurityGroup creates firewall rules for the load balancer
// For GCP, this creates the health check firewall rule
func (p *GCPProvider) CreateLoadBalancerSecurityGroup(ctx context.Context, vpcID, installationKey string) (string, error) {
	// LB firewall rules are created as part of EnsureFirewallRules
	return NetworkTag(installationKey), nil
}

// DeleteSecurityGroup deletes firewall rules
// For GCP, the securityGroupID is the network tag (installation key)
func (p *GCPProvider) DeleteSecurityGroup(ctx context.Context, securityGroupID string) error {
	// Extract installation key from network tag
	// Network tag format: kl-{key}-vm
	installationKey := extractInstallationKeyFromTag(securityGroupID)
	if installationKey == "" {
		return fmt.Errorf("invalid security group ID (network tag) format")
	}
	return DeleteFirewallRules(ctx, p.Config, installationKey)
}

// FindSecurityGroupByInstallationKey finds firewall rules by installation key
func (p *GCPProvider) FindSecurityGroupByInstallationKey(ctx context.Context, installationKey string, isLoadBalancer bool) (string, error) {
	// For GCP, we use network tags instead of security groups
	// Return the network tag
	return NetworkTag(installationKey), nil
}

// Helper functions

// extractInstallationKeyFromLB extracts the installation key from a GCP load balancer ID
// Format: projects/{project}/global/forwardingRules/kl-{key}-fwd
func extractInstallationKeyFromLB(lbID string) string {
	// Simple extraction - find "kl-" and "-fwd"
	const prefix = "kl-"
	const suffix = "-fwd"

	start := len(lbID) - 1
	for i := len(lbID) - 1; i >= 0; i-- {
		if i+len(suffix) <= len(lbID) && lbID[i:i+len(suffix)] == suffix {
			start = i
			break
		}
	}
	for i := start - 1; i >= 0; i-- {
		if i+len(prefix) <= len(lbID) && lbID[i:i+len(prefix)] == prefix {
			return lbID[i+len(prefix) : start]
		}
	}
	return ""
}

// extractInstallationKeyFromTag extracts the installation key from a network tag
// Format: kl-{key}-vm
func extractInstallationKeyFromTag(tag string) string {
	const prefix = "kl-"
	const suffix = "-vm"

	if len(tag) < len(prefix)+len(suffix) {
		return ""
	}
	if tag[:len(prefix)] != prefix {
		return ""
	}
	if tag[len(tag)-len(suffix):] != suffix {
		return ""
	}
	return tag[len(prefix) : len(tag)-len(suffix)]
}

// Verify GCPProvider implements CloudProvider interface
var _ provider.CloudProvider = (*GCPProvider)(nil)
