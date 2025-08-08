// Package template provides a template for creating new backend services
// Copy this template when creating a new service in /api/apps/{service-name}

package template

import (
	"context"
	"fmt"

	"github.com/kloudlite/api/pkg/repos"
	"go.uber.org/fx"
	"log/slog"
)

// ===== ENTITIES =====

// Entity represents your domain model
type Entity struct {
	repos.BaseEntity `json:",inline" bson:",inline"`
	
	// Core fields
	Name        string            `json:"name" bson:"name" validate:"required,min=1,max=100"`
	Description string            `json:"description,omitempty" bson:"description,omitempty"`
	Status      string            `json:"status" bson:"status" validate:"required,oneof=active inactive pending"`
	
	// Metadata
	Metadata    map[string]string `json:"metadata,omitempty" bson:"metadata,omitempty"`
}

// GetIndexes returns MongoDB indexes for this entity
func (e *Entity) GetIndexes() []repos.IndexSpec {
	return []repos.IndexSpec{
		{
			Name: "name_unique",
			Key:  repos.IndexKey{Key: "name", Value: repos.IndexAsc},
			Unique: true,
		},
		{
			Name: "status_created",
			Key: []repos.IndexKey{
				{Key: "status", Value: repos.IndexAsc},
				{Key: "createdAt", Value: repos.IndexDesc},
			},
		},
	}
}

// ===== DOMAIN INTERFACE =====

// Domain defines the business logic interface
type Domain interface {
	// Entity operations
	CreateEntity(ctx context.Context, entity *Entity) (*Entity, error)
	GetEntity(ctx context.Context, id repos.ID) (*Entity, error)
	GetEntityByName(ctx context.Context, name string) (*Entity, error)
	UpdateEntity(ctx context.Context, id repos.ID, updates repos.Document) (*Entity, error)
	DeleteEntity(ctx context.Context, id repos.ID) error
	ListEntities(ctx context.Context, filter repos.Filter, pagination repos.Pagination) ([]*Entity, int64, error)
}

// ===== DOMAIN IMPLEMENTATION =====

type domain struct {
	entityRepo repos.DbRepo[*Entity]
	logger     *slog.Logger
}

// NewDomain creates a new domain instance
func NewDomain(
	entityRepo repos.DbRepo[*Entity],
	logger *slog.Logger,
) Domain {
	return &domain{
		entityRepo: entityRepo,
		logger:     logger,
	}
}

func (d *domain) CreateEntity(ctx context.Context, entity *Entity) (*Entity, error) {
	// Validate business rules
	if err := d.validateEntity(entity); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}
	
	// Check for duplicates
	existing, err := d.entityRepo.FindOne(ctx, repos.Filter{"name": entity.Name})
	if err == nil && existing != nil {
		return nil, fmt.Errorf("entity with name %s already exists", entity.Name)
	}
	
	// Create entity
	created, err := d.entityRepo.Create(ctx, entity)
	if err != nil {
		return nil, fmt.Errorf("failed to create entity: %w", err)
	}
	
	d.logger.Info("entity created", "id", created.ID, "name", created.Name)
	return created, nil
}

func (d *domain) GetEntity(ctx context.Context, id repos.ID) (*Entity, error) {
	entity, err := d.entityRepo.FindById(ctx, id)
	if err != nil {
		if repos.IsErrNotFound(err) {
			return nil, fmt.Errorf("entity not found")
		}
		return nil, fmt.Errorf("failed to get entity: %w", err)
	}
	return entity, nil
}

func (d *domain) GetEntityByName(ctx context.Context, name string) (*Entity, error) {
	entity, err := d.entityRepo.FindOne(ctx, repos.Filter{"name": name})
	if err != nil {
		if repos.IsErrNotFound(err) {
			return nil, fmt.Errorf("entity not found")
		}
		return nil, fmt.Errorf("failed to get entity: %w", err)
	}
	return entity, nil
}

func (d *domain) UpdateEntity(ctx context.Context, id repos.ID, updates repos.Document) (*Entity, error) {
	// Get existing entity
	entity, err := d.GetEntity(ctx, id)
	if err != nil {
		return nil, err
	}
	
	// Apply updates and validate
	if name, ok := updates["name"].(string); ok {
		entity.Name = name
	}
	if description, ok := updates["description"].(string); ok {
		entity.Description = description
	}
	
	if err := d.validateEntity(entity); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}
	
	// Update in database
	updated, err := d.entityRepo.UpdateById(ctx, id, updates)
	if err != nil {
		return nil, fmt.Errorf("failed to update entity: %w", err)
	}
	
	d.logger.Info("entity updated", "id", updated.ID)
	return updated, nil
}

func (d *domain) DeleteEntity(ctx context.Context, id repos.ID) error {
	// Check if entity exists
	entity, err := d.GetEntity(ctx, id)
	if err != nil {
		return err
	}
	
	// Perform any cleanup or checks before deletion
	// For example, check if entity is not in use
	
	// Delete entity
	if err := d.entityRepo.DeleteById(ctx, id); err != nil {
		return fmt.Errorf("failed to delete entity: %w", err)
	}
	
	d.logger.Info("entity deleted", "id", id, "name", entity.Name)
	return nil
}

func (d *domain) ListEntities(ctx context.Context, filter repos.Filter, pagination repos.Pagination) ([]*Entity, int64, error) {
	// Add default sorting if not specified
	if pagination.SortBy == "" {
		pagination.SortBy = "createdAt"
		pagination.SortOrder = repos.SortOrderDesc
	}
	
	// Get entities
	entities, err := d.entityRepo.Find(ctx, filter, repos.WithPagination(pagination))
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list entities: %w", err)
	}
	
	// Get total count
	count, err := d.entityRepo.Count(ctx, filter)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count entities: %w", err)
	}
	
	return entities, count, nil
}

func (d *domain) validateEntity(entity *Entity) error {
	// Add business validation logic here
	if entity.Name == "" {
		return fmt.Errorf("name is required")
	}
	
	if len(entity.Name) > 100 {
		return fmt.Errorf("name must be less than 100 characters")
	}
	
	// Add more validation as needed
	return nil
}

// ===== DEPENDENCY INJECTION MODULE =====

var Module = fx.Module(
	"template-service",
	fx.Provide(
		NewDomain,
		// Add other providers here
	),
)