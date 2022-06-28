package app

import (
	"github.com/gofiber/fiber/v2"
	"go.uber.org/fx"
	"kloudlite.io/apps/finance/internal/app/graph"
	"kloudlite.io/apps/finance/internal/app/graph/generated"
	"kloudlite.io/apps/finance/internal/domain"
	"kloudlite.io/common"
	"kloudlite.io/pkg/cache"
	"kloudlite.io/pkg/config"
	httpServer "kloudlite.io/pkg/http-server"
	"kloudlite.io/pkg/repos"
)

type Env struct {
	CookieDomain string `env:"COOKIE_DOMAIN" required:"true"`
}

var Module = fx.Module(
	"application",
	config.EnvFx[Env](),
	repos.NewFxMongoRepo[*domain.Account]("accounts", "acc", domain.AccountIndexes),
	CiClientFx,
	IAMClientFx,
	ConsoleClientFx,
	AuthClientFx,
	fx.Invoke(
		func(server *fiber.App, d domain.Domain, env *Env, cacheClient cache.Client) {
			schema := generated.NewExecutableSchema(
				generated.Config{Resolvers: graph.NewResolver(d)},
			)
			httpServer.SetupGQLServer(
				server,
				schema,
				httpServer.NewSessionMiddleware[*common.AuthSession](
					cacheClient,
					common.CookieName,
					env.CookieDomain,
					common.CacheSessionPrefix,
				),
			)
		},
	),
	domain.Module,
)
