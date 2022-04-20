package entities

import "kloudlite.io/pkg/repos"

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
	ProjectId        repos.ID     `json:"project_id" bson:"project_id"`
	Namespace        string       `json:"namespace" bson:"namespace"`
	Description      *string      `json:"description" bson:"description"`
	Name             string       `json:"name" bson:"name"`
	Data             []*Entry     `json:"data" bson:"data"`
	Status           ConfigStatus `json:"status" bson:"status"`
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
		},
		Unique: true,
	},
}
