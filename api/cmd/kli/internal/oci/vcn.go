package oci

import (
	"context"
	"fmt"

	"github.com/oracle/oci-go-sdk/v65/core"
)

// Default values for OCI networking
const (
	DefaultVCNCIDR = "10.0.0.0/16"
)

// GetDefaultVCN finds the first VCN in the compartment or returns the default VCN
func GetDefaultVCN(ctx context.Context, cfg *OCIConfig) (string, string, error) {
	vnClient, err := core.NewVirtualNetworkClientWithConfigurationProvider(cfg.ConfigProvider)
	if err != nil {
		return "", "", fmt.Errorf("failed to create virtual network client: %w", err)
	}

	resp, err := vnClient.ListVcns(ctx, core.ListVcnsRequest{
		CompartmentId: &cfg.CompartmentOCID,
	})
	if err != nil {
		return "", "", fmt.Errorf("failed to list VCNs: %w", err)
	}

	// Return the first available VCN
	for _, vcn := range resp.Items {
		if vcn.Id == nil || vcn.LifecycleState != core.VcnLifecycleStateAvailable {
			continue
		}
		cidr := DefaultVCNCIDR
		if len(vcn.CidrBlocks) > 0 {
			cidr = vcn.CidrBlocks[0]
		}
		return *vcn.Id, cidr, nil
	}

	return "", "", fmt.Errorf("no VCN found in compartment %s; please create a VCN first", cfg.CompartmentOCID)
}

// GetDefaultSubnet finds a public subnet in the given VCN
func GetDefaultSubnet(ctx context.Context, cfg *OCIConfig, vcnID string) (string, string, error) {
	vnClient, err := core.NewVirtualNetworkClientWithConfigurationProvider(cfg.ConfigProvider)
	if err != nil {
		return "", "", fmt.Errorf("failed to create virtual network client: %w", err)
	}

	resp, err := vnClient.ListSubnets(ctx, core.ListSubnetsRequest{
		CompartmentId: &cfg.CompartmentOCID,
		VcnId:         &vcnID,
	})
	if err != nil {
		return "", "", fmt.Errorf("failed to list subnets: %w", err)
	}

	// Prefer a public subnet (no prohibitPublicIpOnVnic)
	for _, subnet := range resp.Items {
		if subnet.Id == nil || subnet.LifecycleState != core.SubnetLifecycleStateAvailable {
			continue
		}
		// Public subnets don't prohibit public IPs
		if subnet.ProhibitPublicIpOnVnic != nil && *subnet.ProhibitPublicIpOnVnic {
			continue
		}
		cidr := ""
		if subnet.CidrBlock != nil {
			cidr = *subnet.CidrBlock
		}
		return *subnet.Id, cidr, nil
	}

	// Fall back to any available subnet
	for _, subnet := range resp.Items {
		if subnet.Id == nil || subnet.LifecycleState != core.SubnetLifecycleStateAvailable {
			continue
		}
		cidr := ""
		if subnet.CidrBlock != nil {
			cidr = *subnet.CidrBlock
		}
		return *subnet.Id, cidr, nil
	}

	return "", "", fmt.Errorf("no subnet found in VCN %s; please create a subnet first", vcnID)
}

// GetVCNCIDR returns the CIDR of the VCN
func GetVCNCIDR(ctx context.Context, cfg *OCIConfig, vcnID string) (string, error) {
	vnClient, err := core.NewVirtualNetworkClientWithConfigurationProvider(cfg.ConfigProvider)
	if err != nil {
		return DefaultVCNCIDR, nil
	}

	resp, err := vnClient.GetVcn(ctx, core.GetVcnRequest{
		VcnId: &vcnID,
	})
	if err != nil {
		return DefaultVCNCIDR, nil
	}

	if len(resp.CidrBlocks) > 0 {
		return resp.CidrBlocks[0], nil
	}

	return DefaultVCNCIDR, nil
}
