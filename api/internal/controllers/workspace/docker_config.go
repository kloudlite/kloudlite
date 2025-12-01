package workspace

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"

	domainrequestv1 "github.com/kloudlite/kloudlite/api/internal/controllers/domainrequest/v1"
	workspacev1 "github.com/kloudlite/kloudlite/api/internal/controllers/workspace/v1"
	fn "github.com/kloudlite/kloudlite/api/pkg/operator-toolkit/functions"
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
// for the workspace's image registry authentication
func (r *WorkspaceReconciler) ensureDockerConfigSecret(ctx context.Context, workspace *workspacev1.Workspace, registryHost, targetNamespace string, logger *zap.Logger) error {
	if r.JWTSecret == "" {
		logger.Warn("JWTSecret not configured, skipping Docker config creation")
		return nil
	}

	// Generate a long-lived JWT token for the workspace owner
	// Token expires in 8760 hours (1 year) for long-lived workspace credentials
	token, err := r.generateDockerRegistryToken(workspace.Spec.OwnedBy, 8760)
	if err != nil {
		return fmt.Errorf("failed to generate Docker registry token: %w", err)
	}

	// Create Docker auth string: base64(username:password)
	// Username is the workspace owner, password is the JWT token
	authString := base64.StdEncoding.EncodeToString(
		[]byte(fmt.Sprintf("%s:%s", workspace.Spec.OwnedBy, token)),
	)

	// Build Docker config.json
	dockerConfig := DockerConfig{
		Auths: map[string]DockerAuth{
			registryHost: {
				Auth: authString,
			},
		},
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

// getImageRegistryHost returns the image registry host from DomainRequest
// Returns empty string if registry is not available
func (r *WorkspaceReconciler) getImageRegistryHost(ctx context.Context) string {
	domainRequest := &domainrequestv1.DomainRequest{}
	if err := r.Get(ctx, fn.NN("", "installation-domain"), domainRequest); err == nil && domainRequest.Status.Subdomain != "" {
		// Use HTTPS endpoint via ingress: cr.{subdomain}
		// subdomain is already full domain like "beanbag.khost.dev"
		return fmt.Sprintf("cr.%s", domainRequest.Status.Subdomain)
	}
	// Fallback to internal service if subdomain not available
	return "image-registry.kloudlite.svc.cluster.local:5000"
}
