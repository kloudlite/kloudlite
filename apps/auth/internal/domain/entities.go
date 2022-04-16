package domain

import (
	"kloudlite.io/pkg/repos"
	"time"
)

type InvitationStatus string

const (
	InvitationAccepted = InvitationStatus("accepted")
	InvitationRejected = InvitationStatus("rejected")
	InvitationNone     = InvitationStatus("none")
	InvitationSent     = InvitationStatus("sent")
)

type UserMetadata map[string]any

type ProviderDetail struct {
	TokenId string `json:"token_id" bson:"token_id"`
	Avatar  string `json:"avatar" bson:"avatar"`
}

type Session struct {
	ID           repos.ID `json:"id"`
	UserID       repos.ID `json:"userId"`
	UserEmail    string   `json:"userEmail"`
	LoginMethod  string   `json:"loginMethod"`
	UserVerified bool     `json:"userVerified"`
}

type User struct {
	repos.BaseEntity `bson:",inline"`
	Name             string           `json:"name"`
	Avatar           *string          `json:"avatar"`
	ProviderGithub   *ProviderDetail  `json:"provider_github",bson:"provider_github"`
	ProviderGitlab   *ProviderDetail  `json:"provider_gitlab" bson:"provider_gitlab"`
	ProviderGoogle   *ProviderDetail  `json:"provider_github",bson:"provider_github"`
	Email            string           `json:"email"`
	Password         string           `json:"password"`
	InvitationStatus InvitationStatus `json:"invite"`
	Verified         bool             `json:"verified"`
	Metadata         UserMetadata     `json:"metadata"`
	Joined           time.Time        `json:"joined"`
	PasswordSalt     string           `json:"password_salt"`
}

var UserIndexes = []string{"email", "id"}

type AccessToken struct {
	repos.BaseEntity `bson:",inline"`
	UserId           *string        `json:"user_id" bson:"user_id"`
	Email            *string        `json:"email" bson:"email"`
	Provider         string         `json:"provider" bson:"provider"`
	Token            string         `json:"token" bson:"token"`
	Data             map[string]any `json:"data" bson:"data"`
}

var AccessTokenIndexes = []string{"user_id", "id"}

type InviteToken struct {
	Token  string `json:"token"`
	UserId string `json:"user_id"`
}

type VerifyToken struct {
	Token  string `json:"token"`
	UserId string `json:"user_id"`
}

type ResetPasswordToken struct {
	Token  string `json:"token"`
	UserId string `json:"user_id"`
}

type ChangeEmailToken struct {
	Token  string `json:"token"`
	UserId string `json:"user_id"`
}
