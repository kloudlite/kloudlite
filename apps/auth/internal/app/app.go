package app

import (
	"net/http"

	"github.com/gofiber/fiber/v2"
	"go.uber.org/fx"
	"google.golang.org/grpc"

	"kloudlite.io/apps/auth/internal/app/graph"
	"kloudlite.io/apps/auth/internal/app/graph/generated"
	"kloudlite.io/apps/auth/internal/domain"
	"kloudlite.io/apps/auth/internal/env"
	"kloudlite.io/common"
	"kloudlite.io/constants"
	"kloudlite.io/grpc-interfaces/kloudlite.io/rpc/auth"
	"kloudlite.io/grpc-interfaces/kloudlite.io/rpc/comms"
	"kloudlite.io/pkg/cache"
	httpServer "kloudlite.io/pkg/http-server"
	"kloudlite.io/pkg/repos"
)

type CommsClientConnection *grpc.ClientConn

var Module = fx.Module(
	"app",
	repos.NewFxMongoRepo[*domain.User]("users", "usr", domain.UserIndexes),
	repos.NewFxMongoRepo[*domain.AccessToken]("access_tokens", "tkn", domain.AccessTokenIndexes),
	repos.NewFxMongoRepo[*domain.RemoteLogin]("remote_logins", "rlgn", domain.RemoteTokenIndexes),
	cache.NewFxRepo[*domain.VerifyToken](),
	cache.NewFxRepo[*domain.ResetPasswordToken](),

	fx.Provide(
		func(conn CommsClientConnection) comms.CommsClient {
			return comms.NewCommsClient((*grpc.ClientConn)(conn))
		},
	),

	fx.Provide(fxGithub),
	fx.Provide(fxGitlab),
	fx.Provide(fxGoogle),

	fx.Provide(fxRPCServer),
	fx.Invoke(
		func(server *grpc.Server, authServer auth.AuthServer) {
			auth.RegisterAuthServer(server, authServer)
		},
	),

	fx.Invoke(func(server *fiber.App, cacheClient cache.Client, ev *env.Env) {
		sessionMiddleware := httpServer.NewSessionMiddleware[*common.AuthSession](cacheClient, constants.CookieName, ev.CookieDomain, constants.CacheSessionPrefix)
		// INFO: (route: `/.check/logged-in`)  is supposed to be used by nginx, as authentication url for other kloudlite services,
		// where this api acts as authentication provider
		server.Get("/.check/logged-in", sessionMiddleware, func(c *fiber.Ctx) error {
			session := httpServer.GetSession[*common.AuthSession](c.Context())
			if session == nil {
				return c.SendStatus(http.StatusUnauthorized)
			}
			return c.SendStatus(http.StatusOK)
		})
	}),

	fx.Invoke(
		func(
			server *fiber.App,
			d domain.Domain,
			ev *env.Env,
			cacheClient cache.Client,
		) {
			schema := generated.NewExecutableSchema(
				generated.Config{Resolvers: graph.NewResolver(d, ev)},
			)
			httpServer.SetupGQLServer(
				server, schema,
				httpServer.NewSessionMiddleware[*common.AuthSession](
					cacheClient,
					constants.CookieName,
					ev.CookieDomain,
					constants.CacheSessionPrefix,
				),
			)
		},
	),

	domain.Module,
)
