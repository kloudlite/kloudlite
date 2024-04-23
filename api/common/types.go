package common

import (
	"encoding/json"

	"github.com/kloudlite/api/pkg/repos"
)

type CreatedOrUpdatedBy struct {
	UserId    repos.ID `json:"userId"`
	UserName  string   `json:"userName"`
	UserEmail string   `json:"userEmail"`
}

type ResourceMetadata struct {
	DisplayName string `json:"displayName"`

	CreatedBy     CreatedOrUpdatedBy `json:"createdBy" graphql:"noinput"`
	LastUpdatedBy CreatedOrUpdatedBy `json:"lastUpdatedBy" graphql:"noinput"`
}

type ValidationError struct {
	Label  string
	Errors []string
}

func (v ValidationError) Error() string {
	b, _ := json.Marshal(map[string]any{
		"label":  v.Label,
		"errors": v.Errors,
	})
	return string(b)
}

const (
	CreatedByResourceSyncUsername  string = "created-by-resource-sync-user-name"
	CreatedByResourceSyncUserEmail string = "created-by-resource-sync-user-email"
	CreatedByResourceSyncUserId    string = "created-by-resource-sync-user-id"
)

var CreatedOrUpdatedByResourceSync = CreatedOrUpdatedBy{
	UserId:    repos.ID(CreatedByResourceSyncUserId),
	UserName:  CreatedByResourceSyncUsername,
	UserEmail: CreatedByResourceSyncUserEmail,
}

var CreatedOrUpdatedByKloudlite = CreatedOrUpdatedBy{
	UserId:    repos.ID(""),
	UserName:  "kloudlite",
	UserEmail: "support@kloudlite.io",
}
