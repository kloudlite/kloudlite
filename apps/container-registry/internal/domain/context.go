package domain

import (
	"context"
	"kloudlite.io/pkg/repos"
)

type RegistryContext struct {
	context.Context
	UserId      repos.ID
	UserName    string
	AccountName string
	UserEmail   string
}

func (c *RegistryContext) GetAccountName() string {
	return c.AccountName
}
