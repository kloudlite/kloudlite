package entities

import (
	"github.com/kloudlite/api/common"
	"github.com/kloudlite/api/pkg/repos"
)

type PlatformRole string

const (
	PlatformRoleSuperAdmin PlatformRole = "super_admin"
	PlatformRoleAdmin      PlatformRole = "admin"
	PlatformRoleUser       PlatformRole = "user"
)

type PlatformUser struct {
	repos.BaseEntity        `json:",inline" graphql:"noinput"`
	common.ResourceMetadata `json:",inline"`

	UserId   repos.ID     `json:"userId"`
	Role     PlatformRole `json:"role"`
}

var PlatformUserIndices = []repos.IndexField{
	{
		Field: []repos.IndexKey{
			{Key: "userId", Value: repos.IndexAsc},
		},
		Unique: true,
	},
}