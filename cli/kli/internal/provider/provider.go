package provider

import (
	"context"
	"time"
)

// LoadBalancerInfo contains information about a created load balancer
type LoadBalancerInfo struct {
	ARN     string // Provider-specific identifier (ARN for AWS, resource ID for GCP/Azure)
	DNSName string // DNS name for accessing the load balancer
}

// ValidationRecord represents a DNS validation record for TLS certificates
type ValidationRecord struct {
	Name  string // DNS record name
	Value string // DNS record value
	Type  string // DNS record type (CNAME, TXT, etc.)
}

// CertificateInfo contains information about a TLS certificate
type CertificateInfo struct {
	ARN               string             // Provider-specific identifier
	Domain            string             // Primary domain
	Status            string             // Certificate status
	ValidationRecords []ValidationRecord // DNS validation records
}

// SubnetInfo contains information about a subnet
type SubnetInfo struct {
	ID               string // Subnet ID
	AvailabilityZone string // Availability zone
	CIDR             string // CIDR block
}

// LoadBalancerProvider defines the interface for load balancer operations
// This allows different cloud providers to implement their own load balancer logic
type LoadBalancerProvider interface {
	// CreateLoadBalancer creates a load balancer with the given configuration
	CreateLoadBalancer(ctx context.Context, installationKey, vpcID string, subnetIDs []string, securityGroupID string) (*LoadBalancerInfo, error)

	// CreateTargetGroup creates a target group for the load balancer
	CreateTargetGroup(ctx context.Context, installationKey, vpcID string) (string, error)

	// RegisterTargets registers instances with the target group
	RegisterTargets(ctx context.Context, targetGroupID string, instanceIDs ...string) error

	// CreateHTTPSListener creates an HTTPS listener with TLS termination
	CreateHTTPSListener(ctx context.Context, loadBalancerID, targetGroupID, certificateID string) (string, error)

	// CreateHTTPRedirectListener creates an HTTP listener that redirects to HTTPS
	CreateHTTPRedirectListener(ctx context.Context, loadBalancerID string) (string, error)

	// WaitForLoadBalancerActive waits for the load balancer to become active
	WaitForLoadBalancerActive(ctx context.Context, loadBalancerID string) error

	// DeleteLoadBalancer deletes a load balancer by installation key
	DeleteLoadBalancer(ctx context.Context, installationKey string) error

	// DeleteTargetGroup deletes a target group by installation key
	DeleteTargetGroup(ctx context.Context, installationKey string) error

	// FindLoadBalancerByInstallationKey finds a load balancer by installation key
	FindLoadBalancerByInstallationKey(ctx context.Context, installationKey string) (string, error)

	// FindTargetGroupByInstallationKey finds a target group by installation key
	FindTargetGroupByInstallationKey(ctx context.Context, installationKey string) (string, error)

	// GetLoadBalancerDNSName gets the DNS name of a load balancer
	GetLoadBalancerDNSName(ctx context.Context, loadBalancerID string) (string, error)
}

// TLSCertificateProvider defines the interface for TLS certificate operations
type TLSCertificateProvider interface {
	// RequestCertificate requests a new TLS certificate for the given domain
	RequestCertificate(ctx context.Context, domain string, installationKey string) (string, error)

	// GetValidationRecords retrieves DNS validation records for a certificate
	GetValidationRecords(ctx context.Context, certificateID string) ([]ValidationRecord, error)

	// WaitForValidation waits for the certificate to be validated
	WaitForValidation(ctx context.Context, certificateID string, timeout time.Duration) error

	// GetCertificateStatus returns the current status of a certificate
	GetCertificateStatus(ctx context.Context, certificateID string) (string, error)

	// DeleteCertificate deletes a TLS certificate
	DeleteCertificate(ctx context.Context, certificateID string) error

	// FindCertificateByInstallationKey finds a certificate by installation key
	FindCertificateByInstallationKey(ctx context.Context, installationKey string) (string, error)
}

// NetworkProvider defines the interface for network operations
type NetworkProvider interface {
	// GetDefaultVPC returns the default VPC ID and CIDR
	GetDefaultVPC(ctx context.Context) (vpcID string, cidr string, err error)

	// GetDefaultSubnet returns a default subnet in the VPC
	GetDefaultSubnet(ctx context.Context, vpcID string) (subnetID string, availabilityZone string, err error)

	// GetAllDefaultSubnets returns all default subnets in the VPC (for ALB which requires multiple AZs)
	GetAllDefaultSubnets(ctx context.Context, vpcID string) ([]SubnetInfo, error)

	// CreateSecurityGroup creates a security group for the instance
	CreateSecurityGroup(ctx context.Context, vpcID, vpcCIDR, installationKey string) (string, error)

	// CreateLoadBalancerSecurityGroup creates a security group for the load balancer
	CreateLoadBalancerSecurityGroup(ctx context.Context, vpcID, installationKey string) (string, error)

	// DeleteSecurityGroup deletes a security group
	DeleteSecurityGroup(ctx context.Context, securityGroupID string) error

	// FindSecurityGroupByInstallationKey finds a security group by installation key
	FindSecurityGroupByInstallationKey(ctx context.Context, installationKey string, isLoadBalancer bool) (string, error)
}

// CloudProvider combines all provider interfaces for a complete cloud provider implementation
type CloudProvider interface {
	LoadBalancerProvider
	TLSCertificateProvider
	NetworkProvider

	// Name returns the provider name (aws, gcp, azure)
	Name() string
}
