package entities

import "kloudlite.io/pkg/repos"

type ClusterStatus string

const (
	ClusterStateSyncing = ClusterStatus("sync-in-progress")
	ClusterStateLive    = ClusterStatus("live")
	ClusterStateError   = ClusterStatus("error")
	ClusterStateDown    = ClusterStatus("down")
)

type Cluster struct {
	repos.BaseEntity `bson:",inline"`
	Name             string        `json:"name" bson:"name"`
	Provider         string        `json:"provider" bson:"provider"`
	Region           string        `json:"region" bson:"region"`
	Ip               *string       `json:"ip,omitempty" bson:"ip,omitempty"`
	PublicKey        *string       `json:"public_key,omitempty" bson:"public_key,omitempty"`
	NodesCount       int           `json:"nodes_count" bson:"nodes_count"`
	Status           ClusterStatus `json:"status" bson:"status"`
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
