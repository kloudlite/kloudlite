package gcp

import (
	"context"
	"fmt"
	"strings"

	compute "cloud.google.com/go/compute/apiv1"
	"cloud.google.com/go/compute/apiv1/computepb"
)

// FirewallRuleNames returns all firewall rule names for an installation
func FirewallRuleNames(installationKey string) []string {
	return []string{
		fmt.Sprintf("kl-%s-fw-lb", installationKey),
		fmt.Sprintf("kl-%s-fw-http", installationKey),
		fmt.Sprintf("kl-%s-fw-k3s", installationKey),
	}
}

// NetworkTag returns the network tag for VMs in this installation
func NetworkTag(installationKey string) string {
	return fmt.Sprintf("kl-%s-vm", installationKey)
}

// EnsureFirewallRules creates all required firewall rules for the installation
func EnsureFirewallRules(ctx context.Context, cfg *GCPConfig, vpcCIDR, installationKey string) error {
	// Create all firewall rules
	if err := CreateLBHealthCheckFirewall(ctx, cfg, installationKey); err != nil {
		return err
	}
	if err := CreateHTTPFirewall(ctx, cfg, installationKey); err != nil {
		return err
	}
	if err := CreateK3sFirewall(ctx, cfg, vpcCIDR, installationKey); err != nil {
		return err
	}
	return nil
}

// CreateLBHealthCheckFirewall allows GCP health check IPs to reach the VM
// GCP health check probes come from specific IP ranges
func CreateLBHealthCheckFirewall(ctx context.Context, cfg *GCPConfig, installationKey string) error {
	firewallsClient, err := compute.NewFirewallsRESTClient(ctx)
	if err != nil {
		return fmt.Errorf("failed to create firewalls client: %w", err)
	}
	defer firewallsClient.Close()

	ruleName := fmt.Sprintf("kl-%s-fw-lb", installationKey)
	networkTag := NetworkTag(installationKey)

	// Check if rule exists
	_, err = firewallsClient.Get(ctx, &computepb.GetFirewallRequest{
		Project:  cfg.Project,
		Firewall: ruleName,
	})
	if err == nil {
		// Rule exists
		return nil
	}

	// GCP health check IP ranges
	// https://cloud.google.com/load-balancing/docs/health-check-concepts#ip-ranges
	healthCheckRanges := []string{
		"35.191.0.0/16",
		"130.211.0.0/22",
	}

	rule := &computepb.Firewall{
		Name:         ptrString(ruleName),
		Description:  ptrString("Allow GCP health check probes"),
		Network:      ptrString(GetNetworkURL(cfg.Project, "default")),
		TargetTags:   []string{networkTag},
		SourceRanges: healthCheckRanges,
		Allowed: []*computepb.Allowed{
			{
				IPProtocol: ptrString("tcp"),
				Ports:      []string{"80"},
			},
		},
		Direction: ptrString("INGRESS"),
		Priority:  ptrInt32(1000),
	}

	op, err := firewallsClient.Insert(ctx, &computepb.InsertFirewallRequest{
		Project:          cfg.Project,
		FirewallResource: rule,
	})
	if err != nil {
		if strings.Contains(err.Error(), "already exists") {
			return nil
		}
		return fmt.Errorf("failed to create LB health check firewall rule: %w", err)
	}

	// Wait for operation to complete
	if err := op.Wait(ctx); err != nil {
		return fmt.Errorf("failed waiting for firewall creation: %w", err)
	}

	return nil
}

