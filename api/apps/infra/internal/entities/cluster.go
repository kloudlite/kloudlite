package entities

import (
	clustersv1 "github.com/kloudlite/operator/apis/clusters/v1"
	"kloudlite.io/pkg/repos"
	t "kloudlite.io/pkg/types"
)

type Cluster struct {
	repos.BaseEntity   `json:",inline" graphql:"noinput"`
	clustersv1.Cluster `json:",inline" graphql:"uri=k8s://clusters.clusters.kloudlite.io"`
	AccountName        string       `json:"accountName" graphql:"noinput"`
	SyncStatus         t.SyncStatus `json:"syncStatus" graphql:"noinput"`
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
