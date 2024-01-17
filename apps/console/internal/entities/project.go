package entities

import (
	fc "github.com/kloudlite/api/apps/console/internal/entities/field-constants"
	"github.com/kloudlite/api/common"
	"github.com/kloudlite/api/pkg/repos"
	t "github.com/kloudlite/api/pkg/types"
	crdsv1 "github.com/kloudlite/operator/apis/crds/v1"
)

type Project struct {
	repos.BaseEntity `json:",inline" graphql:"noinput"`

	crdsv1.Project `json:",inline"`

	common.ResourceMetadata `json:",inline"`

	AccountName string  `json:"accountName" graphql:"noinput"`
	ClusterName *string `json:"clusterName"`

	SyncStatus t.SyncStatus `json:"syncStatus" graphql:"noinput"`
}

func (p *Project) GetResourceType() ResourceType {
	return ResourceTypeProject
}

var ProjectIndexes = []repos.IndexField{
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
		},
		Unique: true,
	},
	{
		Field: []repos.IndexKey{
			{Key: fc.AccountName, Value: repos.IndexAsc},
			{Key: fc.ProjectSpecTargetNamespace, Value: repos.IndexAsc},
		},
		Unique: true,
	},

	{
		Field: []repos.IndexKey{
			{Key: fc.ProjectClusterName, Value: repos.IndexAsc},
		},
	},
}
