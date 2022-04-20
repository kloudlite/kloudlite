package app

import (
	"context"
	_ "fmt"
	"google.golang.org/grpc"
	"kloudlite.io/grpc-interfaces/kloudlite.io/rpc/console"
	"net/http"
	_ "net/http"

	"kloudlite.io/common"
	httpServer "kloudlite.io/pkg/http-server"

	"kloudlite.io/pkg/cache"
	"kloudlite.io/pkg/config"
	_ "kloudlite.io/pkg/logger"

	_ "github.com/99designs/gqlgen/graphql/handler"
	_ "github.com/99designs/gqlgen/graphql/playground"

	"go.uber.org/fx"
	"kloudlite.io/apps/console/internal/app/graph"
	"kloudlite.io/apps/console/internal/app/graph/generated"
	"kloudlite.io/apps/console/internal/domain"
	"kloudlite.io/apps/console/internal/domain/entities"
	"kloudlite.io/pkg/messaging"
	"kloudlite.io/pkg/repos"
)

type Env struct {
	KafkaInfraTopic         string `env:"KAFKA_INFRA_TOPIC"`
	KafkaInfraResponseTopic string `env:"KAFKA_INFRA_RESP_TOPIC"`
	KafkaConsumerGroupId    string `env:"KAFKA_GROUP_ID"`
	CookieDomain            string `env:"COOKIE_DOMAIN"`
}

var Module = fx.Module(
	"app",
	config.EnvFx[Env](),
	repos.NewFxMongoRepo[*entities.Cluster]("clusters", "clus", entities.ClusterIndexes),
	repos.NewFxMongoRepo[*entities.Device]("devices", "dev", entities.DeviceIndexes),
	repos.NewFxMongoRepo[*entities.Project]("project", "proj", entities.ProjectIndexes),
	repos.NewFxMongoRepo[*entities.Config]("config", "cfg", entities.ConfigIndexes),
	repos.NewFxMongoRepo[*entities.Secret]("secret", "sec", entities.SecretIndexes),
	repos.NewFxMongoRepo[*entities.Router]("router", "route", entities.RouterIndexes),
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
