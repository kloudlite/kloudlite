package app

import (
	"context"

	"github.com/99designs/gqlgen/graphql"
	"github.com/gofiber/fiber/v2"
	"github.com/kloudlite/api/apps/accounts/internal/app/graph"
	"github.com/kloudlite/api/apps/accounts/internal/app/graph/generated"
	"github.com/kloudlite/api/apps/accounts/internal/domain"
	"github.com/kloudlite/api/apps/accounts/internal/entities"
	"github.com/kloudlite/api/apps/accounts/internal/env"
	"github.com/kloudlite/api/common"
	"github.com/kloudlite/api/constants"
	"github.com/kloudlite/api/grpc-interfaces/kloudlite.io/rpc/accounts"
	"github.com/kloudlite/api/grpc-interfaces/kloudlite.io/rpc/auth"
	"github.com/kloudlite/api/grpc-interfaces/kloudlite.io/rpc/comms"
	"github.com/kloudlite/api/grpc-interfaces/kloudlite.io/rpc/console"
	"github.com/kloudlite/api/grpc-interfaces/kloudlite.io/rpc/iam"
	"github.com/kloudlite/api/pkg/cache"
	"github.com/kloudlite/api/pkg/grpc"
	httpServer "github.com/kloudlite/api/pkg/http-server"
	"github.com/kloudlite/api/pkg/repos"
	"go.uber.org/fx"
)

type AuthCacheClient cache.Client

type AuthClient grpc.Client

type ConsoleClient grpc.Client

type (
	ContainerRegistryClient grpc.Client
	CommsClient             grpc.Client
	IAMClient               grpc.Client
)

var Module = fx.Module("app",
	repos.NewFxMongoRepo[*entities.Account]("accounts", "acc", entities.AccountIndices),
	repos.NewFxMongoRepo[*entities.Invitation]("invitations", "invite", entities.InvitationIndices),

	fx.Provide(func(client AuthCacheClient) cache.Repo[*entities.Invitation] {
		return cache.NewRepo[*entities.Invitation](client)
	}),

	// grpc clients
	fx.Provide(func(conn ConsoleClient) console.ConsoleClient {
		return console.NewConsoleClient(conn)
	}),

	fx.Provide(func(conn IAMClient) iam.IAMClient {
		return iam.NewIAMClient(conn)
	}),

	fx.Provide(func(conn CommsClient) comms.CommsClient {
		return comms.NewCommsClient(conn)
	}),

	fx.Provide(func(conn AuthClient) auth.AuthClient {
		return auth.NewAuthClient(conn)
	}),

	fx.Provide(func(d domain.Domain) accounts.AccountsServer {
		return &accountsGrpcServer{d: d}
	}),

	fx.Invoke(func(d domain.Domain, gserver AccountsGrpcServer) {
		registerAccountsGRPCServer(gserver, d)
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
)
