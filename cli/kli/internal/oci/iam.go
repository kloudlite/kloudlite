package oci

import (
	"context"
	"fmt"
	"strings"

	"github.com/oracle/oci-go-sdk/v65/identity"
)

// DynamicGroupName returns the dynamic group name for an installation
func DynamicGroupName(installationKey string) string {
	return fmt.Sprintf("kl-%s-dg", shortKey(installationKey))
}

// PolicyName returns the policy name for an installation
func PolicyName(installationKey string) string {
	return fmt.Sprintf("kl-%s-policy", shortKey(installationKey))
}

// EnsureDynamicGroup creates a dynamic group matching instances by compartment + tag
func EnsureDynamicGroup(ctx context.Context, cfg *OCIConfig, installationKey string) (string, error) {
	client, err := identity.NewIdentityClientWithConfigurationProvider(cfg.ConfigProvider)
	if err != nil {
		return "", fmt.Errorf("failed to create identity client: %w", err)
	}

	dgName := DynamicGroupName(installationKey)

	// Check if dynamic group already exists
	existingID, err := findDynamicGroupByName(ctx, client, cfg.TenancyOCID, dgName)
	if err == nil && existingID != "" {
		return existingID, nil
	}

	// Matching rule: all instances in the compartment
	// Note: OCI dynamic groups cannot match on freeform tags, only defined tags
	matchingRule := fmt.Sprintf(
		"All {instance.compartment.id = '%s'}",
		cfg.CompartmentOCID,
	)

	description := fmt.Sprintf("Kloudlite installation %s dynamic group", installationKey)
	resp, err := client.CreateDynamicGroup(ctx, identity.CreateDynamicGroupRequest{
		CreateDynamicGroupDetails: identity.CreateDynamicGroupDetails{
			CompartmentId: &cfg.TenancyOCID, // Dynamic groups are tenancy-level
			Name:          &dgName,
			Description:   &description,
			MatchingRule:  &matchingRule,
			FreeformTags:  freeformTags(installationKey),
		},
	})
	if err != nil {
		if strings.Contains(err.Error(), "already exists") || strings.Contains(err.Error(), "Conflict") {
			existingID, findErr := findDynamicGroupByName(ctx, client, cfg.TenancyOCID, dgName)
			if findErr == nil && existingID != "" {
				return existingID, nil
			}
		}
		return "", fmt.Errorf("failed to create dynamic group: %w", err)
	}

	return *resp.Id, nil
}

// EnsurePolicy creates an IAM policy granting object storage access to the dynamic group
func EnsurePolicy(ctx context.Context, cfg *OCIConfig, installationKey string) (string, error) {
	client, err := identity.NewIdentityClientWithConfigurationProvider(cfg.ConfigProvider)
	if err != nil {
		return "", fmt.Errorf("failed to create identity client: %w", err)
	}

	policyName := PolicyName(installationKey)
	dgName := DynamicGroupName(installationKey)

	// Check if policy already exists
	existingID, err := findPolicyByName(ctx, client, cfg.CompartmentOCID, policyName)
	if err == nil && existingID != "" {
		return existingID, nil
	}

	description := fmt.Sprintf("Kloudlite installation %s policy", installationKey)
	// Use OCI resource family types for broader, correct permissions
	// instance-family: instances, instance-images, instance-agent-plugins
	// virtual-network-family: vcns, subnets, vnics, security-lists, network-security-groups, route-tables
	// volume-family: volumes, boot-volumes, volume-attachments, boot-volume-attachments
	// object-family: objects, buckets
	compartmentClause := fmt.Sprintf("in compartment id %s", cfg.CompartmentOCID)
	// If compartment is the tenancy itself, use "in tenancy" syntax
	if cfg.CompartmentOCID == cfg.TenancyOCID {
		compartmentClause = "in tenancy"
	}

	statements := []string{
		fmt.Sprintf("Allow dynamic-group %s to manage object-family %s", dgName, compartmentClause),
		fmt.Sprintf("Allow dynamic-group %s to manage instance-family %s", dgName, compartmentClause),
		fmt.Sprintf("Allow dynamic-group %s to use virtual-network-family %s", dgName, compartmentClause),
		fmt.Sprintf("Allow dynamic-group %s to manage volume-family %s", dgName, compartmentClause),
	}

	resp, err := client.CreatePolicy(ctx, identity.CreatePolicyRequest{
		CreatePolicyDetails: identity.CreatePolicyDetails{
			CompartmentId: &cfg.CompartmentOCID,
			Name:          &policyName,
			Description:   &description,
			Statements:    statements,
			FreeformTags:  freeformTags(installationKey),
		},
	})
	if err != nil {
		if strings.Contains(err.Error(), "already exists") || strings.Contains(err.Error(), "Conflict") {
			existingID, findErr := findPolicyByName(ctx, client, cfg.CompartmentOCID, policyName)
			if findErr == nil && existingID != "" {
				return existingID, nil
			}
		}
		return "", fmt.Errorf("failed to create policy: %w", err)
	}

	return *resp.Id, nil
}

