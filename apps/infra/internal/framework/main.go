package framework

import (
	"context"

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
	fx.Provide(func() (*Env, error) {
		var envC Env
		err := config.LoadConfigFromEnv(&envC)
		return &envC, err
	}),
	fx.Provide(func(env *Env) logger.Logger {
		return logger.NewLogger(env.isProd)
	}),
	fx.Provide(func(e *Env) (messaging.Producer[messaging.Json], error) {
		return messaging.NewKafkaProducer[messaging.Json](e.KafkaBrokers)
	}),
	fx.Provide(func(env *Env, d domain.Domain) (messaging.Consumer[domain.SetupClusterAction], error) {
		return messaging.NewKafkaConsumer[domain.SetupClusterAction](
			[]string{env.KafkaInfraActionTopic},
			env.KafkaBrokers,
			env.KafkaGroupId,
			func(topic string, action domain.SetupClusterAction) error {
				d.CreateCluster(action)
				return nil
			},
		)
	}),
	application.Module,

	fx.Invoke(
		func(lf fx.Lifecycle, msgConsumer messaging.Consumer[domain.SetupClusterAction]) error {
			lf.Append(fx.Hook{
				OnStart: func(ctx context.Context) error {
					return msgConsumer.Subscribe()
				},
				OnStop: func(ctx context.Context) error {
					return msgConsumer.Unsubscribe()
				},
			})
			return nil
		},
	),
)
