package app

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/gofiber/fiber/v2"
	fWebsocket "github.com/gofiber/websocket/v2"
	"google.golang.org/grpc"
	opcrds "kloudlite.io/apps/console/internal/domain/op-crds"
	"kloudlite.io/common"
	"kloudlite.io/grpc-interfaces/kloudlite.io/rpc/auth"
	"kloudlite.io/grpc-interfaces/kloudlite.io/rpc/ci"
	"kloudlite.io/grpc-interfaces/kloudlite.io/rpc/console"
	"kloudlite.io/grpc-interfaces/kloudlite.io/rpc/finance"
	"kloudlite.io/grpc-interfaces/kloudlite.io/rpc/iam"
	"kloudlite.io/grpc-interfaces/kloudlite.io/rpc/jseval"
	"kloudlite.io/pkg/cache"
	"kloudlite.io/pkg/config"
	httpServer "kloudlite.io/pkg/http-server"
	lokiserver "kloudlite.io/pkg/loki-server"
	"kloudlite.io/pkg/redpanda"
	"time"

	"go.uber.org/fx"
	"kloudlite.io/apps/console/internal/app/graph"
	"kloudlite.io/apps/console/internal/app/graph/generated"
	"kloudlite.io/apps/console/internal/domain"
	"kloudlite.io/apps/console/internal/domain/entities"
	"kloudlite.io/pkg/repos"
)

type WorkloadStatusConsumerEnv struct {
	Topic string `env:"KAFKA_WORKLOAD_STATUS_TOPIC"`
}

func (i *WorkloadStatusConsumerEnv) GetSubscriptionTopics() []string {
	return []string{
		i.Topic,
	}
}

func (i *WorkloadStatusConsumerEnv) GetConsumerGroupId() string {
	return "console-workload-consumer-2"
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
	AuthRedisPrefix      string `env:"REDIS_AUTH_PREFIX"`
}

type InfraClientConnection *grpc.ClientConn
type IAMClientConnection *grpc.ClientConn
type JSEvalClientConnection *grpc.ClientConn
type AuthClientConnection *grpc.ClientConn
type CIClientConnection *grpc.ClientConn
type FinanceClientConnection *grpc.ClientConn

type AuthCacheClient cache.Client
type CacheClient cache.Client

