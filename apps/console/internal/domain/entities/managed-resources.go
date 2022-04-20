package entities

import "kloudlite.io/pkg/repos"

type ManagedResourceStatus string

const (
	ManagedResourceStateSyncing = ManagedResourceStatus("sync-in-progress")
	ManagedResourceStateLive    = ManagedResourceStatus("live")
	ManagedResourceStateError   = ManagedResourceStatus("error")
	ManagedResourceStateDown    = ManagedResourceStatus("down")
)

type ManagedResource struct {
	repos.BaseEntity `bson:",inline"`
	Name             string                `bson:"name" json:"name"`
	ServiceType      ManagedResourceType   `json:"resource_type" bson:"resource_type"`
	Service          string                `bson:"service_name" json:"service_name"`
	Values           map[string]any        `json:"values" bson:"values"`
	Status           ManagedResourceStatus `json:"status" bson:"status"`
}
