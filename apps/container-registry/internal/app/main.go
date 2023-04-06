package app

import (
	"context"
	"fmt"
	"github.com/99designs/gqlgen/graphql"
	"github.com/gofiber/fiber/v2"
	"go.uber.org/fx"
	"google.golang.org/grpc"
	"kloudlite.io/apps/container-registry/internal/app/graph"
	"kloudlite.io/apps/container-registry/internal/app/graph/generated"
	"kloudlite.io/apps/container-registry/internal/domain"
	"kloudlite.io/apps/container-registry/internal/domain/entities"
	"kloudlite.io/apps/container-registry/internal/env"
	"kloudlite.io/common"
	"kloudlite.io/constants"
	"kloudlite.io/grpc-interfaces/kloudlite.io/rpc/container_registry"
	"kloudlite.io/pkg/cache"
	"kloudlite.io/pkg/harbor"
	httpServer "kloudlite.io/pkg/http-server"
	"kloudlite.io/pkg/repos"
)

type AuthCacheClient cache.Client

var Module = fx.Module("app",
	repos.NewFxMongoRepo[*entities.HarborProject]("project", "prj", entities.HarborProjectIndexes),

	fx.Provide(fxRPCServer),
	fx.Invoke(
		func(server *grpc.Server, crServer container_registry.ContainerRegistryServer) {
			container_registry.RegisterContainerRegistryServer(server, crServer)
		},
	),

	fx.Provide(
		func(ev *env.Env) (*harbor.Client, error) {
			return harbor.NewClient(
				harbor.Args{
					HarborAdminUsername: ev.HarborAdminUsername,
					HarborAdminPassword: ev.HarborAdminPassword,
					HarborRegistryHost:  ev.HarborRegistryHost,
				},
			)
		},
	),

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

			gqlConfig.Directives.HasAccount = func(ctx context.Context, obj interface{}, next graphql.Resolver) (res interface{}, err error) {
				sess := httpServer.GetSession[*common.AuthSession](ctx)
				if sess == nil {
					return nil, fiber.ErrUnauthorized
				}

				m := httpServer.GetHttpCookies(ctx)
				klAccount := m["kloudlite-account"]
				if klAccount == "" {
					return nil, fmt.Errorf("no cookie named '%s' present in request", "kloudlite-account")
				}

				cc := domain.NewRegistryContext(ctx, sess.UserId, klAccount)
				return next(context.WithValue(ctx, "kloudlite-ctx", cc))
			}

			gqlConfig.Directives.CanActOnAccount = func(ctx context.Context, obj interface{}, next graphql.Resolver, action *string) (res interface{}, err error) {
				if action == nil {

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
					constants.CacheSessionPrefix,
				),
			)
		},
	),
	domain.Module,
)
