package entities

import (
	fc "github.com/kloudlite/api/apps/console/internal/entities/field-constants"
	"github.com/kloudlite/api/common"
	"github.com/kloudlite/api/pkg/repos"
	t "github.com/kloudlite/api/pkg/types"
	crdsv1 "github.com/kloudlite/operator/apis/crds/v1"
)

type Router struct {
	repos.BaseEntity `json:",inline" graphql:"noinput"`
	crdsv1.Router    `json:",inline"`

	AccountName     string `json:"accountName" graphql:"noinput"`
	ProjectName     string `json:"projectName" graphql:"noinput"`
	EnvironmentName string `json:"environmentName" graphql:"noinput"`

	common.ResourceMetadata `json:",inline"`
	SyncStatus              t.SyncStatus `json:"syncStatus" graphql:"noinput"`
}

func (r *Router) GetResourceType() ResourceType {
	return ResourceTypeRouter
}

var RouterIndexes = []repos.IndexField{
	{
		Field: []repos.IndexKey{
			{Key: fc.Id, Value: repos.IndexAsc},
		},
		Unique: true,
	},
	{
		Field: []repos.IndexKey{
			{Key: fc.MetadataName, Value: repos.IndexAsc},
			{Key: fc.MetadataNamespace, Value: repos.IndexAsc},
			{Key: fc.AccountName, Value: repos.IndexAsc},
			{Key: fc.ProjectName, Value: repos.IndexAsc},
			{Key: fc.EnvironmentName, Value: repos.IndexAsc},
		},
		Unique: true,
	},
}
