package fx_app

import (
	"github.com/kloudlite/api/apps/auth/internal/app"
	"github.com/kloudlite/api/apps/auth/internal/env"
	"github.com/kloudlite/api/pkg/errors"
	"go.uber.org/fx"
)

func NewAuthModule() fx.Option {
	authApp := fx.Module(
		"auth:app",
		fx.Provide(func() (*env.AuthEnv, error) {
			if e, err := env.LoadEnv(); err != nil {
				return nil, errors.NewE(err)
			} else {
				return e, nil
			}
		}),
		app.Module,
	)
	return authApp
}
