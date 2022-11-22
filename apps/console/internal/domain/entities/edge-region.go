package entities

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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

type CloudProviderStatus string

const (
	CloudProviderStateSyncing = CloudProviderStatus("sync-in-progress")
	CloudProviderStateLive    = CloudProviderStatus("live")
	CloudProviderStateError   = CloudProviderStatus("error")
	CloudProviderStateDown    = CloudProviderStatus("down")
)

type CloudProvider struct {
	repos.BaseEntity `bson:",inline"`
	Name             string              `bson:"name"`
	AccountId        *repos.ID           `json:"account_id,omitempty" bson:"account_id"`
	Provider         string              `json:"provider" bson:"provider"`
	Credentials      map[string]string   `json:"credentials" bson:"credentials"`
	Status           CloudProviderStatus `json:"status" bson:"status"`
	Conditions       []metav1.Condition  `json:"conditions" bson:"conditions"`
}

var CloudProviderIndexes = []repos.IndexField{
	{
		Field: []repos.IndexKey{
			{Key: "id", Value: repos.IndexAsc},
		},
		Unique: true,
	},
	{
		Field: []repos.IndexKey{
			{Key: "account_id", Value: repos.IndexAsc},
		},
	},
	{
		Field: []repos.IndexKey{
			{Key: "provider", Value: repos.IndexAsc},
		},
	},
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
