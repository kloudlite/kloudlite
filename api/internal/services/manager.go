package services

import (
	"context"
	"fmt"
	"time"

	"github.com/kloudlite/kloudlite/api/internal/config"
	"github.com/kloudlite/kloudlite/api/internal/repository"
	"go.uber.org/zap"
)

// Manager coordinates all services and provides a unified interface
type Manager struct {
	// Repository manager (exposed for direct repository access)
	RepositoryManager *repository.Manager

	// Individual services
	Users UserService
	Auth  AuthService

	// You can add other services here as needed:
	// Workspaces WorkspaceService
	// Environments EnvironmentService
	// etc.
}

// ManagerOptions contains options for creating a services manager
type ManagerOptions struct {
	// Repository manager
	RepositoryManager *repository.Manager

	// Configuration
	Config *config.Config

	// Logger
	Logger *zap.Logger
}

// NewManager creates a new services manager
func NewManager(ctx context.Context, opts *ManagerOptions) (*Manager, error) {
	if opts == nil {
		return nil, fmt.Errorf("ManagerOptions is required")
	}

	if opts.RepositoryManager == nil {
		return nil, fmt.Errorf("RepositoryManager is required")
	}

	if opts.Config == nil {
		return nil, fmt.Errorf("Config is required")
	}

	if opts.Logger == nil {
		return nil, fmt.Errorf("Logger is required")
	}

	// Create individual services
	userService := NewUserService(opts.RepositoryManager.Users, opts.RepositoryManager.WorkMachines)

	// Create auth service
	tokenExpiry := time.Duration(opts.Config.Auth.TokenExpiryHours) * time.Hour
	authService := NewAuthService(
		opts.Config.Auth.JWTSecret,
		tokenExpiry,
		userService,
		opts.Logger,
	)

	return &Manager{
		RepositoryManager: opts.RepositoryManager,
		Users:             userService,
		Auth:              authService,
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
