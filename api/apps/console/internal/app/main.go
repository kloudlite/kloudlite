package app

import (
	"context"
	"fmt"
	"github.com/gofiber/fiber/v2"
	"google.golang.org/grpc"
	op_crds "kloudlite.io/apps/console/internal/domain/op-crds"
	"kloudlite.io/common"
	"kloudlite.io/grpc-interfaces/kloudlite.io/rpc/auth"
	"kloudlite.io/grpc-interfaces/kloudlite.io/rpc/ci"
	"kloudlite.io/grpc-interfaces/kloudlite.io/rpc/console"
	"kloudlite.io/grpc-interfaces/kloudlite.io/rpc/iam"
	"kloudlite.io/grpc-interfaces/kloudlite.io/rpc/infra"
	"kloudlite.io/pkg/errors"
	httpServer "kloudlite.io/pkg/http-server"

	"kloudlite.io/pkg/cache"
	"kloudlite.io/pkg/config"

	"go.uber.org/fx"
	"kloudlite.io/apps/console/internal/app/graph"
	"kloudlite.io/apps/console/internal/app/graph/generated"
	"kloudlite.io/apps/console/internal/domain"
	"kloudlite.io/apps/console/internal/domain/entities"
	"kloudlite.io/pkg/messaging"
	"kloudlite.io/pkg/repos"
)

type InfraConsumerEnv struct {
	Topic         string `env:"KAFKA_INFRA_TOPIC"`
	ResponseTopic string `env:"KAFKA_INFRA_RESP_TOPIC"`
}

func (i *InfraConsumerEnv) GetSubscriptionTopics() []string {
	return []string{
		i.ResponseTopic,
	}
}

func (i *InfraConsumerEnv) GetConsumerGroupId() string {
	return "console-infra-consumer"
}

type WorkloadConsumerEnv struct {
	Topic         string `env:"KAFKA_WORKLOAD_TOPIC"`
	ResponseTopic string `env:"KAFKA_WORKLOAD_RESP_TOPIC"`
}

func (i *WorkloadConsumerEnv) GetSubscriptionTopics() []string {
	return []string{
		i.ResponseTopic,
	}
}

func (i *WorkloadConsumerEnv) GetConsumerGroupId() string {
	return "console-workload-consumer-2"
}

type Env struct {
	KafkaConsumerGroupId string `env:"KAFKA_GROUP_ID"`
	CookieDomain         string `env:"COOKIE_DOMAIN"`
}

type InfraClientConnection *grpc.ClientConn
type IAMClientConnection *grpc.ClientConn
type AuthClientConnection *grpc.ClientConn
type CIClientConnection *grpc.ClientConn

