package entities

import (
	"github.com/kloudlite/api/common"
	"github.com/kloudlite/api/pkg/repos"
	t "github.com/kloudlite/api/pkg/types"
	clusterv1 "github.com/kloudlite/operator/apis/clusters/v1"
)

type HelmStatusVal struct {
	IsReady *bool  `json:"isReady"`
	Message string `json:"message"`
}

type BYOCCluster struct {
	repos.BaseEntity `json:",inline" graphql:"noinput"`
	clusterv1.BYOC   `json:",inline" graphql:"uri=k8s://byocs.clusters.kloudlite.io"`

	IncomingKafkaTopicName  string `json:"incomingKafkaTopicName" graphql:"noinput"`
	common.ResourceMetadata `json:",inline"`

	IsConnected bool                     `json:"isConnected" graphql:"noinput"`
	HelmStatus  map[string]HelmStatusVal `json:"helmStatus" graphql:"noinput"`

	AccountName string `json:"accountName"`

	SyncStatus t.SyncStatus `json:"syncStatus" graphql:"noinput"`
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
			{Key: "accountName", Value: repos.IndexAsc},
			{Key: "spec.region", Value: repos.IndexAsc},
			{Key: "spec.provider", Value: repos.IndexAsc},
		},
		Unique: true,
	},
}
