package oci

import (
	"context"
	"fmt"
	"strings"

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

// EnsureSubnetSecurityListRules ensures the subnet's security list has the required
// ingress/egress rules for NLB health checks and traffic. OCI NLBs require the
// subnet's security list (not just the NSG) to allow health check traffic to backends.
func EnsureSubnetSecurityListRules(ctx context.Context, cfg *OCIConfig, subnetID string) error {
	vnClient, err := core.NewVirtualNetworkClientWithConfigurationProvider(cfg.ConfigProvider)
	if err != nil {
		return fmt.Errorf("failed to create virtual network client: %w", err)
	}

	// Get the subnet to find its security list
	subnetResp, err := vnClient.GetSubnet(ctx, core.GetSubnetRequest{
		SubnetId: &subnetID,
	})
	if err != nil {
		return fmt.Errorf("failed to get subnet: %w", err)
	}

	if len(subnetResp.SecurityListIds) == 0 {
		return fmt.Errorf("subnet has no security lists")
	}

	slID := subnetResp.SecurityListIds[0]

	// Get current security list
	slResp, err := vnClient.GetSecurityList(ctx, core.GetSecurityListRequest{
		SecurityListId: &slID,
	})
	if err != nil {
		return fmt.Errorf("failed to get security list: %w", err)
	}

	// Check if rules already exist (look for HTTP port 80)
	for _, rule := range slResp.IngressSecurityRules {
		if rule.TcpOptions != nil && rule.TcpOptions.DestinationPortRange != nil &&
			rule.TcpOptions.DestinationPortRange.Min != nil && *rule.TcpOptions.DestinationPortRange.Min == 80 {
			return nil // Already has port 80 rule
		}
	}

	// Build ingress rules — ports needed for NLB health checks and traffic
	requiredPorts := []struct {
		port        int
		description string
	}{
		{22, "Allow SSH"},
		{80, "Allow HTTP"},
		{443, "Allow HTTPS"},
		{6443, "Allow K3s API"},
	}

	ingressRules := slResp.IngressSecurityRules
	for _, p := range requiredPorts {
		port := p.port
		desc := p.description
		ingressRules = append(ingressRules, core.IngressSecurityRule{
			Protocol:   strPtr("6"), // TCP
			Source:     strPtr("0.0.0.0/0"),
			SourceType: core.IngressSecurityRuleSourceTypeCidrBlock,
			TcpOptions: &core.TcpOptions{
				DestinationPortRange: &core.PortRange{
					Min: &port,
					Max: &port,
				},
			},
			Description: &desc,
		})
	}

	// Ensure egress rule exists
	egressRules := slResp.EgressSecurityRules
	hasEgress := false
	for _, rule := range egressRules {
		if rule.Destination != nil && *rule.Destination == "0.0.0.0/0" && *rule.Protocol == "all" {
			hasEgress = true
			break
		}
	}
	if !hasEgress {
		egressDesc := "Allow all egress"
		egressRules = append(egressRules, core.EgressSecurityRule{
			Protocol:        strPtr("all"),
			Destination:     strPtr("0.0.0.0/0"),
			DestinationType: core.EgressSecurityRuleDestinationTypeCidrBlock,
			Description:     &egressDesc,
		})
	}

	_, err = vnClient.UpdateSecurityList(ctx, core.UpdateSecurityListRequest{
		SecurityListId: &slID,
		UpdateSecurityListDetails: core.UpdateSecurityListDetails{
			IngressSecurityRules: ingressRules,
			EgressSecurityRules:  egressRules,
		},
	})
	if err != nil {
		if strings.Contains(err.Error(), "already exists") || strings.Contains(err.Error(), "Conflict") {
			return nil
		}
		return fmt.Errorf("failed to update security list: %w", err)
	}

	return nil
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
