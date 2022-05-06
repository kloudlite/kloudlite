package app

import (
	"context"
	"fmt"
	"github.com/gofiber/fiber/v2"
	fWebsocket "github.com/gofiber/websocket/v2"
	"google.golang.org/grpc"
	"kloudlite.io/common"
	"kloudlite.io/grpc-interfaces/kloudlite.io/rpc/auth"
	"kloudlite.io/grpc-interfaces/kloudlite.io/rpc/ci"
	"kloudlite.io/grpc-interfaces/kloudlite.io/rpc/console"
	"kloudlite.io/grpc-interfaces/kloudlite.io/rpc/iam"
	"kloudlite.io/grpc-interfaces/kloudlite.io/rpc/infra"
	httpServer "kloudlite.io/pkg/http-server"
	loki_server "kloudlite.io/pkg/loki-server"
	"strings"

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

	// Configs
	config.EnvFx[Env](),

	// Repos
	repos.NewFxMongoRepo[*entities.Cluster]("clusters", "clus", entities.ClusterIndexes),
	repos.NewFxMongoRepo[*entities.Device]("devices", "dev", entities.DeviceIndexes),
	repos.NewFxMongoRepo[*entities.Project]("project", "proj", entities.ProjectIndexes),
	repos.NewFxMongoRepo[*entities.Config]("config", "cfg", entities.ConfigIndexes),
	repos.NewFxMongoRepo[*entities.Secret]("secret", "sec", entities.SecretIndexes),
	repos.NewFxMongoRepo[*entities.Router]("router", "route", entities.RouterIndexes),
	repos.NewFxMongoRepo[*entities.ManagedService]("managedservice", "mgsvc", entities.ManagedServiceIndexes),
	repos.NewFxMongoRepo[*entities.App]("app", "app", entities.AppIndexes),
	repos.NewFxMongoRepo[*entities.ManagedResource]("managedresouce", "mgres", entities.ManagedResourceIndexes),

	// Grpc Clients
	fx.Provide(func(conn InfraClientConnection) infra.InfraClient {
		return infra.NewInfraClient((*grpc.ClientConn)(conn))
	}),

	fx.Provide(func(conn CIClientConnection) ci.CIClient {
		return ci.NewCIClient((*grpc.ClientConn)(conn))
	}),

	fx.Provide(func(conn IAMClientConnection) iam.IAMClient {
		return iam.NewIAMClient((*grpc.ClientConn)(conn))
	}),

	fx.Provide(func(conn AuthClientConnection) auth.AuthClient {
		return auth.NewAuthClient((*grpc.ClientConn)(conn))
	}),

	// Grpc Server
	fx.Provide(fxConsoleGrpcServer),
	fx.Invoke(func(server *grpc.Server, consoleServer console.ConsoleServer) {
		console.RegisterConsoleServer(server, consoleServer)
	}),

	// Common Producer
	messaging.NewFxKafkaProducer[messaging.Json](),

	// Infra Message Producer
	fx.Provide(fxInfraMessenger),

	// Infra Message Consumer
	config.EnvFx[InfraConsumerEnv](),
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

	// Workload Message Producer
	fx.Provide(fxWorkloadMessenger),
	// Workload Message Consumer
	config.EnvFx[WorkloadConsumerEnv](),
	messaging.NewFxKafkaConsumer[*WorkloadConsumerEnv](),
	fx.Invoke(func(env *WorkloadConsumerEnv, consumer messaging.Consumer[*WorkloadConsumerEnv], d domain.Domain) {
		fmt.Println(env.ResponseTopic, "env.ResponseTopic")
		consumer.On(env.ResponseTopic, func(context context.Context, message messaging.Message) error {
			var msg struct {
				Status     bool   `json:"status"`
				Key        string `json:"key"`
				Conditions []struct {
					Type   string `json:"type"`
					Status string `json:"status"`
					Reason string `json:"reason"`
				} `json:"conditions"`
			}
			err := message.Unmarshal(&msg)
			if err != nil {
				fmt.Println("Unable to parse messages!!!", err)
				return err
			}
			split := strings.Split(msg.Key, "/")
			namespace := split[0]
			resourceType := split[1]
			resourceName := split[2]
			var s domain.ResourceStatus
			if msg.Status {
				s = domain.ResourceStatusLive
			} else {
				if len(msg.Conditions) > 0 {
					s = domain.ResourceStatusInProgress
				} else if msg.Status {
					s = domain.ResourceStatusError
				}
			}
			_, err = d.UpdateResourceStatus(context, resourceType, namespace, resourceName, s)
			return err
		})
	}),

	domain.Module,

	// Log Service
	fx.Invoke(func(logServer loki_server.LogServer, client loki_server.LokiClient, env *Env, cacheClient cache.Client) {
		var a *fiber.App
		a = logServer
		a.Use(httpServer.NewSessionMiddleware[*common.AuthSession](
			cacheClient,
			"hotspot-session",
			env.CookieDomain,
			"hotspot:auth:sessions",
		))
		a.Get("/", fWebsocket.New(func(conn *fWebsocket.Conn) {
			// Crosscheck session
			client.Tail([]loki_server.StreamSelector{
				{
					Key:       "namespace",
					Operation: "=",
					Value:     "hotspot",
				},
			}, nil, nil, nil, nil, conn)
		}))
	}),

	// GraphQL Service
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
