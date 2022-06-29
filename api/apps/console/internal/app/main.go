package app

import (
	"fmt"
	"github.com/gofiber/fiber/v2"
	fWebsocket "github.com/gofiber/websocket/v2"
	"google.golang.org/grpc"
	"kloudlite.io/common"
	"kloudlite.io/grpc-interfaces/kloudlite.io/rpc/auth"
	"kloudlite.io/grpc-interfaces/kloudlite.io/rpc/ci"
	"kloudlite.io/grpc-interfaces/kloudlite.io/rpc/console"
	"kloudlite.io/grpc-interfaces/kloudlite.io/rpc/finance"
	"kloudlite.io/grpc-interfaces/kloudlite.io/rpc/iam"
	"kloudlite.io/pkg/cache"
	"kloudlite.io/pkg/config"
	httpServer "kloudlite.io/pkg/http-server"
	loki_server "kloudlite.io/pkg/loki-server"
	"kloudlite.io/pkg/redpanda"

	"go.uber.org/fx"
	"kloudlite.io/apps/console/internal/app/graph"
	"kloudlite.io/apps/console/internal/app/graph/generated"
	"kloudlite.io/apps/console/internal/domain"
	"kloudlite.io/apps/console/internal/domain/entities"
	"kloudlite.io/pkg/repos"
)

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
type FinanceClientConnection *grpc.ClientConn

type AuthCacheClient cache.Client
type CacheClient cache.Client

var Module = fx.Module(
	"app",

	// Configs
	config.EnvFx[Env](),

	// Repos
	repos.NewFxMongoRepo[*entities.Cluster]("clusters", "clus", entities.ClusterIndexes),
	repos.NewFxMongoRepo[*entities.ClusterAccount]("cluster_accounts", "clacc", entities.ClusterAccountIndexes),
	repos.NewFxMongoRepo[*entities.Device]("devices", "dev", entities.DeviceIndexes),
	repos.NewFxMongoRepo[*entities.Project]("project", "proj", entities.ProjectIndexes),
	repos.NewFxMongoRepo[*entities.Config]("config", "cfg", entities.ConfigIndexes),
	repos.NewFxMongoRepo[*entities.Secret]("secret", "sec", entities.SecretIndexes),
	repos.NewFxMongoRepo[*entities.Router]("router", "route", entities.RouterIndexes),
	repos.NewFxMongoRepo[*entities.ManagedService]("managedservice", "mgsvc", entities.ManagedServiceIndexes),
	repos.NewFxMongoRepo[*entities.App]("app", "app", entities.AppIndexes),
	repos.NewFxMongoRepo[*entities.ManagedResource]("managedresouce", "mgres", entities.ManagedResourceIndexes),

	// Grpc Clients

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
	redpanda.NewConsumerFx[*WorkloadConsumerEnv](func(m *redpanda.Message) error {
		fmt.Println(m.Payload)
		fmt.Println(m.Action)

		//var msg struct {
		//	Status     bool   `json:"status"`
		//	Key        string `json:"key"`
		//	Conditions []struct {
		//		Type   string `json:"type"`
		//		Status string `json:"status"`
		//		Reason string `json:"reason"`
		//	} `json:"conditions"`
		//}
		//err := message.Unmarshal(&msg)
		//if err != nil {
		//	fmt.Println("Unable to parse messages!!!", err)
		//	return err
		//}
		//split := strings.Split(msg.Key, "/")
		//namespace := split[0]
		//resourceType := split[1]
		//resourceName := split[2]
		//var s domain.ResourceStatus
		//if msg.Status {
		//	s = domain.ResourceStatusLive
		//} else {
		//	if len(msg.Conditions) > 0 {
		//		s = domain.ResourceStatusInProgress
		//	} else if msg.Status {
		//		s = domain.ResourceStatusError
		//	}
		//}
		//_, err = d.UpdateResourceStatus(context, resourceType, namespace, resourceName, s)
		//return err

		return nil
	}),
	//fx.Invoke(func(env *WorkloadConsumerEnv, consumer messaging.Consumer[*WorkloadConsumerEnv], d domain.Domain) {
	//	fmt.Println(env.ResponseTopic, "env.ResponseTopic")
	//	consumer.On(env.ResponseTopic, func(context context.Context, message messaging.Message) error {
	//		var msg struct {
	//			Status     bool   `json:"status"`
	//			Key        string `json:"key"`
	//			Conditions []struct {
	//				Type   string `json:"type"`
	//				Status string `json:"status"`
	//				Reason string `json:"reason"`
	//			} `json:"conditions"`
	//		}
	//		err := message.Unmarshal(&msg)
	//		if err != nil {
	//			fmt.Println("Unable to parse messages!!!", err)
	//			return err
	//		}
	//		split := strings.Split(msg.Key, "/")
	//		namespace := split[0]
	//		resourceType := split[1]
	//		resourceName := split[2]
	//		var s domain.ResourceStatus
	//		if msg.Status {
	//			s = domain.ResourceStatusLive
	//		} else {
	//			if len(msg.Conditions) > 0 {
	//				s = domain.ResourceStatusInProgress
	//			} else if msg.Status {
	//				s = domain.ResourceStatusError
	//			}
	//		}
	//		_, err = d.UpdateResourceStatus(context, resourceType, namespace, resourceName, s)
	//		return err
	//	})
	//}),

	domain.Module,

	// Log Service
	fx.Invoke(func(logServer loki_server.LogServer, client loki_server.LokiClient, env *Env, cacheClient AuthCacheClient) {
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
		cacheClient CacheClient,
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
