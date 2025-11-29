package workspace

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	environmentsv1 "github.com/kloudlite/kloudlite/api/internal/controllers/environment/v1"
	workspacesv1 "github.com/kloudlite/kloudlite/api/internal/controllers/workspace/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// Client wraps the Kubernetes client for workspace operations
type Client struct {
	K8sClient client.Client
	Clientset *kubernetes.Clientset
	Namespace string
	Name      string
}

// New creates a new workspace client
func New() (*Client, error) {
	// Register the workspace API types with the scheme
	if err := workspacesv1.AddToScheme(scheme.Scheme); err != nil {
		return nil, fmt.Errorf("failed to add workspace types to scheme: %w", err)
	}

	// Register the environment API types with the scheme
	if err := environmentsv1.AddToScheme(scheme.Scheme); err != nil {
		return nil, fmt.Errorf("failed to add environment types to scheme: %w", err)
	}

	// Get workspace name and namespace from environment variables
	workspaceName := os.Getenv("WORKSPACE_NAME")
	workspaceNamespace := os.Getenv("WORKSPACE_NAMESPACE")

	if workspaceName == "" {
		workspaceName = os.Getenv("HOSTNAME")
	}
	if workspaceNamespace == "" {
		// Try to read namespace from serviceaccount (in-cluster)
		namespaceBytes, err := os.ReadFile("/var/run/secrets/kubernetes.io/serviceaccount/namespace")
		if err == nil {
			workspaceNamespace = strings.TrimSpace(string(namespaceBytes))
		} else {
			workspaceNamespace = "default"
		}
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

	// Create kubernetes clientset for pod logs
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create kubernetes clientset: %w", err)
	}

	return &Client{
		K8sClient: k8sClient,
		Clientset: clientset,
		Namespace: workspaceNamespace,
		Name:      workspaceName,
	}, nil
}

// Get retrieves the current workspace (namespaced resource)
func (c *Client) Get(ctx context.Context) (*workspacesv1.Workspace, error) {
	workspace := &workspacesv1.Workspace{}
	// Workspace is now namespaced, so we need both Name and Namespace
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

	// Store the in-cluster config error for better debugging
	inClusterErr := err

	// Fall back to kubeconfig file
	kubeconfig := os.Getenv("KUBECONFIG")
	if kubeconfig == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return nil, fmt.Errorf("in-cluster config failed (%v), and failed to get home directory: %w", inClusterErr, err)
		}
		kubeconfig = filepath.Join(home, ".kube", "config")
	}

	config, err = clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		return nil, fmt.Errorf("in-cluster config failed (%v), and kubeconfig failed: %w", inClusterErr, err)
	}

	return config, nil
}

// StreamHostManagerLogs streams logs from the host-manager pod, filtering for lines with the given tag.
// The tag is typically the nixpkgs commit hash used during package installation.
// onLine is called for each matching line with the tag stripped.
// Returns when context is cancelled or an error occurs.
func (c *Client) StreamHostManagerLogs(ctx context.Context, namespace string, tag string, onLine func(line string)) error {
	podName := "host-manager-0"
	containerName := "host-manager"

	// Request pod logs with follow enabled
	req := c.Clientset.CoreV1().Pods(namespace).GetLogs(podName, &corev1.PodLogOptions{
		Container: containerName,
		Follow:    true,
		TailLines: func() *int64 { v := int64(0); return &v }(), // Start from current position
	})

	stream, err := req.Stream(ctx)
	if err != nil {
		return fmt.Errorf("failed to open log stream: %w", err)
	}
	defer stream.Close()

	// Tag prefix to filter for
	tagPrefix := fmt.Sprintf("[pkg:%s]", tag)

	// Read and filter lines
	reader := bufio.NewReader(stream)
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		line, err := reader.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				return nil
			}
			return fmt.Errorf("error reading log stream: %w", err)
		}

		// Check if line contains our tag
		if strings.Contains(line, tagPrefix) {
			// Strip the tag prefix and call the callback
			cleanLine := strings.TrimPrefix(line, tagPrefix)
			cleanLine = strings.TrimSpace(cleanLine)
			if cleanLine != "" {
				onLine(cleanLine)
			}
		}
	}
}
