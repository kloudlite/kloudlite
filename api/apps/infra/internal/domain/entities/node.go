package entities

import (
	cmgrV1 "github.com/kloudlite/cluster-operator/apis/cmgr/v1"
	infraV1 "github.com/kloudlite/cluster-operator/apis/infra/v1"
	"kloudlite.io/pkg/repos"
	t "kloudlite.io/pkg/types"
)

type MasterNode struct {
	repos.BaseEntity  `json:",inline" graphql:"noinput"`
	cmgrV1.MasterNode `json:",inline" graphql:"uri=k8s://masternodes.cmgr.kloudlite.io"`
	AccountName       string       `json:"accountName" graphql:"noinput"`
	ClusterName       string       `json:"clusterName" graphql:"noinput"`
	SyncStatus        t.SyncStatus `json:"syncStatus" graphql:"noinput"`
}

var MasterNodeIndices = []repos.IndexField{
	{
		Field: []repos.IndexKey{
			{Key: "id", Value: repos.IndexAsc},
		},
		Unique: true,
	},
	{
		Field: []repos.IndexKey{
			{Key: "metadata.name", Value: repos.IndexAsc},
			{Key: "spec.clusterName", Value: repos.IndexAsc},
		},
		Unique: true,
	},
	{
		Field: []repos.IndexKey{
			{Key: "accountName", Value: repos.IndexAsc},
		},
	},
	{
		Field: []repos.IndexKey{
			{Key: "clusterName", Value: repos.IndexAsc},
		},
	},
}

type WorkerNode struct {
	repos.BaseEntity   `json:",inline" graphql:"noinput"`
	infraV1.WorkerNode `json:",inline" graphql:"uri=k8s://workernodes.infra.kloudlite.io"`
	AccountName        string       `json:"accountName" graphql:"noinput"`
	ClusterName        string       `json:"clusterName" graphql:"noinput"`
	SyncStatus         t.SyncStatus `json:"syncStatus" graphql:"noinput"`
}

var WorkerNodeIndices = []repos.IndexField{
	{
		Field: []repos.IndexKey{
			{Key: "id", Value: repos.IndexAsc},
		},
		Unique: true,
	},
	{
		Field: []repos.IndexKey{
			{Key: "metadata.name", Value: repos.IndexAsc},
			{Key: "spec.clusterName", Value: repos.IndexAsc},
		},
		Unique: true,
	},
	{
		Field: []repos.IndexKey{
			{Key: "accountName", Value: repos.IndexAsc},
		},
	},
	{
		Field: []repos.IndexKey{
			{Key: "clusterName", Value: repos.IndexAsc},
		},
	},
}

type NodePool struct {
	repos.BaseEntity `json:",inline" graphql:"noinput"`
	infraV1.NodePool `json:",inline" graphql:"uri=k8s://nodepools.infra.kloudlite.io"`
	AccountName      string       `json:"accountName" graphql:"noinput"`
	ClusterName      string       `json:"clusterName" graphql:"noinput"`
	SyncStatus       t.SyncStatus `json:"syncStatus" graphql:"noinput"`
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
			{Key: "spec.edgeName", Value: repos.IndexAsc},
			{Key: "spec.clusterName", Value: repos.IndexAsc},
		},
		Unique: true,
	},
	{
		Field: []repos.IndexKey{
			{Key: "accountName", Value: repos.IndexAsc},
		},
	},
	{
		Field: []repos.IndexKey{
			{Key: "clusterName", Value: repos.IndexAsc},
		},
	},
}
