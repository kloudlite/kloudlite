package application

import (
	"context"
	"go.uber.org/fx"
	"kloudlite.io/apps/infra/internal/domain"
	"kloudlite.io/pkg/config"
	"kloudlite.io/pkg/messaging"
	// "kloudlite.io/pkg/messaging"
)

type action interface {
	domain.AddPeerAction | domain.DeleteClusterAction | domain.DeletePeerAction | domain.UpdateClusterAction | domain.SetupClusterAction
}

type Message[T action] struct {
	messageType string
	ref         T
}

type InfraEnv struct {
	DoImageId               string `env:"DO_IMAGE_ID", required:"true"`
	DoAPIKey                string `env:"DO_API_KEY", required:"true"`
	DataPath                string `env:"DATA_PATH", required:"true"`
	SshKeysPath             string `env:"SSH_KEYS_PATH", required:"true"`
	KafkaInfraTopic         string `env:"KAFKA_INFRA_TOPIC", required:"true"`
	KafkaInfraResponseTopic string `env:"KAFKA_INFRA_RESP_TOPIC", required:"true"`
	KafkaGroupId            string `env:"KAFKA_GROUP_ID", required:"true"`
}

func fxProducer(mc messaging.KafkaClient) (messaging.Producer[any], error) {
	return messaging.NewKafkaProducer[any](mc)
}

func fxJobResponder(p messaging.Producer[any], env InfraEnv) domain.InfraJobResponder {
	return NewInfraResponder(p, env.KafkaInfraResponseTopic)
}

var Module = fx.Module("application",
	fx.Provide(config.LoadEnv[InfraEnv]()),
	fx.Provide(fxInfraClient),
	fx.Provide(fxProducer),
	fx.Provide(fxJobResponder),
	domain.Module,
	fx.Invoke(func(lifecycle fx.Lifecycle, producer messaging.Producer[any]) {
		lifecycle.Append(fx.Hook{
			OnStart: func(c context.Context) error {
				return producer.Connect(c)
			},
		})
	}),

	fx.Provide(fxConsumer),

	fx.Invoke(func(lf fx.Lifecycle, consumer messaging.Consumer) {
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
