package entities

import "kloudlite.io/pkg/repos"

type ProjectStatus string

const (
	ProjectStateSyncing = ProjectStatus("sync-in-progress")
	ProjectStateLive    = ProjectStatus("live")
	ProjectStateError   = ProjectStatus("error")
	ProjectStateDown    = ProjectStatus("down")
)

type Project struct {
	repos.BaseEntity `bson:",inline"`
	AccountId        repos.ID      `json:"account_id" bson:"account_id"`
	Name             string        `json:"name" bson:"name"`
	DisplayName      string        `json:"display_name" bson:"display_name"`
	Description      *string       `json:"description" bson:"description"`
	Logo             *string       `json:"logo" bson:"logo"`
	ReadableId       repos.ID      `json:"readable_id" bson:"readable_id"`
	Status           ProjectStatus `json:"status" bson:"status"`
}

var ProjectIndexes = []repos.IndexField{
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
