package fx_app

import (
	"github.com/kloudlite/api/apps/comms/internal/app"
	"github.com/kloudlite/api/apps/comms/internal/env"
	"github.com/kloudlite/api/pkg/errors"
	"go.uber.org/fx"
)

func NewCommsModule() fx.Option {
	commsApp := fx.Module(
		"comms:app",
		fx.Provide(func() (*env.CommsEnv, error) {
			if e, err := env.LoadEnv(); err != nil {
				return nil, errors.NewE(err)
			} else {
				return e, nil
			}
		}),
		app.Module,
	)
	return commsApp
}
