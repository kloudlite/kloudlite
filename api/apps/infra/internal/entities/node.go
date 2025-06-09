package entities

import (
	"github.com/kloudlite/api/pkg/repos"
	t "github.com/kloudlite/api/pkg/types"
	clustersv1 "github.com/kloudlite/operator/apis/clusters/v1"
)

type Node struct {
	repos.BaseEntity `json:",inline" graphql:"noinput"`
	clustersv1.Node  `json:",inline" graphql:"uri=k8s://nodes.clusters.kloudlite.io"`
	AccountName      string       `json:"accountName" graphql:"noinput"`
	ClusterName      string       `json:"clusterName" graphql:"noinput"`
	SyncStatus       t.SyncStatus `json:"syncStatus" graphql:"noinput"`
}

var NodeIndices = []repos.IndexField{
	{
		Field: []repos.IndexKey{
			{Key: "id", Value: repos.IndexAsc},
		},
		Unique: true,
	},
	{
		Field: []repos.IndexKey{
			{Key: "metadata.name", Value: repos.IndexAsc},
			{Key: "accountName", Value: repos.IndexAsc},
			{Key: "clusterName", Value: repos.IndexAsc},
		},
		Unique: true,
	},
}
