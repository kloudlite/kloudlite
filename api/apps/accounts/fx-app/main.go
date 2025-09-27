package fx_app

import (
	"github.com/kloudlite/api/apps/accounts/internal/app"
	"github.com/kloudlite/api/apps/accounts/internal/env"
	"github.com/kloudlite/api/pkg/errors"
	"github.com/kloudlite/api/pkg/k8s"
	"go.uber.org/fx"
)

func NewAccountsModule() fx.Option {
	accountsApp := fx.Module(
		"accounts:app",
		fx.Provide(func() (*env.Env, error) {
			if e, err := env.LoadEnv(); err != nil {
				return nil, errors.NewE(err)
			} else {
				return e, nil
			}
		}),
		// Provide a nil k8s client for local development
		// The domain should handle nil k8s client gracefully
		fx.Provide(func(e *env.Env) (k8s.Client, error) {
			return nil, nil
		}),

		app.Module,
	)
	return accountsApp
}
