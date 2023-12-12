package domain

import (
	"context"

	"github.com/kloudlite/api/apps/build-agent/internal/env"
	"github.com/kloudlite/api/pkg/logging"
	"go.uber.org/fx"
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
