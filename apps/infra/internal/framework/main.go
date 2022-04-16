package framework

import (
	"go.uber.org/fx"
	"kloudlite.io/apps/infra/internal/application"
	"kloudlite.io/pkg/config"
	"kloudlite.io/pkg/logger"
	"kloudlite.io/pkg/messaging"
)

type Env struct {
	KafkaBrokers string `env:"KAFKA_BOOTSTRAP_SERVERS", required:"true"`
}

var Module = fx.Module("framework",
	fx.Provide(config.LoadEnv[*Env]()),
	fx.Provide(logger.NewLogger),
	fx.Provide(func(env *Env) messaging.KafkaClient {
		return messaging.NewKafkaClient(env.KafkaBrokers)
	}),
	application.Module,
)
