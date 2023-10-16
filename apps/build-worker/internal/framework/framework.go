package framework

import (
	"go.uber.org/fx"
	app "kloudlite.io/apps/build-worker/internal/app"
	"kloudlite.io/apps/build-worker/internal/env"
	"kloudlite.io/pkg/redpanda"
)

type fm struct {
	ev *env.Env
}

func (fm *fm) GetBrokers() string {
	return fm.ev.KafkaBrokers
}

func (fm *fm) GetKafkaSASLAuth() *redpanda.KafkaSASLAuth {
	return nil
}

var Module = fx.Module("framework",
	fx.Provide(func(ev *env.Env) *fm {
		return &fm{ev}
	}),

	redpanda.NewClientFx[*fm](),

	app.Module,
)
