package framework

import (
	"go.uber.org/fx"
	"kloudlite.io/apps/webhooks/internal/app"
	"kloudlite.io/apps/webhooks/internal/env"
	httpServer "kloudlite.io/pkg/http-server"
	"kloudlite.io/pkg/redpanda"
)

type fromEnv struct {
	*env.Env
}

func (v fromEnv) GetBrokerHosts() string {
	return v.KafkaBrokers
}

func (v fromEnv) GetHttpPort() uint16 {
	return v.HttpPort
}

func (v fromEnv) GetHttpCors() string {
	return ""
}

var Module = fx.Module(
	"framework",
	fx.Provide(
		func(vars *env.Env) fromEnv {
			return fromEnv{Env: vars}
		},
	),

	httpServer.NewHttpServerFx[fromEnv](),
	redpanda.NewProducerFx[fromEnv](),
	app.Module,
)
