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

type Project struct {
	repos.BaseEntity `json:",inline" graphql:"noinput"`

	crdsv1.Project `json:",inline"`

	common.ResourceMetadata `json:",inline"`

	AccountName string  `json:"accountName" graphql:"noinput"`
	ClusterName *string `json:"clusterName"`

	SyncStatus t.SyncStatus `json:"syncStatus" graphql:"noinput"`
}

func (p *Project) GetDisplayName() string {
	return p.ResourceMetadata.DisplayName
}

func (p *Project) GetStatus() operator.Status {
	return p.Project.Status
}

func (p *Project) GetResourceType() ResourceType {
	return ResourceTypeProject
}

var ProjectIndexes = []repos.IndexField{
	{
		Field: []repos.IndexKey{
			{Key: fields.Id, Value: repos.IndexAsc},
		},
		Unique: true,
	},
	{
		Field: []repos.IndexKey{
			{Key: fields.MetadataName, Value: repos.IndexAsc},
			{Key: fields.AccountName, Value: repos.IndexAsc},
		},
		Unique: true,
	},
	{
		Field: []repos.IndexKey{
			{Key: fields.AccountName, Value: repos.IndexAsc},
			{Key: fc.ProjectSpecTargetNamespace, Value: repos.IndexAsc},
		},
		Unique: true,
	},

	{
		Field: []repos.IndexKey{
			{Key: fields.ClusterName, Value: repos.IndexAsc},
		},
	},
}
