package entities

import (
	cmgrV1 "github.com/kloudlite/cluster-operator/apis/cmgr/v1"
	"kloudlite.io/pkg/repos"
	t "kloudlite.io/pkg/types"
)

type Cluster struct {
	repos.BaseEntity `json:",inline" graphql:"noinput"`
	cmgrV1.Cluster   `json:",inline" graphql:"uri=k8s://clusters.cmgr.kloudlite.io"`
	AccountName      string       `json:"accountName"`
	SyncStatus       t.SyncStatus `json:"syncStatus" graphql:"noinput"`
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
		},
		Unique: true,
	},
	{
		Field: []repos.IndexKey{
			{Key: "accountName", Value: repos.IndexAsc},
		},
	},
}
