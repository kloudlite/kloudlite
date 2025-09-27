package entities

import (
	"time"

	"github.com/kloudlite/api/common"
	"github.com/kloudlite/api/pkg/repos"
)

type PlatformInvitation struct {
	repos.BaseEntity     `json:",inline" bson:",inline"`
	common.ResourceMetadata `json:",inline" bson:",inline"`
	Email            string    `json:"email" bson:"email"`
	Role             string    `json:"role" bson:"role"` // "super_admin", "admin", "user"
	InvitedBy        string    `json:"invitedBy" bson:"invitedBy"`
	InvitedByEmail   string    `json:"invitedByEmail" bson:"invitedByEmail"`
	Status           string    `json:"status" bson:"status"` // "pending", "accepted", "expired", "cancelled"
	Token            string    `json:"token" bson:"token"`
	ExpiresAt        time.Time `json:"expiresAt" bson:"expiresAt"`
	AcceptedAt       *time.Time `json:"acceptedAt,omitempty" bson:"acceptedAt,omitempty"`
	AcceptedBy       *string    `json:"acceptedBy,omitempty" bson:"acceptedBy,omitempty"`
}

var PlatformInvitationIndices = []repos.IndexField{
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
			{Key: "email", Value: repos.IndexAsc},
			{Key: "status", Value: repos.IndexAsc},
		},
	},
	{
		Field: []repos.IndexKey{
			{Key: "invitedBy", Value: repos.IndexAsc},
		},
	},
	{
		Field: []repos.IndexKey{
			{Key: "expiresAt", Value: repos.IndexAsc},
		},
	},
}