package entities

import (
	iamT "github.com/kloudlite/api/apps/iam/types"
	"github.com/kloudlite/api/pkg/repos"
)

type AccountMembership struct {
	AccountName string    `json:"accountName"`
	UserId      repos.ID  `json:"userId"`
	Role        iamT.Role `json:"role"`
}
