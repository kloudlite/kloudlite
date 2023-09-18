package app

import (
	"context"

	"github.com/99designs/gqlgen/graphql"
	"github.com/gofiber/fiber/v2"
	"go.uber.org/fx"
	"kloudlite.io/apps/accounts/internal/app/graph"
	"kloudlite.io/apps/accounts/internal/app/graph/generated"
	"kloudlite.io/apps/accounts/internal/domain"
	"kloudlite.io/apps/accounts/internal/entities"
	"kloudlite.io/apps/accounts/internal/env"
	"kloudlite.io/common"
	"kloudlite.io/constants"
	"kloudlite.io/grpc-interfaces/kloudlite.io/rpc/accounts"
	"kloudlite.io/grpc-interfaces/kloudlite.io/rpc/auth"
	"kloudlite.io/grpc-interfaces/kloudlite.io/rpc/comms"
	"kloudlite.io/grpc-interfaces/kloudlite.io/rpc/console"
	"kloudlite.io/grpc-interfaces/kloudlite.io/rpc/iam"
	"kloudlite.io/pkg/cache"
	"kloudlite.io/pkg/grpc"
	httpServer "kloudlite.io/pkg/http-server"
	"kloudlite.io/pkg/repos"
)

type AuthCacheClient cache.Client

type AuthClient grpc.Client

type ConsoleClient grpc.Client

type ContainerRegistryClient grpc.Client
type CommsClient grpc.Client
type IAMClient grpc.Client

var Module = fx.Module("app",
	repos.NewFxMongoRepo[*entities.Account]("accountsv2", "acc", entities.AccountIndices),
	repos.NewFxMongoRepo[*entities.Invitation]("invitations", "invite", entities.InvitationIndices),

	fx.Provide(func(client AuthCacheClient) cache.Repo[*entities.Invitation] {
		return cache.NewRepo[*entities.Invitation](client)
	}),

	// grpc clients
	fx.Provide(func(conn ConsoleClient) console.ConsoleClient {
		return console.NewConsoleClient(conn)
	}),

	// fx.Provide(func(conn ContainerRegistryClient) container_registry.ContainerRegistryClient {
	// 	return container_registry.NewContainerRegistryClient(conn)
	// }),

	fx.Provide(func(conn IAMClient) iam.IAMClient {
		return iam.NewIAMClient(conn)
	}),

	fx.Provide(func(conn CommsClient) comms.CommsClient {
		return comms.NewCommsClient(conn)
	}),

	fx.Provide(func(conn AuthClient) auth.AuthClient {
		return auth.NewAuthClient(conn)
	}),

	domain.Module,

	fx.Invoke(
		func(server httpServer.Server, d domain.Domain, env *env.Env, cacheClient AuthCacheClient) {
			gqlConfig := generated.Config{Resolvers: graph.NewResolver(d)}

			gqlConfig.Directives.IsLoggedInAndVerified = func(ctx context.Context, obj interface{}, next graphql.Resolver) (res interface{}, err error) {
				sess := httpServer.GetSession[*common.AuthSession](ctx)
				if sess == nil {
					return nil, fiber.ErrUnauthorized
				}

				if !sess.UserVerified {
					return nil, fiber.ErrForbidden
				}

				return next(context.WithValue(ctx, "kloudlite-user-session", *sess))
			}

			schema := generated.NewExecutableSchema(gqlConfig)
			server.SetupGraphqlServer(schema,
				httpServer.NewSessionMiddleware[*common.AuthSession](
					cacheClient,
					constants.CookieName,
					env.CookieDomain,
					constants.CacheSessionPrefix,
				),
			)
		},
	),

	fx.Provide(func(d domain.Domain) accounts.AccountsServer {
		return &accountsGrpcServer{d: d}
	}),
)
