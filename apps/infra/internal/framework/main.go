package framework

import (
	"go.uber.org/fx"
	"kloudlite.io/apps/infra/internal/application"
	"kloudlite.io/apps/infra/internal/domain"
	"kloudlite.io/pkg/config"
	"kloudlite.io/pkg/logger"
	"kloudlite.io/pkg/messaging"
)

type Env struct {
	KafkaBrokers          string `env:"KAFKA_BOOTSTRAP_SERVERS", required:"true"`
	KafkaGroupId          string `env:"KAFKA_GROUP_ID", required:"true"`
	KafkaInfraActionTopic string `env:"KAFKA_INFRA_ACTION_TOPIC", required:"true"`
	isProd                bool   `env:"PROD"`
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
	// Create Producer
	fx.Provide(func(e *Env) (messaging.Producer[messaging.Json], error) {
		return messaging.NewKafkaProducer[messaging.Json](e.KafkaBrokers)
	}),
	// Setup Consumer
	fx.Provide(func(env *Env) (messaging.Consumer[domain.SetupClusterAction], error) {
		return messaging.NewKafkaConsumer[domain.SetupClusterAction]([]string{env.KafkaInfraActionTopic}, env.KafkaBrokers, env.KafkaGroupId)
	}),
	application.Module,
)
