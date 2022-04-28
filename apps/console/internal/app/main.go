package app

import (
	"context"
	"fmt"
	"google.golang.org/grpc"
	"kloudlite.io/grpc-interfaces/kloudlite.io/rpc/ci"
	"kloudlite.io/grpc-interfaces/kloudlite.io/rpc/infra"
	"net/http"

	op_crds "kloudlite.io/apps/console/internal/domain/op-crds"
	"kloudlite.io/grpc-interfaces/kloudlite.io/rpc/console"
	"kloudlite.io/pkg/logger"

	"kloudlite.io/common"
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

type Env struct {
	KafkaInfraTopic            string `env:"KAFKA_INFRA_TOPIC"`
	KafkaInfraResponseTopic    string `env:"KAFKA_INFRA_RESP_TOPIC"`
	KafkaWorkloadTopic         string `env:"KAFKA_WORKLOAD_TOPIC"`
	KafkaWorkloadResponseTopic string `env:"KAFKA_WORKLOAD_RESP_TOPIC"`
	KafkaConsumerGroupId       string `env:"KAFKA_GROUP_ID"`
	CookieDomain               string `env:"COOKIE_DOMAIN"`
}

type InfraEventConsumer messaging.Consumer
type ClusterEventConsumer messaging.Consumer
type InfraClientConnection *grpc.ClientConn
type AuthClientConnection *grpc.ClientConn
type CIClientConnection *grpc.ClientConn

var Module = fx.Module(
	"app",
	config.EnvFx[Env](),
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

	fx.Module("producer",
		fx.Provide(func(messagingCli messaging.KafkaClient) (messaging.Producer[messaging.Json], error) {
			return messaging.NewKafkaProducer[messaging.Json](messagingCli)
		}),
		fx.Provide(func(env *Env, p messaging.Producer[messaging.Json]) domain.InfraMessenger {
			return &infraMessengerImpl{
				env:      env,
				producer: p,
			}
		}),
		fx.Provide(func(env *Env, p messaging.Producer[messaging.Json]) domain.WorkloadMessenger {
			return &workloadMessengerImpl{
				env:      env,
				producer: p,
			}
		}),
		fx.Invoke(func(producer messaging.Producer[messaging.Json], lifecycle fx.Lifecycle) {
			lifecycle.Append(fx.Hook{
				OnStart: func(ctx context.Context) error {
					return producer.Connect(ctx)
				},
				OnStop: func(ctx context.Context) error {
					producer.Close(ctx)
					return nil
				},
			})
		}),
	),

	fx.Module("infra-event-consumer",
		fx.Provide(func(domain domain.Domain, env *Env, kafkaCli messaging.KafkaClient, logger logger.Logger) (ClusterEventConsumer, error) {
			return messaging.NewKafkaConsumer(
				kafkaCli,
				[]string{env.KafkaInfraResponseTopic},
				env.KafkaConsumerGroupId,
				logger, func(context context.Context, topic string, message messaging.Message) error {
					fmt.Println(string(message))
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
				},
			)
		}),
		fx.Invoke(func(consumer ClusterEventConsumer, lifecycle fx.Lifecycle) {
			lifecycle.Append(fx.Hook{
				OnStart: func(ctx context.Context) error {
					return consumer.Subscribe(ctx)
				},
				OnStop: func(ctx context.Context) error {
					consumer.Unsubscribe(ctx)
					return nil
				},
			})
		}),
	),

	fx.Module("cluster-event-consumer",
		fx.Provide(func(domain domain.Domain, env *Env, kafkaCli messaging.KafkaClient, logger logger.Logger) (InfraEventConsumer, error) {
			return messaging.NewKafkaConsumer(
				kafkaCli,
				[]string{env.KafkaInfraTopic},
				env.KafkaConsumerGroupId,
				logger, func(context context.Context, topic string, message messaging.Message) error {
					var d map[string]any
					err := message.Unmarshal(&d)
					if err != nil {
						return err
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
				},
			)
		}),
	),

	domain.Module,

	fx.Provide(fxConsoleGrpcServer),
	fx.Invoke(func(server *grpc.Server, consoleServer console.ConsoleServer) {
		console.RegisterConsoleServer(server, consoleServer)
	}),

	fx.Invoke(func(
		server *http.ServeMux,
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
			cache.NewSessionRepo[*common.AuthSession](
				cacheClient,
				"hotspot-session",
				env.CookieDomain,
				"hotspot:auth:sessions",
			),
		)
	}),
)