var Module = fx.Module(
	"app",

	// Configs
	config.EnvFx[Env](),

	// Repos
	repos.NewFxMongoRepo[*entities.Cluster]("clusters", "clus", entities.ClusterIndexes),
	repos.NewFxMongoRepo[*entities.EdgeRegion]("regions", "reg", entities.EdgeRegionIndexes),
	repos.NewFxMongoRepo[*entities.CloudProvider]("providers", "cp", entities.CloudProviderIndexes),
	repos.NewFxMongoRepo[*entities.Device]("devices", "dev", entities.DeviceIndexes),
	repos.NewFxMongoRepo[*entities.Project]("project", "proj", entities.ProjectIndexes),
	repos.NewFxMongoRepo[*entities.Config]("config", "cfg", entities.ConfigIndexes),
	repos.NewFxMongoRepo[*entities.Secret]("secret", "sec", entities.SecretIndexes),
	repos.NewFxMongoRepo[*entities.Router]("router", "route", entities.RouterIndexes),
	repos.NewFxMongoRepo[*entities.ManagedService]("managedservice", "mgsvc", entities.ManagedServiceIndexes),
	repos.NewFxMongoRepo[*entities.App]("app", "app", entities.AppIndexes),
	repos.NewFxMongoRepo[*entities.ManagedResource]("managedresouce", "mgres", entities.ManagedResourceIndexes),

	// Grpc Clients

	fx.Provide(func(conn JSEvalClientConnection) jseval.JSEvalClient {
		return jseval.NewJSEvalClient((*grpc.ClientConn)(conn))
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

	fx.Provide(func(conn FinanceClientConnection) finance.FinanceClient {
		return finance.NewFinanceClient((*grpc.ClientConn)(conn))
	}),


	// Grpc Server
	fx.Provide(fxConsoleGrpcServer),
	fx.Invoke(func(server *grpc.Server, consoleServer console.ConsoleServer) {
		console.RegisterConsoleServer(server, consoleServer)
	}),

	// Common Producer
	redpanda.NewProducerFx(),

	// Workload Message Producer
	fx.Provide(fxWorkloadMessenger),
	// Workload Message Consumer
	config.EnvFx[WorkloadConsumerEnv](),
	config.EnvFx[WorkloadStatusConsumerEnv](),
	redpanda.NewConsumerFx[*WorkloadStatusConsumerEnv](),
	fx.Invoke(func(domain domain.Domain, consumer redpanda.Consumer) {
		consumer.StartConsuming(func(msg []byte, timestamp time.Time, offset int64) error {
			var update opcrds.StatusUpdate
			if err := json.Unmarshal(msg, &update); err != nil {
				fmt.Println(err)
				return err
			}
			fmt.Println("processing", offset, string(msg), timestamp)
			if update.Stage == "EXISTS" {
				switch update.Metadata.GroupVersionKind.Kind {
				case "App":
					domain.OnUpdateApp(context.TODO(), &update)

				case "Lambda":
					domain.OnUpdateApp(context.TODO(), &update)

				case "Router":
					domain.OnUpdateRouter(context.TODO(), &update)

				case "Project":
					domain.OnUpdateProject(context.TODO(), &update)

				case "ManagedResource":
					domain.OnUpdateManagedRes(context.TODO(), &update)
				case "ManagedService":
					domain.OnUpdateManagedSvc(context.TODO(), &update)

				default:
					fmt.Println("Unknown Kind:", update.Metadata.GroupVersionKind.Kind)
				}
			}
			if update.Stage == "DELETED" {
				switch update.Metadata.GroupVersionKind.Kind {
				case "App":
					domain.OnDeleteApp(context.TODO(), &update)
				case "Lambda":
					domain.OnDeleteApp(context.TODO(), &update)
				case "Router":
					domain.OnDeleteRouter(context.TODO(), &update)
				case "Project":
					domain.OnDeleteProject(context.TODO(), &update)
				case "ManagedService":
					domain.OnDeleteManagedService(context.TODO(), &update)
				case "ManagedResource":
					domain.OnDeleteManagedResource(context.TODO(), &update)
				default:
					fmt.Println("Unknown Kind:", update.Metadata.GroupVersionKind.Kind)
				}
			}
			return nil
		})
	}),

	domain.Module,

	// Log Service
	fx.Invoke(func(logServer lokiserver.LogServer, client lokiserver.LokiClient, env *Env, cacheClient AuthCacheClient, d domain.Domain) {
		var a *fiber.App
		a = logServer
		a.Use(httpServer.NewSessionMiddleware[*common.AuthSession](
			cacheClient,
			"hotspot-session",
			env.CookieDomain,
			common.CacheSessionPrefix,
		))
		a.Get("/build-logs", fWebsocket.New(func(conn *fWebsocket.Conn) {
			appId := conn.Query("app_id", "app_id")
			app, err := d.GetApp(context.TODO(), repos.ID(appId))
			pipelineId, ok := app.Metadata["pipeline_id"]
			if err != nil {
				fmt.Println(err)
				conn.Close()
				return
			}
			if !ok {
				fmt.Println("no pipeline_id found")
				conn.Close()
				return
			}

			// Crosscheck session
			client.Tail([]lokiserver.StreamSelector{
				{
					Key:       "namespace",
					Operation: "=",
					Value:     app.Namespace,
				},
				{
					Key:       "app",
					Operation: "=",
					Value:     pipelineId.(string),
				},
			}, nil, nil, nil, nil, conn)
		}))
		a.Get("/app-logs", fWebsocket.New(func(conn *fWebsocket.Conn) {
			appId := conn.Query("app_id", "app_id")
			app, err := d.GetApp(context.TODO(), repos.ID(appId))
			if err != nil {
				fmt.Println(err)
				conn.Close()
				return
			}
			// Crosscheck session
			client.Tail([]lokiserver.StreamSelector{
				{
					Key:       "namespace",
					Operation: "=",
					Value:     app.Namespace,
				},
				{
					Key:       "app",
					Operation: "=",
					Value:     app.ReadableId,
				},
			}, nil, nil, nil, nil, conn)
		}))
	}),

	// GraphQL Service
	fx.Invoke(func(
		server *fiber.App,
		d domain.Domain,
		cacheClient AuthCacheClient,
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
				common.CacheSessionPrefix,
			),
		)
	}),
)
