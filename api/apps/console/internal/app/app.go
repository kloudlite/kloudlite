package app

import (
	"context"
	"fmt"

	"github.com/99designs/gqlgen/graphql"
	"github.com/gofiber/fiber/v2"
	"go.uber.org/fx"
	"google.golang.org/grpc"

	"kloudlite.io/apps/console/internal/app/graph"
	"kloudlite.io/apps/console/internal/app/graph/generated"
	domain "kloudlite.io/apps/console/internal/domain"
	"kloudlite.io/apps/console/internal/domain/entities"
	"kloudlite.io/apps/console/internal/env"
	"kloudlite.io/common"
	"kloudlite.io/constants"
	"kloudlite.io/grpc-interfaces/kloudlite.io/rpc/iam"
	"kloudlite.io/pkg/cache"
	httpServer "kloudlite.io/pkg/http-server"
	"kloudlite.io/pkg/logging"
	"kloudlite.io/pkg/redpanda"
	"kloudlite.io/pkg/repos"
)

type AuthCacheClient cache.Client

type IAMGrpcClient *grpc.ClientConn

var Module = fx.Module("app",
	repos.NewFxMongoRepo[*entities.Project]("projects", "prj", entities.ProjectIndexes),
	repos.NewFxMongoRepo[*entities.Workspace]("workspaces", "ws", entities.WorkspaceIndexes),
	repos.NewFxMongoRepo[*entities.App]("apps", "app", entities.AppIndexes),
	repos.NewFxMongoRepo[*entities.Config]("configs", "cfg", entities.ConfigIndexes),
	repos.NewFxMongoRepo[*entities.Secret]("secrets", "scrt", entities.SecretIndexes),
	repos.NewFxMongoRepo[*entities.MRes]("managed_resources", "mres", entities.MresIndexes),
	repos.NewFxMongoRepo[*entities.MSvc]("managed_services", "msvc", entities.MsvcIndexes),
	repos.NewFxMongoRepo[*entities.Router]("routers", "rt", entities.RouterIndexes),

	fx.Invoke(
		func(
			server *fiber.App,
			d domain.Domain,
			cacheClient AuthCacheClient,
			ev *env.Env,
		) {
			gqlConfig := generated.Config{Resolvers: &graph.Resolver{Domain: d}}
			gqlConfig.Directives.IsLoggedIn = func(ctx context.Context, obj interface{}, next graphql.Resolver) (res interface{}, err error) {
				sess := httpServer.GetSession[*common.AuthSession](ctx)
				if sess == nil {
					return nil, fiber.ErrUnauthorized
				}

				return next(ctx)
			}

			gqlConfig.Directives.HasAccountAndCluster = func(ctx context.Context, obj interface{}, next graphql.Resolver) (res interface{}, err error) {
				sess := httpServer.GetSession[*common.AuthSession](ctx)
				if sess == nil {
					return nil, fiber.ErrUnauthorized
				}

				m := httpServer.GetHttpCookies(ctx)
				klAccount := m["kloudlite-account"]
				if klAccount == "" {
					return nil, fmt.Errorf("no cookie named '%s' present in request", "kloudlite-account")
				}
				klCluster := m["kloudlite-cluster"]
				if klCluster == "" {
					return nil, fmt.Errorf("no cookie named '%s' present in request", "kloudlite-cluster")
				}

				cc := domain.NewConsoleContext(ctx, sess.UserId, klAccount, klCluster)
				return next(context.WithValue(ctx, "kloudlite-ctx", cc))
			}

			gqlConfig.Directives.HasAccount = func(ctx context.Context, obj interface{}, next graphql.Resolver) (res interface{}, err error) {
				sess := httpServer.GetSession[*common.AuthSession](ctx)
				if sess == nil {
					return nil, fiber.ErrUnauthorized
				}
				m := httpServer.GetHttpCookies(ctx)
				klAccount := m["kloudlite-account"]
				if klAccount == "" {
					return nil, fmt.Errorf("no cookie named %q present in request", "kloudlite-account")
				}

				cc := domain.NewConsoleContext(ctx, sess.UserId, klAccount, "")
				return next(context.WithValue(ctx, "kloudlite-ctx", cc))
			}

			schema := generated.NewExecutableSchema(gqlConfig)
			httpServer.SetupGQLServer(
				server,
				schema,
				httpServer.NewSessionMiddleware[*common.AuthSession](
					cacheClient,
					constants.CookieName,
					ev.CookieDomain,
					constants.CacheSessionPrefix,
				),
			)
		},
	),
	redpanda.NewProducerFx[redpanda.Client](),

	fx.Provide(func(cli redpanda.Client, ev *env.Env, logger logging.Logger) (ErrorOnApplyConsumer, error) {
		return redpanda.NewConsumer(cli.GetBrokerHosts(), ev.KafkaConsumerGroupId, redpanda.ConsumerOpts{
			SASLAuth: cli.GetKafkaSASLAuth(),
			Logger:   logger,
		}, []string{ev.KafkaErrorOnApplyTopic})
	}),
	fx.Invoke(ProcessErrorOnApply),

	fx.Provide(func(cli redpanda.Client, ev *env.Env, logger logging.Logger) (ResourceUpdateConsumer, error) {
		return redpanda.NewConsumer(cli.GetBrokerHosts(), ev.KafkaConsumerGroupId, redpanda.ConsumerOpts{
			SASLAuth: cli.GetKafkaSASLAuth(),
			Logger:   logger,
		}, []string{ev.KafkaStatusUpdatesTopic})
	}),

	fx.Invoke(ProcessResourceUpdates),

	fx.Provide(
		func(clientConn IAMGrpcClient) iam.IAMClient {
			return iam.NewIAMClient((*grpc.ClientConn)(clientConn))
		},
	),

	domain.Module,
)
