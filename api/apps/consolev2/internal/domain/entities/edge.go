package entities

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	// op_crds "kloudlite.io/apps/console/internal/domain/op-crds"
	"kloudlite.io/pkg/repos"
)

type EdgeStatus string

const (
	EdgeStateSyncing = EdgeStatus("sync-in-progress")
	EdgeStateLive    = EdgeStatus("live")
	EdgeStateError   = EdgeStatus("error")
	EdgeStateDown    = EdgeStatus("down")
)

type NodePool struct {
	Name   string   `json:"name"`
	Config string   `json:"config"`
	Min    int      `json:"min"`
	Max    int      `json:"max"`
	Nodes  []string `bson:"nodes"`
}

type EdgeRegion struct {
	repos.BaseEntity `bson:",inline"`
	IsDeleting       bool               `json:"is_deleting" bson:"is_deleting"`
	Name             string             `bson:"name"`
	ProviderId       repos.ID           `bson:"provider_id"`
	Region           string             `bson:"region"`
	Pools            []NodePool         `bson:"pools"`
	Status           EdgeStatus         `json:"status" bson:"status"`
	Conditions       []metav1.Condition `json:"conditions" bson:"conditions"`
}

var EdgeRegionIndexes = []repos.IndexField{
	{
		Field: []repos.IndexKey{
			{Key: "id", Value: repos.IndexAsc},
		},
		Unique: true,
	},
	{
		Field: []repos.IndexKey{
			{Key: "region", Value: repos.IndexAsc},
		},
		Unique: true,
	},
	{
		Field: []repos.IndexKey{
			{Key: "provider", Value: repos.IndexAsc},
		},
	},
}
