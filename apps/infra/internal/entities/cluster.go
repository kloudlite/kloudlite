package entities

import (
	"github.com/kloudlite/api/common"
	"github.com/kloudlite/api/pkg/repos"
	t "github.com/kloudlite/api/pkg/types"
	clustersv1 "github.com/kloudlite/operator/apis/clusters/v1"
	"github.com/kloudlite/operator/pkg/operator"
)

type Cluster struct {
	repos.BaseEntity `json:",inline" graphql:"noinput"`

	clustersv1.Cluster `json:",inline"`

	common.ResourceMetadata `json:",inline"`

	ClusterGroupName *string      `json:"clusterGroupName"`
	AccountName      string       `json:"accountName" graphql:"noinput"`
	SyncStatus       t.SyncStatus `json:"syncStatus" graphql:"noinput"`
}

func (c *Cluster) GetDisplayName() string {
	return c.ResourceMetadata.DisplayName
}

func (c *Cluster) GetStatus() operator.Status {
	return c.Cluster.Status
}

var ClusterIndices = []repos.IndexField{
	{
		Field: []repos.IndexKey{
			{Key: "id", Value: repos.IndexAsc},
		},
		Unique: true,
	},
	{
		Field: []repos.IndexKey{
			{Key: "metadata.name", Value: repos.IndexAsc},
			{Key: "metadata.namespace", Value: repos.IndexAsc},
		},
		Unique: true,
	},
	{
		Field: []repos.IndexKey{
			{Key: "accountName", Value: repos.IndexAsc},
		},
	},
}
