package domain

import (
	"go.uber.org/fx"
	"golang.org/x/net/context"
	"kloudlite.io/pkg/repos"
)

type AccountsContext struct {
	context.Context
	AccountName string
	UserId      repos.ID
}

//type Domain interface {
//	GetAccount(ctx context.Context, accountName string) (*entities.Account, error)
//}

var Module = fx.Module("domain", fx.Provide(fxDomain))
