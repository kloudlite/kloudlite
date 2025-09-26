package k8s

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	platformv1alpha1 "github.com/kloudlite/kloudlite/v2/api/pkg/apis/platform/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// Client wraps both the standard Kubernetes client and the controller-runtime client
type Client struct {
	// Standard Kubernetes client for native resources
	Clientset kubernetes.Interface

	// Controller-runtime client for custom resources and unified API
	RuntimeClient client.Client

	// REST config
	Config *rest.Config
}

// ClientOptions contains options for creating a Kubernetes client
type ClientOptions struct {
	// Kubeconfig path (optional, will try to auto-detect)
	KubeconfigPath string

	// Context to use from kubeconfig (optional, uses current context)
	Context string

	// Master URL (optional, for in-cluster config override)
	MasterURL string
}

// NewClient creates a new Kubernetes client with support for both standard and custom resources
func NewClient(ctx context.Context, opts *ClientOptions) (*Client, error) {
	if opts == nil {
		opts = &ClientOptions{}
	}

	// Get REST config
	config, err := getRestConfig(opts)
	if err != nil {
		return nil, fmt.Errorf("failed to get REST config: %w", err)
	}

	// Create standard Kubernetes clientset
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create Kubernetes clientset: %w", err)
	}

	// Create runtime scheme with our custom resources
	scheme := runtime.NewScheme()
	if err := platformv1alpha1.AddToScheme(scheme); err != nil {
		return nil, fmt.Errorf("failed to add platform scheme: %w", err)
	}

	// Create controller-runtime client
	runtimeClient, err := client.New(config, client.Options{
		Scheme: scheme,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create controller-runtime client: %w", err)
	}

	return &Client{
		Clientset:     clientset,
		RuntimeClient: runtimeClient,
		Config:        config,
	}, nil
}

// getRestConfig creates a REST config from various sources
func getRestConfig(opts *ClientOptions) (*rest.Config, error) {
	// Try in-cluster config first
	if config, err := rest.InClusterConfig(); err == nil {
		// If master URL is provided, override it
		if opts.MasterURL != "" {
			config.Host = opts.MasterURL
		}
		return config, nil
	}

	// Try kubeconfig file
	kubeconfigPath := opts.KubeconfigPath
	if kubeconfigPath == "" {
		// Try default locations
		if home := os.Getenv("HOME"); home != "" {
			kubeconfigPath = filepath.Join(home, ".kube", "config")
		} else if kubeconfig := os.Getenv("KUBECONFIG"); kubeconfig != "" {
			kubeconfigPath = kubeconfig
		} else {
			// Try the kubeconfig directory in the project (for local development)
			if wd, err := os.Getwd(); err == nil {
				projectKubeconfig := filepath.Join(wd, "kubeconfig", "config.yaml")
				if _, err := os.Stat(projectKubeconfig); err == nil {
					kubeconfigPath = projectKubeconfig
				}
			}
		}
	}

	if kubeconfigPath == "" || !fileExists(kubeconfigPath) {
		return nil, fmt.Errorf("no valid kubeconfig found")
	}

	// Load kubeconfig
	config, err := clientcmd.LoadFromFile(kubeconfigPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load kubeconfig from %s: %w", kubeconfigPath, err)
	}

	// Use specific context if provided
	if opts.Context != "" {
		config.CurrentContext = opts.Context
	}

	// Build REST config
	clientConfig := clientcmd.NewDefaultClientConfig(*config, &clientcmd.ConfigOverrides{})
	restConfig, err := clientConfig.ClientConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to build REST config: %w", err)
	}

	// Override master URL if provided
	if opts.MasterURL != "" {
		restConfig.Host = opts.MasterURL
	}

	return restConfig, nil
}

// fileExists checks if a file exists
func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

// HealthCheck performs a basic health check on the Kubernetes connection
func (c *Client) HealthCheck(ctx context.Context) error {
	// Check if we can reach the API server
	_, err := c.Clientset.Discovery().ServerVersion()
	if err != nil {
		return fmt.Errorf("failed to connect to Kubernetes API server: %w", err)
	}

	// Try to list namespaces as a basic permission check
	_, err = c.Clientset.CoreV1().Namespaces().List(ctx, metav1.ListOptions{})
	if err != nil {
		return fmt.Errorf("failed to list namespaces: %w", err)
	}

	return nil
}

// GetNamespace returns the default namespace for operations
func (c *Client) GetNamespace() string {
	// Try to get namespace from service account (in-cluster)
	if ns, err := os.ReadFile("/var/run/secrets/kubernetes.io/serviceaccount/namespace"); err == nil {
		return string(ns)
	}

	// Default to "default" namespace
	return "default"
}