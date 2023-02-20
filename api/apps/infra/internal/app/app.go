package app

import (
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
	"kloudlite.io/pkg/redpanda"
	"kloudlite.io/pkg/repos"
)

type AuthCacheClient cache.Client
type FinanceClientConnection *grpc.ClientConn

var Module = fx.Module(
	"app",
	repos.NewFxMongoRepo[*entities.CloudProvider]("cloud_providers", "cp", entities.CloudProviderIndices),
	repos.NewFxMongoRepo[*entities.Edge]("edges", "edge", entities.EdgeIndices),
	repos.NewFxMongoRepo[*entities.Cluster]("clusters", "clus", entities.ClusterIndices),
	repos.NewFxMongoRepo[*entities.MasterNode]("clusters", "clus", entities.MasterNodeIndices),
	repos.NewFxMongoRepo[*entities.WorkerNode]("clusters", "clus", entities.WorkerNodeIndices),
	repos.NewFxMongoRepo[*entities.NodePool]("clusters", "clus", entities.NodePoolIndices),
	repos.NewFxMongoRepo[*entities.Secret]("secrets", "scrt", entities.SecretIndices),

	fx.Provide(
		func(conn FinanceClientConnection) finance.FinanceClient {
			return finance.NewFinanceClient((*grpc.ClientConn)(conn))
		},
	),

	redpanda.NewProducerFx[redpanda.Client](),

	fx.Provide(func(producer redpanda.Producer) domain.AgentMessenger {
		return domain.NewAgentMessenger(producer)
	}),

	domain.Module,

	fx.Invoke(
		func(
			server *fiber.App,
			d domain.Domain,
			cacheClient AuthCacheClient,
			env *env.Env,
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
					constants.CacheSessionPrefix,
				),
			)
		},
	),
)