var Module = fx.Module(
	"app",

	config.EnvFx[Env](),
	config.EnvFx[InfraConsumerEnv](),
	config.EnvFx[WorkloadConsumerEnv](),

	repos.NewFxMongoRepo[*entities.Cluster]("clusters", "clus", entities.ClusterIndexes),
	repos.NewFxMongoRepo[*entities.Device]("devices", "dev", entities.DeviceIndexes),
	repos.NewFxMongoRepo[*entities.Project]("project", "proj", entities.ProjectIndexes),
	repos.NewFxMongoRepo[*entities.Config]("config", "cfg", entities.ConfigIndexes),
	repos.NewFxMongoRepo[*entities.Secret]("secret", "sec", entities.SecretIndexes),
	repos.NewFxMongoRepo[*entities.Router]("router", "route", entities.RouterIndexes),
	repos.NewFxMongoRepo[*entities.ManagedService]("managedservice", "mgsvc", entities.ManagedServiceIndexes),
	repos.NewFxMongoRepo[*entities.App]("app", "app", entities.AppIndexes),
	repos.NewFxMongoRepo[*entities.ManagedResource]("managedresouce", "mgres", entities.ManagedResourceIndexes),

	fx.Provide(func(conn InfraClientConnection) infra.InfraClient {
		return infra.NewInfraClient((*grpc.ClientConn)(conn))
	}),

	fx.Provide(func(conn CIClientConnection) ci.CIClient {
		return ci.NewCIClient((*grpc.ClientConn)(conn))
	}),

	fx.Provide(func(conn CIClientConnection) iam.IAMClient {
		return iam.NewIAMClient((*grpc.ClientConn)(conn))
	}),

	fx.Provide(func(conn CIClientConnection) auth.AuthClient {
		return auth.NewAuthClient((*grpc.ClientConn)(conn))
	}),

	fx.Provide(fxConsoleGrpcServer),
	fx.Invoke(func(server *grpc.Server, consoleServer console.ConsoleServer) {
		console.RegisterConsoleServer(server, consoleServer)
	}),

	// Common Producer
	messaging.NewFxKafkaProducer[messaging.Json](),

	fx.Provide(fxInfraMessenger),
	messaging.NewFxKafkaConsumer[*InfraConsumerEnv](),
	fx.Invoke(func(env *InfraConsumerEnv, consumer messaging.Consumer[*InfraConsumerEnv], domain domain.Domain) {
		consumer.On(env.ResponseTopic, func(context context.Context, message messaging.Message) error {
			var d map[string]any
			err := message.Unmarshal(&d)
			if err != nil {
				return err
			}
			switch d["type"].(string) {
			case "create-cluster":
				var m struct {
					Type    string
					Payload entities.SetupClusterResponse
				}
				err := message.Unmarshal(&m)
				if err != nil {
					return err
				}
				return domain.OnSetupCluster(context, m.Payload)
			case "delete-cluster":
				var m struct {
					Type    string
					Payload entities.DeleteClusterResponse
				}
				err := message.Unmarshal(&m)
				if err != nil {
					return err
				}
				return nil
			case "update-cluster":
				var m struct {
					Type    string
					Payload entities.UpdateClusterResponse
				}
				err := message.Unmarshal(&m)
				if err != nil {
					return err
				}
				return domain.OnUpdateCluster(context, m.Payload)
			case "add-peer":
				var m struct {
					Type    string
					Payload entities.AddPeerResponse
				}
				err := message.Unmarshal(&m)
				if err != nil {
					return err
				}
				return domain.OnAddPeer(context, m.Payload)
			case "delete-peer":
				var m struct {
					Type    string
					Payload entities.DeletePeerResponse
				}
				err := message.Unmarshal(&m)
				if err != nil {
					return err
				}
				return nil
			}
			return nil
		})
	}),

	fx.Provide(fxWorkloadMessenger),
	messaging.NewFxKafkaConsumer[*WorkloadConsumerEnv](),
	fx.Invoke(func(env *WorkloadConsumerEnv, consumer messaging.Consumer[*WorkloadConsumerEnv], domain domain.Domain) {
		fmt.Println(env.ResponseTopic, "env.ResponseTopic")
		consumer.On(env.ResponseTopic, func(context context.Context, message messaging.Message) error {
			var d map[string]any
			err := message.Unmarshal(&d)
			fmt.Println(string(message), "HERE")
			if err != nil {
				return err
			}
			if d["type"] == nil || d["payload"] == nil {
				return errors.New("invalid message")
			}

			switch d["type"].(string) {
			case "project-update":
				var m struct {
					Type    string
					Payload *op_crds.Project
				}
				err := message.Unmarshal(&m)
				if err != nil {
					return err
				}
				return domain.OnUpdateProject(context, m.Payload)
			case "app-update":
				var m struct {
					Type    string
					Payload *op_crds.App
				}
				err := message.Unmarshal(&m)
				if err != nil {
					return err
				}
				return domain.OnUpdateApp(context, m.Payload)
			case "config-update":
				var m struct {
					Type    string
					Payload repos.ID
				}
				err := message.Unmarshal(&m)
				if err != nil {
					return err
				}
				return domain.OnUpdateConfig(context, m.Payload)
			case "secret-update":
				var m struct {
					Type    string
					Payload repos.ID
				}
				err := message.Unmarshal(&m)
				if err != nil {
					return err
				}
				return domain.OnUpdateSecret(context, m.Payload)
			case "router-update":
				var m struct {
					Type    string
					Payload *op_crds.Router
				}
				err := message.Unmarshal(&m)
				if err != nil {
					return err
				}
				return domain.OnUpdateRouter(context, m.Payload)
			case "managed-svc-update":
				var m struct {
					Type    string
					Payload *op_crds.ManagedService
				}
				err := message.Unmarshal(&m)
				if err != nil {
					return err
				}
				return domain.OnUpdateManagedSvc(context, m.Payload)
			case "managed-res-update":
				var m struct {
					Type    string
					Payload *op_crds.ManagedResource
				}
				err := message.Unmarshal(&m)
				if err != nil {
					return err
				}
				return domain.OnUpdateManagedRes(context, m.Payload)
			}
			return nil
		})
	}),

	domain.Module,

	fx.Invoke(func(
		server *fiber.App,
		d domain.Domain,
		cacheClient cache.Client,
		env *Env,
	) {
		schema := generated.NewExecutableSchema(
			generated.Config{Resolvers: &graph.Resolver{Domain: d}},
		)
		httpServer.SetupGQLServer(
			server,
			schema,
			httpServer.NewSessionMiddleware[*common.AuthSession](
				cacheClient,
				"hotspot-session",
				env.CookieDomain,
				"hotspot:auth:sessions",
			),
		)
	}),
)
