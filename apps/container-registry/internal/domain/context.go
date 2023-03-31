package domain

import (
	"context"
	"kloudlite.io/pkg/repos"
)

type RegistryContext struct {
	context.Context
	userId      repos.ID
	accountName string
}

func (c *RegistryContext) GetAccountName() string {
	return c.accountName
}
