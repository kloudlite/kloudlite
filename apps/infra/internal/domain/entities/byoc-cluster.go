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
	repos.BaseEntity `json:",inline" graphql:"noinput"`
	clusterv1.BYOC   `json:",inline" graphql:"uri=k8s://byocs.clusters.kloudlite.io"`
	IsConnected      bool                     `json:"isConnected" graphql:"noinput"`
	SyncStatus       t.SyncStatus             `json:"syncStatus" graphql:"noinput"`
	HelmStatus       map[string]HelmStatusVal `json:"helmStatus" graphql:"noinput"`
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
