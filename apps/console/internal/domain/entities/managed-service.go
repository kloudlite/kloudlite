package entities

import "kloudlite.io/pkg/repos"

type ManagedServiceStatus string

const (
	ManagedServiceStateSyncing = ManagedServiceStatus("sync-in-progress")
	ManagedServiceStateLive    = ManagedServiceStatus("live")
	ManagedServiceStateError   = ManagedServiceStatus("error")
	ManagedServiceStateDown    = ManagedServiceStatus("down")
)

type ManagedService struct {
	repos.BaseEntity `bson:",inline"`
	Name             string               `bson:"name" json:"name"`
	Namespace        string               `bson:"namespace" json:"namespace"`
	ServiceType      ManagedServiceType   `json:"service_type" bson:"service_type"`
	Values           map[string]any       `json:"values" bson:"values"`
	Status           ManagedServiceStatus `json:"status" bson:"status"`
}
