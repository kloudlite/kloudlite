package entities

import (
	"encoding/json"
	"io"

	crdsv1 "github.com/kloudlite/operator/apis/crds/v1"
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
	crdsv1.Router    `json:",inline" bson:",inline"`
	//ClusterId        repos.ID           `json:"cluster_id" bson:"cluster_id"`
	//ProjectId        repos.ID           `json:"project_id" bson:"project_id"`
	//Name             string             `json:"name" bson:"name"`
	//Namespace        string             `json:"namespace" bson:"namespace"`
	//Domains          []string           `bson:"domains" json:"domains"`
	//Routes           []*Route           `bson:"routes" json:"routes"`
	//Status           RouterStatus       `json:"status" bson:"status"`
	//Conditions       []metav1.Condition `json:"conditions" bson:"conditions"`
}

func (r *Router) UnmarshalGQL(v interface{}) error {
	if err := json.Unmarshal([]byte(v.(string)), r); err != nil {
		return err
	}

	// if err := validator.Validate(*c); err != nil {
	//  return err
	// }

	return nil
}

func (r Router) MarshalGQL(w io.Writer) {
	b, err := json.Marshal(r)
	if err != nil {
		w.Write([]byte("{}"))
	}
	w.Write(b)
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
			{Key: "metadata.name", Value: repos.IndexAsc},
			{Key: "metadata.namespace", Value: repos.IndexAsc},
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
