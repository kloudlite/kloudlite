package entities

import (
	iamT "kloudlite.io/apps/iam/types"
	"kloudlite.io/pkg/repos"
)

type Invitation struct {
	repos.BaseEntity `json:",inline" graphql:"noinput"`
	InvitedBy        string    `json:"invitedBy"`
	Token            string    `json:"token"`
	UserId           repos.ID  `json:"userId"`
	Role             iamT.Role `json:"role"`
	AccountName      string    `json:"accountName"`
}

var InvitationIndices = []repos.IndexField{
	{
		Field: []repos.IndexKey{
			{Key: "id", Value: repos.IndexAsc},
		},
		Unique: true,
	},
	{
		Field: []repos.IndexKey{
			{Key: "token", Value: repos.IndexAsc},
		},
		Unique: true,
	},
	{
		Field: []repos.IndexKey{
			{Key: "userId", Value: repos.IndexAsc},
			{Key: "accountName", Value: repos.IndexAsc},
		},
		Unique: true,
	},
}
