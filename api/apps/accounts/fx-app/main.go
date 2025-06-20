package fx_app

import (
	"github.com/kloudlite/api/apps/accounts/internal/app"
	"github.com/kloudlite/api/apps/accounts/internal/env"
	"github.com/kloudlite/api/pkg/errors"
	"go.uber.org/fx"
)

func NewAccountsModule() fx.Option {
	accountsApp := fx.Module(
		"accounts:app",
		fx.Provide(func() (*env.AccountsEnv, error) {
			if e, err := env.LoadEnv(); err != nil {
				return nil, errors.NewE(err)
			} else {
				return e, nil
			}
		}),
		app.Module,
	)
	return accountsApp
}
