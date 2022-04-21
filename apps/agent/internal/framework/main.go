package framework

import (
	"go.uber.org/fx"
	"kloudlite.io/apps/agent/internal/app"
	"kloudlite.io/pkg/config"
	"kloudlite.io/pkg/logger"
	"kloudlite.io/pkg/messaging"
)

type Env struct {
	KafkaBrokers string `env:"KAFKA_BOOTSTRAP_SERVERS" required:"true"`
}

var Module = fx.Module("framework",
	config.EnvFx[Env](),
	fx.Provide(logger.NewLogger),
	fx.Provide(func(e *Env) messaging.KafkaClient {
		return messaging.NewKafkaClient(e.KafkaBrokers)
	}),
	app.Module,
)
