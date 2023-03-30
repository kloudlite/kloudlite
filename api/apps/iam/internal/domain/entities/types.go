package entities

import "kloudlite.io/pkg/repos"

type ResourceType string

const (
	AccountResource ResourceType = "account"
	ProjectResource ResourceType = "project"
)

type Role string

const (
	RoleAccountOwner  Role = "account-owner"
	RoleAccountAdmin  Role = "account-admin"
	RoleAccountMember Role = "account-member"

	RoleProjectAdmin   Role = "project-admin"
	RoleProjectMember Role = "project-member"
)

type RoleBinding struct {
	repos.BaseEntity `json:",inline" bson:",inline"`
	UserId           string       `json:"user_id" bson:"user_id"`
	ResourceType     ResourceType `json:"resource_type" bson:"resource_type"`
	ResourceRef      string       `json:"resource_ref" bson:"resource_ref"`
	Role             Role         `json:"role" bson:"role"`
	Accepted         bool         `json:"accepted" bson:"accepted"`
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
			{Key: "resource_ref", Value: repos.IndexDesc},
		},
		Unique: true,
	},
	{
		Field: []repos.IndexKey{
			{Key: "role", Value: repos.IndexAsc},
		},
	},
}
