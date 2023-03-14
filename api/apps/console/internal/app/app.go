package app

import (
	"context"
	"fmt"

	"github.com/99designs/gqlgen/graphql"
	"github.com/gofiber/fiber/v2"
	"go.uber.org/fx"
	"kloudlite.io/apps/console/internal/app/graph"
	"kloudlite.io/apps/console/internal/app/graph/generated"
	domain "kloudlite.io/apps/console/internal/domain"
	"kloudlite.io/apps/console/internal/domain/entities"
	"kloudlite.io/apps/console/internal/env"
	"kloudlite.io/common"
	"kloudlite.io/pkg/cache"
	httpServer "kloudlite.io/pkg/http-server"
	"kloudlite.io/pkg/repos"
)

type AuthCacheClient cache.Client

var Module = fx.Module("app",
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
				if httpServer.GetSession[*common.AuthSession](ctx) == nil {
					return nil, fiber.ErrUnauthorized
				}

				return next(ctx)

			}

			schema := generated.NewExecutableSchema(gqlConfig)
			httpServer.SetupGQLServer(
				server,
				schema,
				httpServer.NewSessionMiddleware[*common.AuthSession](
					cacheClient,
					"hotspot-session",
					ev.CookieDomain,
					ev.AuthRedisPrefix,
				),
			)
		},
	),
	domain.Module,
)
