package entities

import (
	iamT "kloudlite.io/apps/iam/types"
	"kloudlite.io/pkg/repos"
)

type AccountMembership struct {
	AccountName string    `json:"accountName"`
	UserId      repos.ID  `json:"userId"`
	Role        iamT.Role `json:"role"`
}
