package entities

import (
	"time"
	"github.com/kloudlite/api/pkg/repos"
)

// {{ENTITY}} represents a {{ENTITY_LOWER}} in the system
type {{ENTITY}} struct {
	repos.BaseEntity `json:",inline" bson:",inline"`
	
	// Core fields
	Name        string `json:"name" bson:"name" validate:"required,min=1,max=100"`
	Slug        string `json:"slug" bson:"slug" validate:"required,slug,min=3,max=63"`
	DisplayName string `json:"displayName" bson:"displayName" validate:"required,min=1,max=200"`
	Description string `json:"description,omitempty" bson:"description,omitempty" validate:"max=500"`
	
	// Status and type
	Status {{ENTITY}}Status `json:"status" bson:"status"`
	Type   {{ENTITY}}Type   `json:"type" bson:"type"`
	
	// Relationships
	OwnerId  repos.ID `json:"ownerId" bson:"ownerId"`
	TeamId   repos.ID `json:"teamId,omitempty" bson:"teamId,omitempty"`
	ParentId repos.ID `json:"parentId,omitempty" bson:"parentId,omitempty"`
	
	// Metadata
	Tags     []string          `json:"tags,omitempty" bson:"tags,omitempty"`
	Metadata map[string]string `json:"metadata,omitempty" bson:"metadata,omitempty"`
	
	// Audit fields
	CreatedBy string     `json:"createdBy" bson:"createdBy"`
	UpdatedBy string     `json:"updatedBy,omitempty" bson:"updatedBy,omitempty"`
	DeletedAt *time.Time `json:"deletedAt,omitempty" bson:"deletedAt,omitempty"`
}

// {{ENTITY}}Status represents the status of a {{ENTITY_LOWER}}
type {{ENTITY}}Status string

const (
	{{ENTITY}}StatusActive   {{ENTITY}}Status = "active"
	{{ENTITY}}StatusInactive {{ENTITY}}Status = "inactive"
	{{ENTITY}}StatusPending  {{ENTITY}}Status = "pending"
	{{ENTITY}}StatusDeleted  {{ENTITY}}Status = "deleted"
)

// IsValid checks if the status is valid
func (s {{ENTITY}}Status) IsValid() bool {
	switch s {
	case {{ENTITY}}StatusActive, {{ENTITY}}StatusInactive, {{ENTITY}}StatusPending, {{ENTITY}}StatusDeleted:
		return true
	default:
		return false
	}
}

// {{ENTITY}}Type represents the type of {{ENTITY_LOWER}}
type {{ENTITY}}Type string

const (
	{{ENTITY}}TypeStandard {{ENTITY}}Type = "standard"
	{{ENTITY}}TypePremium  {{ENTITY}}Type = "premium"
	{{ENTITY}}TypeCustom   {{ENTITY}}Type = "custom"
)

// IsValid checks if the type is valid
func (t {{ENTITY}}Type) IsValid() bool {
	switch t {
	case {{ENTITY}}TypeStandard, {{ENTITY}}TypePremium, {{ENTITY}}TypeCustom:
		return true
	default:
		return false
	}
}

// GetCollectionName returns the MongoDB collection name
func ({{ENTITY}} {{ENTITY}}) GetCollectionName() string {
	return "{{ENTITY_LOWER_PLURAL}}"
}

// GetIndexes returns the indexes for the collection
func ({{ENTITY}} {{ENTITY}}) GetIndexes() []repos.IndexSpec {
	return []repos.IndexSpec{
		{
			Name: "slug_unique",
			Key: repos.IndexKey{
				Key:   "slug",
				Value: repos.IndexAsc,
			},
			Unique: true,
		},
		{
			Name: "owner_status",
			Key: []repos.IndexKey{
				{Key: "ownerId", Value: repos.IndexAsc},
				{Key: "status", Value: repos.IndexAsc},
			},
		},
		{
			Name: "team_status",
			Key: []repos.IndexKey{
				{Key: "teamId", Value: repos.IndexAsc},
				{Key: "status", Value: repos.IndexAsc},
			},
		},
		{
			Name: "created_at_desc",
			Key: repos.IndexKey{
				Key:   "createdAt",
				Value: repos.IndexDesc,
			},
		},
		{
			Name: "search_text",
			Key: []repos.IndexKey{
				{Key: "name", Value: repos.IndexText},
				{Key: "displayName", Value: repos.IndexText},
				{Key: "description", Value: repos.IndexText},
			},
		},
	}
}

// Validate performs custom validation
func (e *{{ENTITY}}) Validate() error {
	if !e.Status.IsValid() {
		return fmt.Errorf("invalid status: %s", e.Status)
	}
	
	if !e.Type.IsValid() {
		return fmt.Errorf("invalid type: %s", e.Type)
	}
	
	// Add more custom validation as needed
	
	return nil
}

// BeforeCreate hook - called before creating
func (e *{{ENTITY}}) BeforeCreate() error {
	// Set defaults
	if e.Status == "" {
		e.Status = {{ENTITY}}StatusPending
	}
	
	if e.Type == "" {
		e.Type = {{ENTITY}}TypeStandard
	}
	
	// Generate slug from name if not provided
	if e.Slug == "" && e.Name != "" {
		e.Slug = generateSlug(e.Name)
	}
	
	return e.Validate()
}

// BeforeUpdate hook - called before updating
func (e *{{ENTITY}}) BeforeUpdate() error {
	return e.Validate()
}

// Related types

// {{ENTITY}}Filter represents filter options
type {{ENTITY}}Filter struct {
	Status   {{ENTITY}}Status   `json:"status,omitempty"`
	Type     {{ENTITY}}Type     `json:"type,omitempty"`
	OwnerId  repos.ID           `json:"ownerId,omitempty"`
	TeamId   repos.ID           `json:"teamId,omitempty"`
	Tags     []string           `json:"tags,omitempty"`
	Search   string             `json:"search,omitempty"`
}

// {{ENTITY}}ListOptions represents list options
type {{ENTITY}}ListOptions struct {
	Page     int      `json:"page"`
	PageSize int      `json:"pageSize"`
	SortBy   string   `json:"sortBy"`
	SortDesc bool     `json:"sortDesc"`
}