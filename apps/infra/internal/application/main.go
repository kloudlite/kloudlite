package application

import (
	"context"
	"go.uber.org/fx"
	"kloudlite.io/apps/infra/internal/domain"
	"kloudlite.io/pkg/config"
	"kloudlite.io/pkg/messaging"
)

type InfraEnv struct {
	DoAPIKey              string `env:"DO_API_KEY", required:"true"`
	DataPath              string `env:"DATA_PATH", required:"true"`
	SshKeysPath           string `env:"SSH_KEYS_PATH", required:"true"`
	KafkaInfraActionTopic string `env:"KAFKA_INFRA_ACTION_TOPIC", required:"true"`
	KafkaGroupId          string `env:"KAFKA_GROUP_ID", required:"true"`
}

func fxConsumer(env *InfraEnv, mc messaging.KafkaClient, d domain.Domain) (messaging.Consumer[domain.SetupClusterAction], error) {
	consumer, err := messaging.NewKafkaConsumer[domain.SetupClusterAction](
		mc,
		[]string{env.KafkaInfraActionTopic},
		env.KafkaGroupId,
		func(topic string, action domain.SetupClusterAction) error {
			d.CreateCluster(action)
			return nil
		},
	)
	return consumer, err
}

func fxEnv() (*InfraEnv, error) {
	var envC InfraEnv
	err := config.LoadConfigFromEnv(&envC)
	return &envC, err
}

func fxProducer(mc messaging.KafkaClient) (messaging.Producer[messaging.Json], error) {
	return messaging.NewKafkaProducer[messaging.Json](mc)
}

var Module = fx.Module("application",
	fx.Provide(fxEnv),
	fx.Provide(fxInfraClient),
	fx.Provide(fxProducer),
	domain.Module,
	fx.Provide(fxConsumer),
	fx.Invoke(func(lf fx.Lifecycle, consumer messaging.Consumer[domain.SetupClusterAction]) {
		lf.Append(fx.Hook{
			OnStart: func(ctx context.Context) error {
				return consumer.Subscribe()
			},
			OnStop: func(ctx context.Context) error {
				return consumer.Unsubscribe()
			},
		})
	}),
)
