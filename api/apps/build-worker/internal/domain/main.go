package domain

import (
	"context"

	"go.uber.org/fx"
	"kloudlite.io/apps/build-worker/internal/env"
	"kloudlite.io/pkg/logging"
)

type Impl struct {
	envs   *env.Env
	logger logging.Logger
}

func (d *Impl) ProcessRegistryEvents(ctx context.Context) error {
	panic("not implemented")
}

var Module = fx.Module(
	"domain",
	fx.Provide(
		func(e *env.Env,
			logger logging.Logger,
		) (Domain, error) {
			return &Impl{
				envs:   e,
				logger: logger,
			}, nil
		}),
)
