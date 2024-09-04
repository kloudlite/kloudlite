package entities

import (
	fc "github.com/kloudlite/api/apps/console/internal/entities/field-constants"
	"github.com/kloudlite/api/common/fields"
	"github.com/kloudlite/api/pkg/repos"
)

type ResourceType string

const (
	ResourceTypeEnvironment             ResourceType = "environment"
	ResourceTypeApp                     ResourceType = "app"
	ResourceTypeExternalApp             ResourceType = "external_app"
	ResourceTypeConfig                  ResourceType = "config"
	ResourceTypeSecret                  ResourceType = "secret"
	ResourceTypeImagePullSecret         ResourceType = "image_pull_secret"
	ResourceTypeRouter                  ResourceType = "router"
	ResourceTypeManagedResource         ResourceType = "managed_resource"
	ResourceTypeImportedManagedResource ResourceType = "imported_managed_resource"
	ResourceTypeClusterManagedService   ResourceType = "cluster_managed_service"
	ResourceTypeServiceBinding          ResourceType = "service_binding"
)

type ResourceHeirarchy string

const (
	ResourceHeirarchyEnvironment ResourceHeirarchy = "environment"
)

// ResourceMapping represents a relationship
// between a resource (i.e. Environment, App, Router etc.) with it's {account, cluster and environment}
type ResourceMapping struct {
	repos.BaseEntity `bson:",inline"`

	ResourceHeirarchy ResourceHeirarchy `json:"resourceHeirarchy"`

	ResourceType      ResourceType `json:"resourceType"`
	ResourceName      string       `json:"resourceName"`
	ResourceNamespace string       `json:"resourceNamespace"`

	AccountName string `json:"accountName"`
	ClusterName string `json:"clusterName"`

	EnvironmentName string `json:"environmentName"`
}

var ResourceMappingIndices = []repos.IndexField{
	{
		Field: []repos.IndexKey{
			{Key: fields.Id, Value: repos.IndexAsc},
		},
		Unique: true,
	},
	{
		Field: []repos.IndexKey{
			{Key: fields.AccountName, Value: repos.IndexAsc},
			{Key: fields.EnvironmentName, Value: repos.IndexAsc},
			{Key: fc.ResourceMappingResourceType, Value: repos.IndexAsc},
			{Key: fc.ResourceMappingResourceName, Value: repos.IndexAsc},
		},
		Unique: true,
	},
	{
		Field: []repos.IndexKey{
			{Key: fields.ClusterName, Value: repos.IndexAsc},
			{Key: fc.ResourceMappingResourceType, Value: repos.IndexAsc},
			{Key: fc.ResourceMappingResourceName, Value: repos.IndexAsc},
			{Key: fc.ResourceMappingResourceNamespace, Value: repos.IndexAsc},
		},
		Unique: true,
	},
}
