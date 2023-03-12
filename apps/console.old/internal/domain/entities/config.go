package entities

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"kloudlite.io/pkg/repos"
)

type ConfigStatus string

const (
	ConfigStateSyncing = ConfigStatus("sync-in-progress")
	ConfigStateLive    = ConfigStatus("live")
	ConfigStateError   = ConfigStatus("error")
	ConfigStateDown    = ConfigStatus("down")
)

type Entry struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

type Config struct {
	repos.BaseEntity `bson:",inline"`
	ClusterId        repos.ID           `json:"cluster_id" bson:"cluster_id"`
	ProjectId        repos.ID           `json:"project_id" bson:"project_id"`
	Namespace        string             `json:"namespace" bson:"namespace"`
	Description      *string            `json:"description" bson:"description"`
	Name             string             `json:"name" bson:"name"`
	Data             []*Entry           `json:"data" bson:"data"`
	Status           ConfigStatus       `json:"status" bson:"status"`
	Conditions       []metav1.Condition `json:"conditions" bson:"conditions"`
}

var ConfigIndexes = []repos.IndexField{
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
