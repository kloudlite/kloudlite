package app

import (
	"context"
	"fmt"

	"github.com/99designs/gqlgen/graphql"
	"github.com/gofiber/fiber/v2"
	"go.uber.org/fx"
	"google.golang.org/grpc"
	"kloudlite.io/apps/infra/internal/app/graph"
	"kloudlite.io/apps/infra/internal/app/graph/generated"
	"kloudlite.io/apps/infra/internal/domain"
	"kloudlite.io/apps/infra/internal/domain/entities"
	"kloudlite.io/apps/infra/internal/env"
	"kloudlite.io/common"
	"kloudlite.io/constants"
	"kloudlite.io/grpc-interfaces/kloudlite.io/rpc/finance"
	"kloudlite.io/pkg/cache"
	httpServer "kloudlite.io/pkg/http-server"
	"kloudlite.io/pkg/logging"
	"kloudlite.io/pkg/redpanda"
	"kloudlite.io/pkg/repos"

	fWebsocket "github.com/gofiber/websocket/v2"
)

type AuthCacheClient cache.Client
type FinanceClientConnection *grpc.ClientConn

var Module = fx.Module(
	"app",
	repos.NewFxMongoRepo[*entities.CloudProvider]("cloud_providers", "cprovider", entities.CloudProviderIndices),
	repos.NewFxMongoRepo[*entities.Edge]("edges", "edge", entities.EdgeIndices),
	repos.NewFxMongoRepo[*entities.Cluster]("clusters", "clus", entities.ClusterIndices),
	repos.NewFxMongoRepo[*entities.BYOCCluster]("byoc_clusters", "byoc", entities.BYOCClusterIndices),
	repos.NewFxMongoRepo[*entities.MasterNode]("master_nodes", "mnode", entities.MasterNodeIndices),
	repos.NewFxMongoRepo[*entities.WorkerNode]("worker_nodes", "wnode", entities.WorkerNodeIndices),
	repos.NewFxMongoRepo[*entities.NodePool]("node_pools", "npool", entities.NodePoolIndices),
	repos.NewFxMongoRepo[*entities.Secret]("secrets", "scrt", entities.SecretIndices),

	fx.Provide(
		func(conn FinanceClientConnection) finance.FinanceClient {
			return finance.NewFinanceClient((*grpc.ClientConn)(conn))
		},
	),

	redpanda.NewProducerFx[redpanda.Client](),

	domain.Module,

	fx.Provide(func(cli redpanda.Client, ev *env.Env, logger logging.Logger) (ByocClientUpdatesConsumer, error) {
		return redpanda.NewConsumer(cli.GetBrokerHosts(), ev.KafkaConsumerGroupId, redpanda.ConsumerOpts{
			SASLAuth: cli.GetKafkaSASLAuth(),
			Logger:   logger.WithName("byoc-client-updates"),
		}, []string{ev.KafkaTopicByocClientUpdates})
	}),

	fx.Invoke(processByocClientUpdates),

	fx.Provide(func(cli redpanda.Client, ev *env.Env, logger logging.Logger) (InfraUpdatesConsumer, error) {
		return redpanda.NewConsumer(cli.GetBrokerHosts(), ev.KafkaConsumerGroupId, redpanda.ConsumerOpts{
			SASLAuth: cli.GetKafkaSASLAuth(),
			Logger:   logger.WithName("infra-updates"),
		}, []string{ev.KafkaTopicByocClientUpdates})
	}),

	fx.Invoke(processInfraUpdates),

	fx.Invoke(func(server *fiber.App, logger logging.Logger) {
		server.Use("/ws", func(ctx *fiber.Ctx) error {
			if fWebsocket.IsWebSocketUpgrade(ctx) {
				return ctx.Next()
			}
			return fiber.ErrUpgradeRequired
		})

		server.Get("/ws/status-updates", fWebsocket.New(func(conn *fWebsocket.Conn) {
			logger.Infof("new socket request received ...")
			defer conn.Close()
			conn.WriteJSON(map[string]any{"hello": "world"})
		}))
	}),

	fx.Invoke(
		func(
			server *fiber.App,
			d domain.Domain,
			cacheClient AuthCacheClient,
			env *env.Env,
		) {
			config := generated.Config{Resolvers: &graph.Resolver{Domain: d}}
			config.Directives.IsLoggedIn = func(ctx context.Context, obj interface{}, next graphql.Resolver) (res interface{}, err error) {
				sess := httpServer.GetSession[*common.AuthSession](ctx)
				if sess == nil {
					return nil, fiber.ErrUnauthorized
				}
				return next(ctx)
			}

			config.Directives.HasAccount = func(ctx context.Context, obj interface{}, next graphql.Resolver) (res interface{}, err error) {
				m := httpServer.GetHttpCookies(ctx)
				klAccount := m[env.AccountCookieName]
				if klAccount == "" {
					return nil, fmt.Errorf("no cookie named '%s' present in request", "kloudlite-cluster")
				}
				cc := domain.InfraContext{
					Context:     ctx,
					AccountName: klAccount,
				}
				return next(context.WithValue(ctx, "infra-ctx", cc))
			}

			schema := generated.NewExecutableSchema(config)
			httpServer.SetupGQLServer(
				server,
				schema,
				httpServer.NewSessionMiddleware[*common.AuthSession](
					cacheClient,
					"hotspot-session",
					env.CookieDomain,
					constants.CacheSessionPrefix,
				),
			)
		},
	),
)
