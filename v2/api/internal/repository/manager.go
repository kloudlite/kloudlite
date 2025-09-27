package repository

import (
	"context"

	"github.com/kloudlite/kloudlite/v2/api/internal/k8s"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// Manager coordinates all repositories and provides a unified interface
type Manager struct {
	// K8s client
	k8sClient client.Client

	// Individual repositories
	Users UserRepository
}

// ManagerOptions contains options for creating a repository manager
type ManagerOptions struct {
	// K8s client (if nil, will create a new one)
	K8sClient client.Client

	// K8s client options (only used if K8sClient is nil)
	K8sClientOptions *k8s.ClientOptions
}

// NewManager creates a new repository manager
func NewManager(ctx context.Context, opts *ManagerOptions) (*Manager, error) {
	if opts == nil {
		opts = &ManagerOptions{}
	}

	var k8sClient client.Client

	// Create or use existing k8s client
	if opts.K8sClient != nil {
		k8sClient = opts.K8sClient
	} else {
		// Create new k8s client
		client, err := k8s.NewClient(ctx, opts.K8sClientOptions)
		if err != nil {
			return nil, err
		}
		k8sClient = client.RuntimeClient
	}

	// Create individual repositories
	users := NewUserRepository(k8sClient)

	return &Manager{
		k8sClient: k8sClient,
		Users:     users,
	}, nil
}

// HealthCheck performs health checks on all repositories
func (m *Manager) HealthCheck(ctx context.Context) error {
	// Basic connectivity check by attempting to list namespaces
	// This uses the underlying k8s client to verify connectivity

	// Try to list a few resources to ensure the connection is working
	_, err := m.Users.List(ctx, "", WithLimit(1))
	if err != nil {
		return ErrConnection("failed to connect to user repository", err)
	}

	return nil
}

// Close cleans up resources if needed
func (m *Manager) Close() error {
	// If we need to close any connections or clean up resources,
	// we can do it here. For now, the k8s client doesn't need explicit cleanup.
	return nil
}