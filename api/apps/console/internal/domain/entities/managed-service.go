package entities

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"kloudlite.io/pkg/repos"
)

type ManagedServiceStatus string

const (
	ManagedServiceStateSyncing  = ManagedServiceStatus("sync-in-progress")
	ManagedServiceStateDeleting = ManagedServiceStatus("deleting")
	ManagedServiceStateLive     = ManagedServiceStatus("live")
	ManagedServiceStateError    = ManagedServiceStatus("error")
	ManagedServiceStateDown     = ManagedServiceStatus("down")
)

type ManagedService struct {
	repos.BaseEntity `bson:",inline"`
	ClusterId        repos.ID             `json:"cluster_id" bson:"cluster_id"`
	ProjectId        repos.ID             `json:"project_id" bson:"project_id"`
	Name             string               `json:"name" bson:"name"`
	Namespace        string               `json:"namespace" bson:"namespace"`
	ServiceType      ManagedServiceType   `json:"service_type" bson:"service_type"`
	Values           map[string]any       `json:"values" bson:"values"`
	Status           ManagedServiceStatus `json:"status" bson:"status"`
	Conditions       []metav1.Condition   `json:"conditions" bson:"conditions"`
}

var ManagedServiceIndexes = []repos.IndexField{
	{
		Field: []repos.IndexKey{
			{Key: "id", Value: repos.IndexAsc},
		},
		Unique: true,
	},
	{
		Field: []repos.IndexKey{
			{Key: "name", Value: repos.IndexAsc},
			{Key: "namespace", Value: repos.IndexAsc},
			{Key: "cluster_id", Value: repos.IndexAsc},
		},
		Unique: true,
	},
}