// DeleteDynamicGroup removes a dynamic group
func DeleteDynamicGroup(ctx context.Context, cfg *OCIConfig, installationKey string) error {
	client, err := identity.NewIdentityClientWithConfigurationProvider(cfg.ConfigProvider)
	if err != nil {
		return fmt.Errorf("failed to create identity client: %w", err)
	}

	dgName := DynamicGroupName(installationKey)
	dgID, err := findDynamicGroupByName(ctx, client, cfg.TenancyOCID, dgName)
	if err != nil || dgID == "" {
		return nil // Already gone
	}

	_, err = client.DeleteDynamicGroup(ctx, identity.DeleteDynamicGroupRequest{
		DynamicGroupId: &dgID,
	})
	if err != nil {
		if strings.Contains(err.Error(), "404") || strings.Contains(err.Error(), "not found") {
			return nil
		}
		return fmt.Errorf("failed to delete dynamic group: %w", err)
	}

	return nil
}

// DeletePolicy removes an IAM policy
func DeletePolicy(ctx context.Context, cfg *OCIConfig, installationKey string) error {
	client, err := identity.NewIdentityClientWithConfigurationProvider(cfg.ConfigProvider)
	if err != nil {
		return fmt.Errorf("failed to create identity client: %w", err)
	}

	policyName := PolicyName(installationKey)
	policyID, err := findPolicyByName(ctx, client, cfg.CompartmentOCID, policyName)
	if err != nil || policyID == "" {
		return nil // Already gone
	}

	_, err = client.DeletePolicy(ctx, identity.DeletePolicyRequest{
		PolicyId: &policyID,
	})
	if err != nil {
		if strings.Contains(err.Error(), "404") || strings.Contains(err.Error(), "not found") {
			return nil
		}
		return fmt.Errorf("failed to delete policy: %w", err)
	}

	return nil
}

// findDynamicGroupByName finds a dynamic group by name
func findDynamicGroupByName(ctx context.Context, client identity.IdentityClient, tenancyID, name string) (string, error) {
	resp, err := client.ListDynamicGroups(ctx, identity.ListDynamicGroupsRequest{
		CompartmentId: &tenancyID,
	})
	if err != nil {
		return "", err
	}

	for _, dg := range resp.Items {
		if dg.Name != nil && *dg.Name == name && dg.LifecycleState == identity.DynamicGroupLifecycleStateActive {
			return *dg.Id, nil
		}
	}

	return "", fmt.Errorf("dynamic group %s not found", name)
}

// findPolicyByName finds a policy by name
func findPolicyByName(ctx context.Context, client identity.IdentityClient, compartmentID, name string) (string, error) {
	resp, err := client.ListPolicies(ctx, identity.ListPoliciesRequest{
		CompartmentId: &compartmentID,
	})
	if err != nil {
		return "", err
	}

	for _, p := range resp.Items {
		if p.Name != nil && *p.Name == name && p.LifecycleState == identity.PolicyLifecycleStateActive {
			return *p.Id, nil
		}
	}

	return "", fmt.Errorf("policy %s not found", name)
}
