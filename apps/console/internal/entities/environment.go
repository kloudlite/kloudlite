package entities

import (
	fc "github.com/kloudlite/api/apps/console/internal/entities/field-constants"
	"github.com/kloudlite/api/common"
	"github.com/kloudlite/api/common/fields"
	"github.com/kloudlite/api/pkg/repos"
	t "github.com/kloudlite/api/pkg/types"
	crdsv1 "github.com/kloudlite/operator/apis/crds/v1"
	"github.com/kloudlite/operator/pkg/operator"
)

type Environment struct {
	repos.BaseEntity `json:",inline" graphql:"noinput"`

	crdsv1.Environment `json:",inline"`

	AccountName string `json:"accountName" graphql:"noinput"`
	ClusterName string `json:"clusterName"`

	IsArchived *bool `json:"isArchived,omitempty" graphql:"noinput"`

	common.ResourceMetadata `json:",inline"`

	SyncStatus t.SyncStatus `json:"syncStatus" graphql:"noinput"`
}

func (e *Environment) GetDisplayName() string {
	return e.ResourceMetadata.DisplayName
}

func (e *Environment) GetStatus() operator.Status {
	return e.Environment.Status
}

func (e *Environment) GetResourceType() ResourceType {
	return ResourceTypeEnvironment
}

var EnvironmentIndexes = []repos.IndexField{
	{
		Field: []repos.IndexKey{
			{Key: fields.Id, Value: repos.IndexAsc},
		},
		Unique: true,
	},
	{
		Field: []repos.IndexKey{
			{Key: fields.MetadataName, Value: repos.IndexAsc},
			{Key: fields.MetadataNamespace, Value: repos.IndexAsc},
			{Key: fields.AccountName, Value: repos.IndexAsc},
			{Key: fields.ClusterName, Value: repos.IndexAsc},
		},
		Unique: true,
	},
	{
		Field: []repos.IndexKey{
			{Key: fields.AccountName, Value: repos.IndexAsc},
			{Key: fc.EnvironmentSpecTargetNamespace, Value: repos.IndexAsc},
		},
		Unique: true,
	},
}

func EnvironmentDBFilter(accountName string, envName string) repos.Filter {
	return repos.Filter{
		fc.AccountName:  accountName,
		fc.MetadataName: envName,
	}
}
