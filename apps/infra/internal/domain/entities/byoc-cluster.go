package entities

import (
	clusterv1 "github.com/kloudlite/operator/apis/clusters/v1"
	"kloudlite.io/pkg/repos"
	t "kloudlite.io/pkg/types"
)

type HelmStatusVal struct {
	IsReady *bool  `json:"isReady"`
	Message string `json:"message"`
}

type BYOCCluster struct {
	repos.BaseEntity `bson:",inline" json:",inline"`
	clusterv1.BYOC   `json:",inline"`
	IsConnected      bool                     `json:"isConnected"`
	SyncStatus       t.SyncStatus             `json:"syncStatus"`
	HelmStatus       map[string]HelmStatusVal `json:"helmStatus"`
}

var BYOCClusterIndices = []repos.IndexField{
	{
		Field: []repos.IndexKey{
			{Key: "id", Value: repos.IndexAsc},
		},
		Unique: true,
	},
	{
		Field: []repos.IndexKey{
			{Key: "metadata.name", Value: repos.IndexAsc},
			{Key: "spec.accountName", Value: repos.IndexAsc},
			{Key: "spec.region", Value: repos.IndexAsc},
			{Key: "spec.provider", Value: repos.IndexAsc},
		},
		Unique: true,
	},
}
