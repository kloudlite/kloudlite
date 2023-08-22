package entities

import (
	t "kloudlite.io/apps/iam/types"
	"kloudlite.io/pkg/repos"
)

type RoleBinding struct {
	repos.BaseEntity `json:",inline" bson:",inline"`
	UserId           string         `json:"user_id" bson:"user_id"`
	ResourceType     t.ResourceType `json:"resource_type" bson:"resource_type"`
	ResourceRef      string         `json:"resource_ref" bson:"resource_ref"`
	Role             t.Role         `json:"role" bson:"role"`
	// Accepted         bool           `json:"accepted" bson:"accepted"`
}

var RoleBindingIndices = []repos.IndexField{
	{
		Field: []repos.IndexKey{
			{Key: "id", Value: repos.IndexAsc},
		},
		Unique: true,
	},
	{
		Field: []repos.IndexKey{
			{Key: "user_id", Value: repos.IndexDesc},
			{Key: "resource_ref", Value: repos.IndexDesc},
		},
		Unique: true,
	},
	{
		Field: []repos.IndexKey{
			{Key: "role", Value: repos.IndexAsc},
			{Key: "resource_type", Value: repos.IndexAsc},
		},
	},
}
