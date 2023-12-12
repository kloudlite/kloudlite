package framework

import (
	app "github.com/kloudlite/api/apps/build-agent/internal/app"
	"github.com/kloudlite/api/apps/build-agent/internal/env"
	"github.com/kloudlite/api/pkg/redpanda"
	"go.uber.org/fx"
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
