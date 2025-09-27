// Package domain provides business logic for {{SERVICE_NAME}} service
package domain

import (
	"context"
	"errors"
	"fmt"
	
	"github.com/kloudlite/api/apps/{{SERVICE_NAME}}/internal/entities"
	"github.com/kloudlite/api/pkg/repos"
	"go.uber.org/fx"
	"log/slog"
)

// Common errors
var (
	Err{{ENTITY}}NotFound      = errors.New("{{ENTITY_LOWER}} not found")
	Err{{ENTITY}}AlreadyExists = errors.New("{{ENTITY_LOWER}} already exists")
	ErrInvalid{{ENTITY}}       = errors.New("invalid {{ENTITY_LOWER}}")
	ErrPermissionDenied        = errors.New("permission denied")
)

// {{ENTITY}}Service defines the business logic interface
type {{ENTITY}}Service interface {
	Create{{ENTITY}}(ctx UserContext, {{ENTITY_LOWER}} *entities.{{ENTITY}}) (*entities.{{ENTITY}}, error)
	Get{{ENTITY}}(ctx UserContext, id repos.ID) (*entities.{{ENTITY}}, error)
	Update{{ENTITY}}(ctx UserContext, id repos.ID, updates map[string]interface{}) (*entities.{{ENTITY}}, error)
	Delete{{ENTITY}}(ctx UserContext, id repos.ID) error
	List{{ENTITY}}s(ctx UserContext, filter Filter, opts ...ListOption) ([]*entities.{{ENTITY}}, error)
}

// {{ENTITY_LOWER}}Service implements {{ENTITY}}Service
type {{ENTITY_LOWER}}Service struct {
	repo   repos.DbRepo[*entities.{{ENTITY}}]
	logger *slog.Logger
}

// New{{ENTITY}}Service creates a new {{ENTITY}}Service instance
func New{{ENTITY}}Service(
	repo repos.DbRepo[*entities.{{ENTITY}}],
	logger *slog.Logger,
) {{ENTITY}}Service {
	return &{{ENTITY_LOWER}}Service{
		repo:   repo,
		logger: logger.With("service", "{{ENTITY_LOWER}}"),
	}
}

// Create{{ENTITY}} creates a new {{ENTITY_LOWER}}
func (s *{{ENTITY_LOWER}}Service) Create{{ENTITY}}(ctx UserContext, {{ENTITY_LOWER}} *entities.{{ENTITY}}) (*entities.{{ENTITY}}, error) {
	// Validate input
	if err := s.validate{{ENTITY}}({{ENTITY_LOWER}}); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}
	
	// Check permissions
	if !s.canCreate{{ENTITY}}(ctx) {
		return nil, ErrPermissionDenied
	}
	
	// Check if already exists
	existing, err := s.repo.FindOne(ctx, repos.Filter{
		"name": {{ENTITY_LOWER}}.Name,
	})
	if err == nil && existing != nil {
		return nil, Err{{ENTITY}}AlreadyExists
	}
	
	// Set metadata
	{{ENTITY_LOWER}}.CreatedBy = ctx.GetUserId()
	
	// Create {{ENTITY_LOWER}}
	created, err := s.repo.Create(ctx, {{ENTITY_LOWER}})
	if err != nil {
		s.logger.Error("failed to create {{ENTITY_LOWER}}", "error", err)
		return nil, fmt.Errorf("failed to create {{ENTITY_LOWER}}: %w", err)
	}
	
	s.logger.Info("{{ENTITY_LOWER}} created", "id", created.ID, "userId", ctx.GetUserId())
	return created, nil
}

// Get{{ENTITY}} retrieves a {{ENTITY_LOWER}} by ID
func (s *{{ENTITY_LOWER}}Service) Get{{ENTITY}}(ctx UserContext, id repos.ID) (*entities.{{ENTITY}}, error) {
	{{ENTITY_LOWER}}, err := s.repo.FindById(ctx, id)
	if err != nil {
		if errors.Is(err, repos.ErrNotFound) {
			return nil, Err{{ENTITY}}NotFound
		}
		return nil, fmt.Errorf("failed to get {{ENTITY_LOWER}}: %w", err)
	}
	
	// Check permissions
	if !s.canView{{ENTITY}}(ctx, {{ENTITY_LOWER}}) {
		return nil, ErrPermissionDenied
	}
	
	return {{ENTITY_LOWER}}, nil
}

