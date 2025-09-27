package entities

import (
	"time"

	"github.com/kloudlite/api/pkg/repos"
)

type InvitationStatus string

const (
	InvitationStatusAccepted InvitationStatus = "accepted"
	InvitationStatusRejected InvitationStatus = "rejected"
	InvitationStatusNone     InvitationStatus = "none"
	InvitationStatusSend     InvitationStatus = "sent"
)

type UserMetadata map[string]any

type ProviderDetail struct {
	TokenId repos.ID `json:"token_id" bson:"token_id"`
	Avatar  *string  `json:"avatar" bson:"avatar"`
}

type User struct {
	repos.BaseEntity `json:",inline"`
	Name             string           `json:"name"`
	Avatar           *string          `json:"avatar"`
	ProviderGithub   *ProviderDetail  `json:"provider_github"`
	ProviderGitlab   *ProviderDetail  `json:"provider_gitlab"`
	ProviderGoogle   *ProviderDetail  `json:"provider_google"`
	Email            string           `json:"email"`
	Password         string           `json:"password"`
	InvitationStatus InvitationStatus `json:"invite"`
	Verified         bool             `json:"verified"`
	Metadata         UserMetadata     `json:"metadata"`
	Joined           time.Time        `json:"joined"`
	PasswordSalt     string           `json:"password_salt"`
	Approved         bool             `json:"approved"`
}

var UserIndexes = []repos.IndexField{
	{
		Field: []repos.IndexKey{
			{Key: "id", Value: repos.IndexAsc},
		},
		Unique: true,
	},
	{
		Field: []repos.IndexKey{
			{Key: "email", Value: repos.IndexAsc},
		},
		Unique: true,
	},
}

type VerifyToken struct {
	Token  string   `json:"token"`
	UserId repos.ID `json:"user_id"`
}

type ResetPasswordToken struct {
	Token  string   `json:"token"`
	UserId repos.ID `json:"user_id"`
}

