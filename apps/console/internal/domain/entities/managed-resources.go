package entities

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"kloudlite.io/pkg/repos"
)

type ManagedResourceStatus string

const (
	ManagedResourceStateSyncing = ManagedResourceStatus("sync-in-progress")
	ManagedResourceStateLive    = ManagedResourceStatus("live")
	ManagedResourceStateError   = ManagedResourceStatus("error")
	ManagedResourceStateDown    = ManagedResourceStatus("down")
)

type ManagedResource struct {
	repos.BaseEntity `bson:",inline"`
	ClusterId        repos.ID              `json:"cluster_id" bson:"cluster_id"`
	ProjectId        repos.ID              `json:"project_id" bson:"project_id"`
	Name             string                `json:"name" bson:"name"`
	Namespace        string                `json:"namespace" bson:"namespace"`
	ServiceType      ManagedResourceType   `json:"resource_type" bson:"resource_type"`
	Service          string                `bson:"service_name" json:"service_name"`
	Values           map[string]any        `json:"values" bson:"values"`
	Status           ManagedResourceStatus `json:"status" bson:"status"`
	Conditions       []metav1.Condition    `json:"conditions" bson:"conditions"`
}

var ManagedResourceIndexes = []repos.IndexField{
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
