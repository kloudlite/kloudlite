package application

import (
	"context"
	"fmt"
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

func fxProducer(mc messaging.KafkaClient) (messaging.Producer[messaging.Json], error) {
	return messaging.NewKafkaProducer[messaging.Json](mc)
}

func fxJobResponder(p messaging.Producer[any], env InfraEnv) domain.InfraJobResponder {
	return NewInfraResponder(p, env.KafkaInfraResponseTopic)
}

var Module = fx.Module("application",
	config.EnvFx[InfraEnv](),
	fx.Provide(fxInfraClient),
	fx.Provide(fxProducer),
	fx.Provide(fxConsumer),
	fx.Provide(fxJobResponder),
	domain.Module,
	fx.Invoke(func(lifecycle fx.Lifecycle, producer messaging.Producer[messaging.Json]) {
		lifecycle.Append(fx.Hook{
			OnStart: func(c context.Context) error {
				fmt.Println("CONNECTED")
				return producer.Connect(c)
			},
		})
	}),

	fx.Invoke(func(lf fx.Lifecycle, consumer messaging.Consumer) {
		lf.Append(fx.Hook{
			OnStart: func(ctx context.Context) error {
				return consumer.Subscribe(ctx)
			},
			OnStop: func(ctx context.Context) error {
				return consumer.Unsubscribe(ctx)
			},
		})
	}),

	fx.Invoke(func(lifecycle fx.Lifecycle, p messaging.Producer[messaging.Json]) {
		lifecycle.Append(fx.Hook{
			OnStart: func(ctx context.Context) error {
				fmt.Println("SENT")
				//p.SendMessage("dev-hotspot-infra", "infra", messaging.Json{
				//	"type": "setup-cluster",
				//	"payload": messaging.Json{
				//		"cluster_id":  "hotspot-dev",
				//		"region":      "blr1",
				//		"provider":    "do",
				//		"nodes_count": 1,
				//	},
				//})
				return nil
			},
			OnStop: nil,
		})
	}),
)
