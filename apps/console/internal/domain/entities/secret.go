package entities

import "kloudlite.io/pkg/repos"

type SecretStatus string

const (
	SecretStateSyncing = SecretStatus("sync-in-progress")
	SecretStateLive    = SecretStatus("live")
	SecretStateError   = SecretStatus("error")
	SecretStateDown    = SecretStatus("down")
)

type Secret struct {
	repos.BaseEntity `bson:",inline"`
	ProjectId        repos.ID     `json:"project_id" bson:"project_id"`
	Name             string       `json:"name" bson:"name"`
	Namespace        string       `json:"namespace" bson:"namespace"`
	Description      *string      `json:"description" bson:"description"`
	Data             []*Entry     `json:"data" bson:"data"`
	Status           SecretStatus `json:"status" bson:"status"`
}

var SecretIndexes = []repos.IndexField{
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
