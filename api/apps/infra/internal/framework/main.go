package framework

import (
	"go.uber.org/fx"
	"kloudlite.io/pkg/config"
	"kloudlite.io/pkg/logger"
	"kloudlite.io/pkg/messaging"
)

type Env struct {
	KafkaBrokers string `env:"KAFKA_BOOTSTRAP_SERVERS", required:"true"`
	KafkaGroupId string `env:"KAFKA_GROUP_ID", required:"true"`
	isProd       bool   `env:"PROD"`
}

var Module = fx.Module("framework",
	// Load Env
	fx.Provide(func() (*Env, error) {
		var envC Env
		err := config.LoadConfigFromEnv(&envC)
		return &envC, err
	}),
	// Setup Logger
	fx.Provide(func(env *Env) logger.Logger {
		return logger.NewLogger(env.isProd)
	}),
	// Setup Consumer
	fx.Provide(func(env *Env) (messaging.Consumer, error) {
		return messaging.NewKafkaConsumer(env.KafkaBrokers, env.KafkaGroupId)
	}),
)
