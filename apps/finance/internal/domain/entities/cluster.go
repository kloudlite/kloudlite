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
	PublicKey        *string       `json:"publicKey,omitempty" bson:"publicKey,omitempty"`
	NodesCount       int           `json:"nodes_count" bson:"nodes_count"`
	Status           ClusterStatus `json:"status" bson:"status"`
}
