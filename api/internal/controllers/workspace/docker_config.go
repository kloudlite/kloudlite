package workspace

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	workspacev1 "github.com/kloudlite/kloudlite/api/internal/controllers/workspace/v1"
	"go.uber.org/zap"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

// DockerConfig represents the Docker config.json structure
type DockerConfig struct {
	Auths map[string]DockerAuth `json:"auths"`
}

// DockerAuth represents authentication for a Docker registry
type DockerAuth struct {
	Auth string `json:"auth"`
}

// ensureDockerConfigSecret creates or updates a Secret containing Docker config.json
// Registry runs without authentication - this just creates an empty config
func (r *WorkspaceReconciler) ensureDockerConfigSecret(ctx context.Context, workspace *workspacev1.Workspace, registryHost, targetNamespace string, logger *zap.Logger) error {
	// Build empty Docker config.json - registry runs without auth
	dockerConfig := DockerConfig{
		Auths: map[string]DockerAuth{},
	}

	configJSON, err := json.Marshal(dockerConfig)
	if err != nil {
		return fmt.Errorf("failed to marshal Docker config: %w", err)
	}

	// Create the Secret
	secretName := fmt.Sprintf("%s-docker-config", workspace.Name)
	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      secretName,
			Namespace: targetNamespace,
			Labels: map[string]string{
				"kloudlite.io/workspace-name": workspace.Name,
				"kloudlite.io/docker-config":  "true",
			},
		},
	}

	_, err = controllerutil.CreateOrUpdate(ctx, r.Client, secret, func() error {
		// Set workspace as owner for automatic cleanup
		if err := controllerutil.SetControllerReference(workspace, secret, r.Scheme); err != nil {
			return fmt.Errorf("failed to set owner reference on Docker config Secret: %w", err)
		}

		secret.Type = corev1.SecretTypeOpaque
		secret.Data = map[string][]byte{
			"config.json": configJSON,
		}

		return nil
	})

	if err != nil {
		return fmt.Errorf("failed to create/update Docker config Secret: %w", err)
	}

	logger.Info("Docker config Secret ensured",
		zap.String("secret", secretName),
		zap.String("registryHost", registryHost),
	)

	return nil
}

// getDockerConfigSecretName returns the name of the Docker config Secret for a workspace
func getDockerConfigSecretName(workspaceName string) string {
	return fmt.Sprintf("%s-docker-config", workspaceName)
}

// getImageRegistryHost returns the image registry host from HOSTED_SUBDOMAIN env var
// Returns empty string if registry is not available
func (r *WorkspaceReconciler) getImageRegistryHost(ctx context.Context) string {
	// Get subdomain from HOSTED_SUBDOMAIN env var (e.g., "beanbag.khost.dev")
	hostedSubdomain := os.Getenv("HOSTED_SUBDOMAIN")
	if hostedSubdomain != "" {
		// Use HTTPS endpoint via ingress: cr.{subdomain}
		// subdomain is already full domain like "beanbag.khost.dev"
		return fmt.Sprintf("cr.%s", hostedSubdomain)
	}
	// Fallback to internal service if subdomain not available
	return "image-registry.kloudlite.svc.cluster.local:5000"
}
