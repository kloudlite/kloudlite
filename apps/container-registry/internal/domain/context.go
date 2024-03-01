package domain

import (
	"context"

	"github.com/kloudlite/api/pkg/repos"
)

type RegistryContext struct {
	context.Context
	UserId      repos.ID
	UserName    string
	AccountName string
	UserEmail   string
}

func (c RegistryContext) GetAccountName() string {
	return c.AccountName
}

func (c RegistryContext) GetUserId() repos.ID {
	return c.UserId
}
func (c RegistryContext) GetUserEmail() string {
	return c.UserEmail
}
func (c RegistryContext) GetUserName() string {
	return c.UserName
}
