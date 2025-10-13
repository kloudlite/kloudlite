package workspace

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	workspacesv1 "github.com/kloudlite/kloudlite/api/pkg/apis/workspaces/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// Client wraps the Kubernetes client for workspace operations
type Client struct {
	K8sClient client.Client
	Namespace string
	Name      string
}

// New creates a new workspace client
func New() (*Client, error) {
	// Register the workspace API types with the scheme
	if err := workspacesv1.AddToScheme(scheme.Scheme); err != nil {
		return nil, fmt.Errorf("failed to add workspace types to scheme: %w", err)
	}

	// Get workspace name and namespace from environment variables
	workspaceName := os.Getenv("WORKSPACE_NAME")
	workspaceNamespace := os.Getenv("WORKSPACE_NAMESPACE")

	if workspaceName == "" {
		workspaceName = os.Getenv("HOSTNAME")
	}
	if workspaceNamespace == "" {
		workspaceNamespace = "default"
	}

	// Create Kubernetes config
	config, err := getKubeConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to get kubeconfig: %w", err)
	}

	// Create controller-runtime client
	k8sClient, err := client.New(config, client.Options{
		Scheme: scheme.Scheme,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create kubernetes client: %w", err)
	}

	return &Client{
		K8sClient: k8sClient,
		Namespace: workspaceNamespace,
		Name:      workspaceName,
	}, nil
}

// Get retrieves the current workspace
func (c *Client) Get(ctx context.Context) (*workspacesv1.Workspace, error) {
	workspace := &workspacesv1.Workspace{}
	err := c.K8sClient.Get(ctx, types.NamespacedName{
		Name:      c.Name,
		Namespace: c.Namespace,
	}, workspace)
	if err != nil {
		return nil, fmt.Errorf("failed to get workspace: %w", err)
	}
	return workspace, nil
}

// Update updates the workspace
func (c *Client) Update(ctx context.Context, workspace *workspacesv1.Workspace) error {
	if err := c.K8sClient.Update(ctx, workspace); err != nil {
		return fmt.Errorf("failed to update workspace: %w", err)
	}
	return nil
}

// Patch patches the workspace
func (c *Client) Patch(ctx context.Context, workspace *workspacesv1.Workspace, patch client.Patch) error {
	if err := c.K8sClient.Patch(ctx, workspace, patch); err != nil {
		return fmt.Errorf("failed to patch workspace: %w", err)
	}
	return nil
}

// getKubeConfig returns the Kubernetes config
func getKubeConfig() (*rest.Config, error) {
	// Try in-cluster config first
	config, err := rest.InClusterConfig()
	if err == nil {
		return config, nil
	}

	// Fall back to kubeconfig file
	kubeconfig := os.Getenv("KUBECONFIG")
	if kubeconfig == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return nil, err
		}
		kubeconfig = filepath.Join(home, ".kube", "config")
	}

	return clientcmd.BuildConfigFromFlags("", kubeconfig)
}
