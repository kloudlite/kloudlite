package entities

import (
	"github.com/kloudlite/api/common"
	"github.com/kloudlite/api/pkg/repos"
	t "github.com/kloudlite/api/pkg/types"
	clustersv1 "github.com/kloudlite/operator/apis/clusters/v1"
)

type Cluster struct {
	repos.BaseEntity `json:",inline" graphql:"noinput"`

	// clustersv1.Cluster `json:",inline" graphql:"uri=k8s://clusters.clusters.kloudlite.io"`
	clustersv1.Cluster `json:",inline"`

	common.ResourceMetadata `json:",inline"`

	AccountName string       `json:"accountName" graphql:"noinput"`
	SyncStatus  t.SyncStatus `json:"syncStatus" graphql:"noinput"`
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
