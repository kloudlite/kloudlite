package framework

import (
	"go.uber.org/fx"
	"kloudlite.io/apps/messages-distribution-worker/internal/app"
	"kloudlite.io/apps/messages-distribution-worker/internal/env"
	"kloudlite.io/pkg/kafka"
)

var Module = fx.Module(
	"framework",
	fx.Provide(func(ev *env.Env) (app.KafkaConn, error) {
		return kafka.Connect(ev.KafkaBrokers, kafka.ConnectOpts{})
	}),
	app.Module,
)
