package domain

import (
	"context"

	"github.com/kloudlite/api/pkg/repos"
)

type CommsContext struct {
	context.Context
	UserId      repos.ID
	UserName    string
	AccountName string
	UserEmail   string
}

func (c CommsContext) GetAccountName() string {
	return c.AccountName
}

func (c CommsContext) GetUserId() repos.ID {
	return c.UserId
}
func (c CommsContext) GetUserEmail() string {
	return c.UserEmail
}
func (c CommsContext) GetUserName() string {
	return c.UserName
}
