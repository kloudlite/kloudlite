package entities

import (
	fc "github.com/kloudlite/api/apps/console/internal/entities/field-constants"
	"github.com/kloudlite/api/common"
	"github.com/kloudlite/api/pkg/repos"
	t "github.com/kloudlite/api/pkg/types"
	ct "github.com/kloudlite/operator/apis/common-types"
)

type ImportedManagedResource struct {
	repos.BaseEntity `json:",inline" graphql:"noinput"`

	Name string `json:"name"`

	ManagedResourceRef `json:"managedResourceRef"`

	common.ResourceMetadata `json:",inline"`

	SecretRef ct.SecretRef `json:"secretRef"`

	AccountName     string `json:"accountName" graphql:"noinput"`
	EnvironmentName string `json:"environmentName"`

	SyncStatus t.SyncStatus `json:"syncStatus"`
}

func (m *ImportedManagedResource) GetResourceType() ResourceType {
	return ResourceTypeImportedManagedResource
}

type ManagedResourceRef struct {
	ID        repos.ID `json:"id"`
	Name      string   `json:"name"`
	Namespace string   `json:"namespace"`
}

var ImportedManagedResourceIndexes = []repos.IndexField{
	{
		Field: []repos.IndexKey{
			{Key: fc.Id, Value: repos.IndexAsc},
		},
		Unique: true,
	},
	{
		Field: []repos.IndexKey{
			{Key: fc.AccountName, Value: repos.IndexAsc},
			{Key: fc.EnvironmentName, Value: repos.IndexAsc},
			{Key: fc.ImportedManagedResourceName, Value: repos.IndexAsc},
			{Key: fc.ImportedManagedResourceSecretRefName, Value: repos.IndexAsc},
			{Key: fc.ImportedManagedResourceSecretRefNamespace, Value: repos.IndexAsc},
		},
		Unique: true,
	},
}
