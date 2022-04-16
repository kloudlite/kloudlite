package entities

import "kloudlite.io/pkg/repos"

type RoleBinding struct {
	repos.BaseEntity `json:",inline" bson:",inline"`
	UserId           string `json:"user_id" bson:"user_id"`
	ResourceType     string `json:"resource_type" bson:"resource_type"`
	ResourceId       string `json:"resource_id" bson:"resource_id"`
	Role             string `json:"role" bson:"role"`
}

var RoleBindingIndexes = []string{"id", "user_id", "resource_id", "role"}
