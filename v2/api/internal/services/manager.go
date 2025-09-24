package services

import (
	"context"
	"fmt"

	"github.com/kloudlite/api/v2/internal/repository"
)

// Manager coordinates all services and provides a unified interface
type Manager struct {
	// Individual services
	Users UserService

	// You can add other services here as needed:
	// Workspaces WorkspaceService
	// Environments EnvironmentService
	// etc.
}

// ManagerOptions contains options for creating a services manager
type ManagerOptions struct {
	// Repository manager
	RepositoryManager *repository.Manager
}

// NewManager creates a new services manager
func NewManager(ctx context.Context, opts *ManagerOptions) (*Manager, error) {
	if opts == nil {
		return nil, fmt.Errorf("ManagerOptions is required")
	}

	if opts.RepositoryManager == nil {
		return nil, fmt.Errorf("RepositoryManager is required")
	}

	// Create individual services
	userService := NewUserService(opts.RepositoryManager.Users)

	return &Manager{
		Users: userService,
	}, nil
}

// HealthCheck performs health checks on all services
func (m *Manager) HealthCheck(ctx context.Context) error {
	// For now, we don't have specific health checks for services
	// but we could add them here if needed
	return nil
}

// Close cleans up resources if needed
func (m *Manager) Close() error {
	// If we need to close any connections or clean up resources,
	// we can do it here
	return nil
}