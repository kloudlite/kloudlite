package entities

import (
	"github.com/kloudlite/api/common"
	"github.com/kloudlite/api/common/fields"
	"github.com/kloudlite/api/pkg/repos"
	"github.com/kloudlite/api/pkg/types"
	crdsv1 "github.com/kloudlite/operator/apis/crds/v1"
	"github.com/kloudlite/operator/toolkit/reconciler"
)

type Workspace struct {
	repos.BaseEntity        `json:",inline" graphql:"noinput"`
	common.ResourceMetadata `json:",inline"`

	crdsv1.Workspace `json:",inline"`

	AccountName string `json:"accountName" graphql:"noinput"`
	ClusterName string `json:"clusterName" graphql:"noinput"`

	DispatchAddr *DispatchAddr `json:"dispatchAddr" graphql:"noinput"`

	SyncStatus types.SyncStatus `json:"syncStatus" graphql:"noinput"`
}

func (v *Workspace) GetDisplayName() string {
	return v.ResourceMetadata.DisplayName
}

func (v *Workspace) GetStatus() reconciler.Status {
	return v.Status
}

var WorkspaceIndexes = []repos.IndexField{
	{
		Field: []repos.IndexKey{
			{Key: fields.Id, Value: repos.IndexAsc},
		},
		Unique: true,
	},
	{
		Field: []repos.IndexKey{
			{
				Key:   fields.MetadataName,
				Value: repos.IndexAsc,
			},
			{
				Key:   fields.AccountName,
				Value: repos.IndexAsc,
			},
			{
				Key:   fields.ClusterName,
				Value: repos.IndexAsc,
			},
		},
		Unique: true,
	},
}
