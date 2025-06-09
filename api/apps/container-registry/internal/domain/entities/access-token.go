package entities

import (
	"github.com/kloudlite/api/pkg/repos"
	"golang.org/x/oauth2"
)

type AccessToken struct {
	Id       repos.ID       `json:"_id"`
	UserId   repos.ID       `json:"user_id" bson:"user_id"`
	Email    string         `json:"email" bson:"email"`
	Provider string         `json:"provider" bson:"provider"`
	Token    *oauth2.Token  `json:"token" bson:"token"`
	Data     map[string]any `json:"data" bson:"data"`
}
