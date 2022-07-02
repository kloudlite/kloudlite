package entities

import "kloudlite.io/pkg/repos"

type ClusterStatus string

const (
	ClusterStateSyncing = ClusterStatus("sync-in-progress")
	ClusterStateLive    = ClusterStatus("live")
	ClusterStateError   = ClusterStatus("error")
	ClusterStateDown    = ClusterStatus("down")
)

type ClusterType string

const (
	ClusterTypeKubernetes = ClusterType("kubernetes")
)

type Cluster struct {
	repos.BaseEntity `bson:",inline"`
	Name             string        `json:"name" bson:"name"`
	Provider         string        `json:"provider" bson:"provider"`
	Region           string        `json:"region" bson:"region"`
	Status           ClusterStatus `json:"status" bson:"status"`
	ClusterType      ClusterType   `json:"cluster_type" bson:"cluster_type"`
	ClusterIps       []string      `json:"cluster_ips" bson:"cluster_ips"`
}

var ClusterIndexes = []repos.IndexField{
	{
		Field: []repos.IndexKey{
			{Key: "id", Value: repos.IndexAsc},
		},
		Unique: true,
	},
	{
		Field: []repos.IndexKey{
			{Key: "provider", Value: repos.IndexAsc},
		},
		Unique: false,
	},
	{
		Field: []repos.IndexKey{
			{Key: "region", Value: repos.IndexAsc},
		},
		Unique: false,
	},
	{
		Field: []repos.IndexKey{
			{Key: "ip", Value: repos.IndexAsc},
		},
		Unique: true,
	},
	{
		Field: []repos.IndexKey{
			{Key: "account_id", Value: repos.IndexAsc},
		},
	},
}
