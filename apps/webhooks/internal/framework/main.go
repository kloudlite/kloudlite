package framework

import (
	"go.uber.org/fx"
	"kloudlite.io/apps/webhooks/internal/app"
	"kloudlite.io/apps/webhooks/internal/env"
	httpServer "kloudlite.io/pkg/http-server"
	"kloudlite.io/pkg/redpanda"
)

type fm struct {
	*env.Env
}

func (f fm) GetKafkaSASLAuth() *redpanda.KafkaSASLAuth {
	return nil
	// return &redpanda.KafkaSASLAuth{
	// 	SASLMechanism: redpanda.ScramSHA256,
	// 	User:          v.KafkaUsername,
	// 	Password:      v.KafkaPassword,
	// }
}

func (f fm) GetBrokers() string {
	return f.KafkaBrokers
}

func (f fm) GetHttpPort() uint16 {
	return f.HttpPort
}

func (f fm) GetHttpCors() string {
	return ""
}

var Module = fx.Module(
	"framework",
	fx.Provide(
		func(vars *env.Env) *fm {
			return &fm{Env: vars}
		},
	),

	redpanda.NewClientFx[*fm](),
	httpServer.NewHttpServerFx[*fm](),
	app.Module,
)
