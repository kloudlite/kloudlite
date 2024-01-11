package entities

import "github.com/kloudlite/api/pkg/repos"

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
)

type ResourceMapping struct {
	repos.BaseEntity `bson:",inline"`

	ResourceType      ResourceType `json:"resourceType"`
	ResourceName      string       `json:"resourceName"`
	ResourceNamespace string       `json:"resourceNamespace"`

	AccountName     string `json:"accountName"`
	ProjectName     string `json:"projectName"`
	EnvironmentName string `json:"environmentName"`
}

var ResourceMappingIndices = []repos.IndexField{
	{
		Field: []repos.IndexKey{
			{Key: "id", Value: repos.IndexAsc},
		},
		Unique: true,
	},
	{
		Field: []repos.IndexKey{
			{Key: "accountName", Value: repos.IndexAsc},
			{Key: "projectName", Value: repos.IndexAsc},
			{Key: "environmentName", Value: repos.IndexAsc},
			{Key: "resourceType", Value: repos.IndexAsc},
			{Key: "resourceName", Value: repos.IndexAsc},
			{Key: "resourceNamespace", Value: repos.IndexAsc},
		},
		Unique: true,
	},
}
