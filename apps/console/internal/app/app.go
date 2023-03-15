package app

import (
	"context"

	"github.com/99designs/gqlgen/graphql"
	"github.com/gofiber/fiber/v2"
	"go.uber.org/fx"
	"kloudlite.io/apps/console/internal/app/graph"
	"kloudlite.io/apps/console/internal/app/graph/generated"
	domain "kloudlite.io/apps/console/internal/domain"
	"kloudlite.io/apps/console/internal/domain/entities"
	"kloudlite.io/apps/console/internal/env"
	"kloudlite.io/common"
	"kloudlite.io/constants"
	"kloudlite.io/pkg/cache"
	httpServer "kloudlite.io/pkg/http-server"
	"kloudlite.io/pkg/repos"
)

type AuthCacheClient cache.Client

var Module = fx.Module("app",
	repos.NewFxMongoRepo[*entities.Project]("projects", "prj", entities.ProjectIndexes),
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

				a := ctx.Value("hello")
				print(a)

				return next(ctx)
			}

			schema := generated.NewExecutableSchema(gqlConfig)
			httpServer.SetupGQLServer(
				server,
				schema,
				func(c *fiber.Ctx) error {
					klAccount := c.Cookies("kloudlite-account")
					klCluster := c.Cookies("kloudlite-cluster")

					ctx := domain.NewConsoleContext(c.UserContext(), klAccount, klCluster)
					// ctx := context.WithValue(c.UserContext(), "kloudlite-account", klAccount)
					// ctx = context.WithValue(ctx, "kloudlite-cluster", klCluster)
					c.SetUserContext(context.WithValue(c.UserContext(), "hello", ctx))
					return c.Next()
				},
				httpServer.NewSessionMiddleware[*common.AuthSession](
					cacheClient,
					"hotspot-session",
					ev.CookieDomain,
					ev.AuthRedisPrefix+":"+constants.CacheSessionPrefix,
				),
			)
		},
	),
	domain.Module,
)
