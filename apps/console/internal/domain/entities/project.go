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
	AccountId        string `json:"account_id" bson:"account_id"`
	Name             string `json:"name" bson:"name"`
	DisplayName      string `json:"display_name" bson:"display_name"`
	Description      string `json:"description" bson:"description"`
	Logo             string `json:"logo" bson:"logo"`
}