// CreateHTTPFirewall allows HTTP/HTTPS traffic from anywhere
func CreateHTTPFirewall(ctx context.Context, cfg *GCPConfig, installationKey string) error {
	firewallsClient, err := compute.NewFirewallsRESTClient(ctx)
	if err != nil {
		return fmt.Errorf("failed to create firewalls client: %w", err)
	}
	defer firewallsClient.Close()

	ruleName := fmt.Sprintf("kl-%s-fw-http", installationKey)
	networkTag := NetworkTag(installationKey)

	// Check if rule exists
	_, err = firewallsClient.Get(ctx, &computepb.GetFirewallRequest{
		Project:  cfg.Project,
		Firewall: ruleName,
	})
	if err == nil {
		// Rule exists
		return nil
	}

	rule := &computepb.Firewall{
		Name:         ptrString(ruleName),
		Description:  ptrString("Allow HTTP/HTTPS traffic from anywhere"),
		Network:      ptrString(GetNetworkURL(cfg.Project, "default")),
		TargetTags:   []string{networkTag},
		SourceRanges: []string{"0.0.0.0/0"},
		Allowed: []*computepb.Allowed{
			{
				IPProtocol: ptrString("tcp"),
				Ports:      []string{"80", "443"},
			},
		},
		Direction: ptrString("INGRESS"),
		Priority:  ptrInt32(1000),
	}

	op, err := firewallsClient.Insert(ctx, &computepb.InsertFirewallRequest{
		Project:          cfg.Project,
		FirewallResource: rule,
	})
	if err != nil {
		if strings.Contains(err.Error(), "already exists") {
			return nil
		}
		return fmt.Errorf("failed to create HTTP firewall rule: %w", err)
	}

	// Wait for operation to complete
	if err := op.Wait(ctx); err != nil {
		return fmt.Errorf("failed waiting for firewall creation: %w", err)
	}

	return nil
}

// CreateK3sFirewall allows K3s internal traffic from VPC CIDR
func CreateK3sFirewall(ctx context.Context, cfg *GCPConfig, vpcCIDR, installationKey string) error {
	firewallsClient, err := compute.NewFirewallsRESTClient(ctx)
	if err != nil {
		return fmt.Errorf("failed to create firewalls client: %w", err)
	}
	defer firewallsClient.Close()

	ruleName := fmt.Sprintf("kl-%s-fw-k3s", installationKey)
	networkTag := NetworkTag(installationKey)

	// Check if rule exists
	_, err = firewallsClient.Get(ctx, &computepb.GetFirewallRequest{
		Project:  cfg.Project,
		Firewall: ruleName,
	})
	if err == nil {
		// Rule exists
		return nil
	}

	rule := &computepb.Firewall{
		Name:         ptrString(ruleName),
		Description:  ptrString("Allow K3s internal traffic from VPC"),
		Network:      ptrString(GetNetworkURL(cfg.Project, "default")),
		TargetTags:   []string{networkTag},
		SourceRanges: []string{vpcCIDR},
		Allowed: []*computepb.Allowed{
			{
				IPProtocol: ptrString("tcp"),
				Ports:      []string{"6443", "10250", "5001"},
			},
			{
				IPProtocol: ptrString("udp"),
				Ports:      []string{"8472"},
			},
		},
		Direction: ptrString("INGRESS"),
		Priority:  ptrInt32(1000),
	}

	op, err := firewallsClient.Insert(ctx, &computepb.InsertFirewallRequest{
		Project:          cfg.Project,
		FirewallResource: rule,
	})
	if err != nil {
		if strings.Contains(err.Error(), "already exists") {
			return nil
		}
		return fmt.Errorf("failed to create K3s firewall rule: %w", err)
	}

	// Wait for operation to complete
	if err := op.Wait(ctx); err != nil {
		return fmt.Errorf("failed waiting for firewall creation: %w", err)
	}

	return nil
}

// DeleteFirewallRules removes all firewall rules for an installation
func DeleteFirewallRules(ctx context.Context, cfg *GCPConfig, installationKey string) error {
	firewallsClient, err := compute.NewFirewallsRESTClient(ctx)
	if err != nil {
		return fmt.Errorf("failed to create firewalls client: %w", err)
	}
	defer firewallsClient.Close()

	ruleNames := FirewallRuleNames(installationKey)

	for _, ruleName := range ruleNames {
		// Check if rule exists
		_, err := firewallsClient.Get(ctx, &computepb.GetFirewallRequest{
			Project:  cfg.Project,
			Firewall: ruleName,
		})
		if err != nil {
			// Rule doesn't exist
			continue
		}

		// Delete the rule
		op, err := firewallsClient.Delete(ctx, &computepb.DeleteFirewallRequest{
			Project:  cfg.Project,
			Firewall: ruleName,
		})
		if err != nil {
			if strings.Contains(err.Error(), "not found") || strings.Contains(err.Error(), "404") {
				continue
			}
			return fmt.Errorf("failed to delete firewall rule %s: %w", ruleName, err)
		}

		// Wait for operation to complete
		if err := op.Wait(ctx); err != nil {
			return fmt.Errorf("failed waiting for firewall deletion: %w", err)
		}
	}

	return nil
}

func ptrInt32(i int32) *int32 {
	return &i
}
