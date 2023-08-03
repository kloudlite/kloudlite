package entities

import (
	clustersv1 "github.com/kloudlite/operator/apis/clusters/v1"
	"kloudlite.io/pkg/repos"
	t "kloudlite.io/pkg/types"
)

type NodePool struct {
	repos.BaseEntity    `json:",inline" graphql:"noinput"`
	clustersv1.NodePool `json:",inline" graphql:"uri=k8s://nodepools.clusters.kloudlite.io"`
	AccountName         string       `json:"accountName" graphql:"noinput"`
	ClusterName         string       `json:"clusterName" graphql:"noinput"`
	SyncStatus          t.SyncStatus `json:"syncStatus" graphql:"noinput"`
}

var NodePoolIndices = []repos.IndexField{
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