// Update{{ENTITY}} updates an existing {{ENTITY_LOWER}}
func (s *{{ENTITY_LOWER}}Service) Update{{ENTITY}}(ctx UserContext, id repos.ID, updates map[string]interface{}) (*entities.{{ENTITY}}, error) {
	// Get existing {{ENTITY_LOWER}}
	existing, err := s.Get{{ENTITY}}(ctx, id)
	if err != nil {
		return nil, err
	}
	
	// Check permissions
	if !s.canUpdate{{ENTITY}}(ctx, existing) {
		return nil, ErrPermissionDenied
	}
	
	// Add metadata
	updates["updatedBy"] = ctx.GetUserId()
	
	// Update {{ENTITY_LOWER}}
	updated, err := s.repo.UpdateById(ctx, id, updates)
	if err != nil {
		s.logger.Error("failed to update {{ENTITY_LOWER}}", "error", err, "id", id)
		return nil, fmt.Errorf("failed to update {{ENTITY_LOWER}}: %w", err)
	}
	
	s.logger.Info("{{ENTITY_LOWER}} updated", "id", id, "userId", ctx.GetUserId())
	return updated, nil
}

// Delete{{ENTITY}} deletes a {{ENTITY_LOWER}}
func (s *{{ENTITY_LOWER}}Service) Delete{{ENTITY}}(ctx UserContext, id repos.ID) error {
	// Get existing {{ENTITY_LOWER}}
	existing, err := s.Get{{ENTITY}}(ctx, id)
	if err != nil {
		return err
	}
	
	// Check permissions
	if !s.canDelete{{ENTITY}}(ctx, existing) {
		return ErrPermissionDenied
	}
	
	// Delete {{ENTITY_LOWER}}
	if err := s.repo.DeleteById(ctx, id); err != nil {
		s.logger.Error("failed to delete {{ENTITY_LOWER}}", "error", err, "id", id)
		return fmt.Errorf("failed to delete {{ENTITY_LOWER}}: %w", err)
	}
	
	s.logger.Info("{{ENTITY_LOWER}} deleted", "id", id, "userId", ctx.GetUserId())
	return nil
}

// List{{ENTITY}}s lists {{ENTITY_LOWER}}s with filtering and pagination
func (s *{{ENTITY_LOWER}}Service) List{{ENTITY}}s(ctx UserContext, filter Filter, opts ...ListOption) ([]*entities.{{ENTITY}}, error) {
	// Apply permission filter
	permissionFilter := s.getPermissionFilter(ctx)
	finalFilter := mergeFilters(filter, permissionFilter)
	
	// List {{ENTITY_LOWER}}s
	{{ENTITY_LOWER}}s, err := s.repo.Find(ctx, finalFilter, opts...)
	if err != nil {
		s.logger.Error("failed to list {{ENTITY_LOWER}}s", "error", err)
		return nil, fmt.Errorf("failed to list {{ENTITY_LOWER}}s: %w", err)
	}
	
	return {{ENTITY_LOWER}}s, nil
}

// Helper methods

func (s *{{ENTITY_LOWER}}Service) validate{{ENTITY}}({{ENTITY_LOWER}} *entities.{{ENTITY}}) error {
	if {{ENTITY_LOWER}}.Name == "" {
		return errors.New("name is required")
	}
	
	// Add more validation as needed
	
	return nil
}

func (s *{{ENTITY_LOWER}}Service) canCreate{{ENTITY}}(ctx UserContext) bool {
	// Implement permission logic
	return true
}

func (s *{{ENTITY_LOWER}}Service) canView{{ENTITY}}(ctx UserContext, {{ENTITY_LOWER}} *entities.{{ENTITY}}) bool {
	// Implement permission logic
	return true
}

func (s *{{ENTITY_LOWER}}Service) canUpdate{{ENTITY}}(ctx UserContext, {{ENTITY_LOWER}} *entities.{{ENTITY}}) bool {
	// Implement permission logic
	return {{ENTITY_LOWER}}.CreatedBy == ctx.GetUserId()
}

func (s *{{ENTITY_LOWER}}Service) canDelete{{ENTITY}}(ctx UserContext, {{ENTITY_LOWER}} *entities.{{ENTITY}}) bool {
	// Implement permission logic
	return {{ENTITY_LOWER}}.CreatedBy == ctx.GetUserId()
}

func (s *{{ENTITY_LOWER}}Service) getPermissionFilter(ctx UserContext) Filter {
	// Return filter based on user permissions
	return Filter{
		"createdBy": ctx.GetUserId(),
	}
}

// Module exports for dependency injection
var Module = fx.Module(
	"{{ENTITY_LOWER}}-service",
	fx.Provide(New{{ENTITY}}Service),
)