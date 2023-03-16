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
	"kloudlite.io/pkg/agent"
	"kloudlite.io/pkg/cache"
	httpServer "kloudlite.io/pkg/http-server"
	"kloudlite.io/pkg/redpanda"
	"kloudlite.io/pkg/repos"
)

type AuthCacheClient cache.Client
type FinanceClientConnection *grpc.ClientConn

type consumerOpts struct {
	*env.Env
}

func (c *consumerOpts) GetSubscriptionTopics() []string {
	return []string{c.Env.KafkaTopicInfraUpdates}
}

func (c *consumerOpts) GetConsumerGroupId() string {
	return c.Env.KafkaConsumerGroupId
}

var Module = fx.Module(
	"app",
	repos.NewFxMongoRepo[*entities.CloudProvider]("cloud_providers", "cprovider", entities.CloudProviderIndices),
	repos.NewFxMongoRepo[*entities.Edge]("edges", "edge", entities.EdgeIndices),
	repos.NewFxMongoRepo[*entities.Cluster]("clusters", "clus", entities.ClusterIndices),
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

	fx.Provide(func(p redpanda.Producer) agent.Sender {
		return agent.NewSender(p)
	}),

	domain.Module,

	fx.Provide(func(ev *env.Env) *consumerOpts {
		return &consumerOpts{Env: ev}
	}),

	redpanda.NewConsumerFx[*consumerOpts](),

	fx.Invoke(processStatusUpdates),

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
					env.AuthRedisPrefix+":"+constants.CacheSessionPrefix,
				),
			)
		},
	),
)
