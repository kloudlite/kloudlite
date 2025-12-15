package gcp

import (
	"context"

	compute "cloud.google.com/go/compute/apiv1"
	"cloud.google.com/go/compute/apiv1/computepb"
)

// Default values for GCP networking
const (
	DefaultNetworkName = "default"
	DefaultSubnetName  = "default"
	// GCP auto-mode VPC uses 10.128.0.0/9 range for all regions
	DefaultVPCCIDR = "10.128.0.0/9"
)

// GetDefaultVPC returns the default VPC network name and CIDR
// In GCP, the default network is named "default"
func GetDefaultVPC(ctx context.Context, cfg *GCPConfig) (string, string, error) {
	networksClient, err := compute.NewNetworksRESTClient(ctx)
	if err != nil {
		// Fall back to default network name
		return DefaultNetworkName, "", nil
	}
	defer networksClient.Close()

	req := &computepb.GetNetworkRequest{
		Project: cfg.Project,
		Network: "default",
	}

	network, err := networksClient.Get(ctx, req)
	if err != nil {
		// Fall back to default network name
		return DefaultNetworkName, "", nil
	}

	// GCP default network uses auto subnets, so CIDR is per-region
	// Return the network name; CIDR will be determined from subnet
	return *network.Name, "", nil
}

// GetDefaultSubnet returns the default subnet in the specified region
func GetDefaultSubnet(ctx context.Context, cfg *GCPConfig, networkName string) (string, string, error) {
	subnetworksClient, err := compute.NewSubnetworksRESTClient(ctx)
	if err != nil {
		// Fall back to defaults
		return DefaultSubnetName, DefaultVPCCIDR, nil
	}
	defer subnetworksClient.Close()

	// In the default network with auto subnets, subnet name is "default"
	req := &computepb.GetSubnetworkRequest{
		Project:    cfg.Project,
		Region:     cfg.Region,
		Subnetwork: "default",
	}

	subnet, err := subnetworksClient.Get(ctx, req)
	if err != nil {
		// Fall back to defaults
		return DefaultSubnetName, DefaultVPCCIDR, nil
	}

	cidr := DefaultVPCCIDR
	if subnet.IpCidrRange != nil {
		cidr = *subnet.IpCidrRange
	}

	return *subnet.Name, cidr, nil
}

// GetVPCCIDR returns the CIDR of the subnet in the default network for the region
func GetVPCCIDR(ctx context.Context, cfg *GCPConfig) (string, error) {
	_, cidr, err := GetDefaultSubnet(ctx, cfg, "default")
	if err != nil {
		return DefaultVPCCIDR, nil
	}
	return cidr, nil
}
