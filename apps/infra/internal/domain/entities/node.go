package entities

import (
	cmgrV1 "github.com/kloudlite/cluster-operator/apis/cmgr/v1"
	infraV1 "github.com/kloudlite/cluster-operator/apis/infra/v1"
	"kloudlite.io/pkg/repos"
)

type MasterNode struct {
	repos.BaseEntity  `json:",inline"`
	cmgrV1.MasterNode `json:",inline"`
	ClusterName       string `json:"clusterName"`
	AccountName       string `json:"accountName"`
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
	repos.BaseEntity   `json:",inline"`
	infraV1.WorkerNode `json:",inline"`
	ClusterName        string `json:"clusterName"`
	AccountName        string `json:"accountName"`
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
	repos.BaseEntity `json:",inline"`
	infraV1.NodePool `json:",inline"`
	AccountName      string `json:"accoutName"`
	ClusterName      string `json:"clusterName"`
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
