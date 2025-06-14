package entities

import (
	"github.com/kloudlite/api/common"
	"github.com/kloudlite/api/pkg/repos"
	t "github.com/kloudlite/api/pkg/types"
	clustersv1 "github.com/kloudlite/operator/apis/clusters/v1"
	"github.com/kloudlite/operator/toolkit/reconciler"
)

type NodePool struct {
	repos.BaseEntity    `json:",inline" graphql:"noinput"`
	clustersv1.NodePool `json:",inline" graphql:"uri=k8s://nodepools.clusters.kloudlite.io"`

	common.ResourceMetadata `json:",inline"`

	AccountName string `json:"accountName" graphql:"noinput"`
	ClusterName string `json:"clusterName" graphql:"noinput"`

	DispatchAddr *DispatchAddr `json:"dispatchAddr" graphql:"noinput"`

	SyncStatus t.SyncStatus `json:"syncStatus" graphql:"noinput"`
}

func (n *NodePool) GetDisplayName() string {
	return n.ResourceMetadata.DisplayName
}

func (n *NodePool) GetStatus() reconciler.Status {
	return n.NodePool.Status
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
	{
		Field: []repos.IndexKey{
			{Key: "metadata.name", Value: repos.IndexAsc},
		},
	},
}
