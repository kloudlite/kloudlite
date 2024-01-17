package entities

import (
	fc "github.com/kloudlite/api/apps/console/internal/entities/field-constants"
	"github.com/kloudlite/api/common"
	"github.com/kloudlite/api/pkg/repos"
	t "github.com/kloudlite/api/pkg/types"
	crdsv1 "github.com/kloudlite/operator/apis/crds/v1"
	corev1 "k8s.io/api/core/v1"
)

type ProjectManagedService struct {
	repos.BaseEntity             `json:",inline" graphql:"noinput"`
	crdsv1.ProjectManagedService `json:",inline"`

	AccountName string `json:"accountName" graphql:"noinput"`
	ProjectName string `json:"projectName" graphql:"noinput"`

	SyncedOutputSecretRef *corev1.Secret `json:"syncedOutputSecretRef" graphql:"ignore" struct-json-path:",ignore-nesting"`

	common.ResourceMetadata `json:",inline"`
	SyncStatus              t.SyncStatus `json:"syncStatus" graphql:"noinput"`
}

// GetResourceType implements domain.resource.
func (*ProjectManagedService) GetResourceType() ResourceType {
	return ResourceTypeProjectManagedService
}

var ProjectManagedServiceIndices = []repos.IndexField{
	{
		Field: []repos.IndexKey{
			{Key: fc.Id, Value: repos.IndexAsc},
		},
		Unique: true,
	},
	{
		Field: []repos.IndexKey{
			{Key: fc.MetadataName, Value: repos.IndexAsc},
			{Key: fc.AccountName, Value: repos.IndexAsc},
			{Key: fc.ProjectName, Value: repos.IndexAsc},
		},
		Unique: true,
	},
}
