package gcp

import (
	"context"
	"fmt"
	"strings"
	"time"

	"cloud.google.com/go/iam/apiv1/iampb"
	resourcemanager "cloud.google.com/go/resourcemanager/apiv3"
	iam "google.golang.org/api/iam/v1"
)

// EnsureServiceAccount creates a service account if it doesn't exist
// Returns the service account email
func EnsureServiceAccount(ctx context.Context, cfg *GCPConfig, installationKey string) (string, error) {
	iamService, err := iam.NewService(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to create IAM service: %w", err)
	}

	saName := fmt.Sprintf("kl-%s-sa", installationKey)
	// Service account names can be max 30 chars, so truncate if needed
	if len(saName) > 30 {
		saName = saName[:30]
	}
	saEmail := fmt.Sprintf("%s@%s.iam.gserviceaccount.com", saName, cfg.Project)

	// Check if service account exists
	_, err = iamService.Projects.ServiceAccounts.Get(
		fmt.Sprintf("projects/%s/serviceAccounts/%s", cfg.Project, saEmail),
	).Context(ctx).Do()
	if err == nil {
		// Service account already exists
		return saEmail, nil
	}

	// Create service account
	sa, err := iamService.Projects.ServiceAccounts.Create(
		fmt.Sprintf("projects/%s", cfg.Project),
		&iam.CreateServiceAccountRequest{
			AccountId: saName,
			ServiceAccount: &iam.ServiceAccount{
				DisplayName: fmt.Sprintf("Kloudlite Installation %s", installationKey),
				Description: "Service account for Kloudlite K3s cluster",
			},
		},
	).Context(ctx).Do()
	if err != nil {
		return "", fmt.Errorf("failed to create service account: %w", err)
	}

	return sa.Email, nil
}

// GrantIAMRoles grants required IAM roles to the service account
func GrantIAMRoles(ctx context.Context, cfg *GCPConfig, saEmail, installationKey string) error {
	projectsClient, err := resourcemanager.NewProjectsClient(ctx)
	if err != nil {
		return fmt.Errorf("failed to create resource manager client: %w", err)
	}
	defer projectsClient.Close()

	projectName := fmt.Sprintf("projects/%s", cfg.Project)

	// Get current IAM policy
	policy, err := projectsClient.GetIamPolicy(ctx, &iampb.GetIamPolicyRequest{
		Resource: projectName,
	})
	if err != nil {
		return fmt.Errorf("failed to get IAM policy: %w", err)
	}

	// Roles needed for Kloudlite operations
	// - compute.instanceAdmin.v1: Create/manage VMs
	// - storage.objectAdmin: Backup K3s database to GCS
	// - iam.serviceAccountUser: Use service account
	requiredRoles := []string{
		"roles/compute.instanceAdmin.v1",
		"roles/storage.objectAdmin",
		"roles/iam.serviceAccountUser",
	}

	member := fmt.Sprintf("serviceAccount:%s", saEmail)

	// Add bindings for each role
	for _, role := range requiredRoles {
		found := false
		for _, binding := range policy.Bindings {
			if binding.Role == role {
				// Check if member already has this role
				for _, m := range binding.Members {
					if m == member {
						found = true
						break
					}
				}
				if !found {
					binding.Members = append(binding.Members, member)
					found = true
				}
				break
			}
		}
		if !found {
			// Add new binding
			policy.Bindings = append(policy.Bindings, &iampb.Binding{
				Role:    role,
				Members: []string{member},
			})
		}
	}

	// Set updated policy
	_, err = projectsClient.SetIamPolicy(ctx, &iampb.SetIamPolicyRequest{
		Resource: projectName,
		Policy:   policy,
	})
	if err != nil {
		return fmt.Errorf("failed to set IAM policy: %w", err)
	}

	// Wait for IAM to propagate
	time.Sleep(10 * time.Second)

	return nil
}

// DeleteServiceAccount removes the service account
func DeleteServiceAccount(ctx context.Context, cfg *GCPConfig, installationKey string) error {
	iamService, err := iam.NewService(ctx)
	if err != nil {
		return fmt.Errorf("failed to create IAM service: %w", err)
	}

	saName := fmt.Sprintf("kl-%s-sa", installationKey)
	if len(saName) > 30 {
		saName = saName[:30]
	}
	saEmail := fmt.Sprintf("%s@%s.iam.gserviceaccount.com", saName, cfg.Project)

	// Delete service account
	_, err = iamService.Projects.ServiceAccounts.Delete(
		fmt.Sprintf("projects/%s/serviceAccounts/%s", cfg.Project, saEmail),
	).Context(ctx).Do()
	if err != nil {
		// Check if already deleted
		if strings.Contains(err.Error(), "404") || strings.Contains(err.Error(), "not found") {
			return nil
		}
		return fmt.Errorf("failed to delete service account: %w", err)
	}

	return nil
}

// RemoveIAMRoles removes IAM role bindings for the service account
func RemoveIAMRoles(ctx context.Context, cfg *GCPConfig, saEmail string) error {
	projectsClient, err := resourcemanager.NewProjectsClient(ctx)
	if err != nil {
		return fmt.Errorf("failed to create resource manager client: %w", err)
	}
	defer projectsClient.Close()

	projectName := fmt.Sprintf("projects/%s", cfg.Project)

	// Get current IAM policy
	policy, err := projectsClient.GetIamPolicy(ctx, &iampb.GetIamPolicyRequest{
		Resource: projectName,
	})
	if err != nil {
		return fmt.Errorf("failed to get IAM policy: %w", err)
	}

	member := fmt.Sprintf("serviceAccount:%s", saEmail)

	// Remove member from all bindings
	for _, binding := range policy.Bindings {
		newMembers := make([]string, 0, len(binding.Members))
		for _, m := range binding.Members {
			if m != member {
				newMembers = append(newMembers, m)
			}
		}
		binding.Members = newMembers
	}

	// Set updated policy
	_, err = projectsClient.SetIamPolicy(ctx, &iampb.SetIamPolicyRequest{
		Resource: projectName,
		Policy:   policy,
	})
	if err != nil {
		return fmt.Errorf("failed to set IAM policy: %w", err)
	}

	return nil
}

// GetServiceAccountEmail returns the email for a service account
func GetServiceAccountEmail(cfg *GCPConfig, installationKey string) string {
	saName := fmt.Sprintf("kl-%s-sa", installationKey)
	if len(saName) > 30 {
		saName = saName[:30]
	}
	return fmt.Sprintf("%s@%s.iam.gserviceaccount.com", saName, cfg.Project)
}
