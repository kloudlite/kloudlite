package entities

import "kloudlite.io/pkg/repos"

type RouterStatus string

const (
	RouteStateSyncing = RouterStatus("sync-in-progress")
	RouteStateLive    = RouterStatus("live")
	RouteStateError   = RouterStatus("error")
	RouteStateDown    = RouterStatus("down")
)

type Router struct {
	repos.BaseEntity `bson:",inline"`
	ProjectId        repos.ID `json:"project_id" bson:"project_id"`
	Name             string   `bson:"name" json:"name"`
	Namespace        string   `bson:"namespace" json:"namespace"`
	Domains          []string `bson:"domains" json:"domains"`
	Routes           []*Route `bson:"routes" json:"routes"`
}

type Route struct {
	Path    string `bson:"path" json:"path"`
	AppName string `bson:"app" json:"app"`
	Port    string `bson:"port" json:"port"`
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
		},
		Unique: true,
	},
}
