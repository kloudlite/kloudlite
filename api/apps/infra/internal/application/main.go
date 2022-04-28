package application

import (
	"context"
	"go.uber.org/fx"
	"google.golang.org/grpc"
	"kloudlite.io/apps/infra/internal/domain"
	"kloudlite.io/grpc-interfaces/kloudlite.io/rpc/infra"
	"kloudlite.io/pkg/config"
	"kloudlite.io/pkg/logger"
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

func (i *InfraEnv) GetSubscriptionTopics() []string {
	return []string{i.KafkaInfraTopic}
}
func (i *InfraEnv) GetConsumerGroupId() string {
	return i.KafkaGroupId
}

func fxJobResponder(p messaging.Producer[any], env *InfraEnv) domain.InfraJobResponder {
	return NewInfraResponder(p, env.KafkaInfraResponseTopic)
}

var Module = fx.Module("application",
	config.EnvFx[InfraEnv](),
	fx.Provide(fxInfraClient),

	// Common Producer
	messaging.NewFxKafkaProducer[messaging.Message](),

	fx.Provide(fxJobResponder),

	messaging.NewFxKafkaConsumer[*InfraEnv](),
	fx.Invoke(func(env *InfraEnv, logger logger.Logger, d domain.Domain, consumer messaging.Consumer[*InfraEnv]) {
		consumer.On(env.KafkaInfraTopic, func(context context.Context, action messaging.Message) error {
			var _d struct {
				Type string
			}
			action.Unmarshal(&_d)
			switch _d.Type {
			case "create-cluster":
				var m struct {
					Type    string
					Payload domain.SetupClusterAction
				}
				action.Unmarshal(&m)
				return d.CreateCluster(context, m.Payload)
				break
			case "update-cluster":
				var m struct {
					Type    string
					Payload domain.UpdateClusterAction
				}
				action.Unmarshal(&m)
				return d.UpdateCluster(context, m.Payload)
				break
			case "delete-cluster":
				var m struct {
					Type    string
					Payload domain.DeleteClusterAction
				}
				action.Unmarshal(&m)
				return d.DeleteCluster(context, m.Payload)
				break
			case "add-peer":
				var m struct {
					Type    string
					Payload domain.AddPeerAction
				}
				action.Unmarshal(&m)
				return d.AddPeerToCluster(context, m.Payload)
				break
			case "delete-peer":
				var m struct {
					Type    string
					Payload domain.DeletePeerAction
				}
				action.Unmarshal(&m)
				return d.DeletePeerFromCluster(context, m.Payload)
				break
			}
			return nil
		})
	}),

	// Grpc Server
	fx.Provide(fxInfraGrpcServer),
	fx.Invoke(func(server *grpc.Server, infraServer infra.InfraServer) {
		infra.RegisterInfraServer(server, infraServer)
	}),

	domain.Module,


)
