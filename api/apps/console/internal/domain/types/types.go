package types

import (
	"context"

	"github.com/kloudlite/api/pkg/repos"
)

type ConsoleContext struct {
	context.Context
	AccountName string

	UserId    repos.ID
	UserEmail string
	UserName  string
}

func (c ConsoleContext) GetUserId() repos.ID {
	return c.UserId
}

func (c ConsoleContext) GetUserEmail() string {
	return c.UserEmail
}

func (c ConsoleContext) GetUserName() string {
	return c.UserName
}

func (c ConsoleContext) GetAccountName() string {
	return c.AccountName
}

type ResourceContext struct {
	ConsoleContext
	EnvironmentName string
}

type ManagedResourceContext struct {
	ConsoleContext
	ManagedServiceName *string
	EnvironmentName    *string
}
