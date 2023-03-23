package entities

import (
	cmgrV1 "github.com/kloudlite/cluster-operator/apis/cmgr/v1"
	"kloudlite.io/pkg/repos"
	t "kloudlite.io/pkg/types"
)

type Cluster struct {
	repos.BaseEntity `bson:",inline" json:",inline"`
	cmgrV1.Cluster   `json:",inline"`
	AccountName      string `json:"accountName"`

	SyncStatus t.SyncStatus `json:"syncStatus"`
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
