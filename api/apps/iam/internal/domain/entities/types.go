package entities

import "kloudlite.io/pkg/repos"

type RoleBinding struct {
	repos.BaseEntity `json:",inline" bson:",inline"`
	UserId           repos.ID `json:"user_id" bson:"user_id"`
	ResourceType     string   `json:"resource_type" bson:"resource_type"`
	ResourceId       string   `json:"resource_id" bson:"resource_id"`
	Role             string   `json:"role" bson:"role"`
	Accepted         bool     `json:"accepted" bson:"accepted"`
}

// var RoleBindingIndexes = []string{"id", "user_id", "resource_id", "role"}

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
			{Key: "resource_id", Value: repos.IndexDesc},
		},
		Unique: true,
	},
	{
		Field: []repos.IndexKey{
			{Key: "role", Value: repos.IndexAsc},
		},
	},
}
