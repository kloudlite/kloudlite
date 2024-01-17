package entities

import (
	fc "github.com/kloudlite/api/apps/console/internal/entities/field-constants"
	"github.com/kloudlite/api/pkg/repos"
)

type ResourceType string

const (
	ResourceTypeProject               ResourceType = "project"
	ResourceTypeEnvironment           ResourceType = "environment"
	ResourceTypeApp                   ResourceType = "app"
	ResourceTypeConfig                ResourceType = "config"
	ResourceTypeSecret                ResourceType = "secret"
	ResourceTypeImagePullSecret       ResourceType = "image_pull_secret"
	ResourceTypeRouter                ResourceType = "router"
	ResourceTypeManagedResource       ResourceType = "managed_resource"
	ResourceTypeProjectManagedService ResourceType = "project_managed_service"
	ResourceTypeVPNDevice             ResourceType = "vpn_device"
)

type ResourceHeirarchy string

const (
	ResourceHeirarchyProject     ResourceHeirarchy = "project"
	ResourceHeirarchyEnvironment ResourceHeirarchy = "environment"
)

type ResourceMapping struct {
	repos.BaseEntity `bson:",inline"`

	ResourceHeirarchy ResourceHeirarchy `json:"resourceHeirarchy"`

	ResourceType      ResourceType `json:"resourceType"`
	ResourceName      string       `json:"resourceName"`
	ResourceNamespace string       `json:"resourceNamespace"`

	AccountName string `json:"accountName"`
	ClusterName string `json:"clusterName"`

	ProjectName     string `json:"projectName"`
	EnvironmentName string `json:"environmentName"`
}

var ResourceMappingIndices = []repos.IndexField{
	{
		Field: []repos.IndexKey{
			{Key: fc.Id, Value: repos.IndexAsc},
		},
		Unique: true,
	},
	{
		Field: []repos.IndexKey{
			{Key: fc.AccountName, Value: repos.IndexAsc},
			{Key: fc.ProjectName, Value: repos.IndexAsc},
			{Key: fc.ResourceMappingResourceType, Value: repos.IndexAsc},
			{Key: fc.EnvironmentName, Value: repos.IndexAsc},
			{Key: fc.ResourceMappingResourceName, Value: repos.IndexAsc},
		},
		Unique: true,
	},
	{
		Field: []repos.IndexKey{
			{Key: fc.ResourceMappingClusterName, Value: repos.IndexAsc},
			{Key: fc.ResourceMappingResourceType, Value: repos.IndexAsc},
			{Key: fc.ResourceMappingResourceName, Value: repos.IndexAsc},
			{Key: fc.ResourceMappingResourceNamespace, Value: repos.IndexAsc},
		},
		Unique: true,
	},
}
