package entities

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"kloudlite.io/pkg/repos"
)

type RouterStatus string

const (
	RouteStateSyncing = RouterStatus("sync-in-progress")
	RouteStateLive    = RouterStatus("live")
	RouteStateError   = RouterStatus("error")
	RouteStateDown    = RouterStatus("down")
)

type Router struct {
	repos.BaseEntity `bson:",inline"`
	ClusterId        repos.ID           `json:"cluster_id" bson:"cluster_id"`
	ProjectId        repos.ID           `json:"project_id" bson:"project_id"`
	Name             string             `json:"name" bson:"name"`
	Namespace        string             `json:"namespace" bson:"namespace"`
	Domains          []string           `bson:"domains" json:"domains"`
	Routes           []*Route           `bson:"routes" json:"routes"`
	Status           RouterStatus       `json:"status" bson:"status"`
	Conditions       []metav1.Condition `json:"conditions" bson:"conditions"`
}

type Route struct {
	Path    string `bson:"path" json:"path"`
	AppName string `bson:"app" json:"app"`
	Port    uint16 `bson:"port" json:"port"`
}

var RouterIndexes = []repos.IndexField{
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
