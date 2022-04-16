package app

import (
	"go.uber.org/fx"
	"kloudlite.io/apps/auth/internal/app/graph"
	"kloudlite.io/apps/auth/internal/app/graph/generated"
	"kloudlite.io/apps/auth/internal/domain"
	"kloudlite.io/common"
	"kloudlite.io/pkg/cache"
	"kloudlite.io/pkg/config"
	httpServer "kloudlite.io/pkg/http-server"
	"kloudlite.io/pkg/repos"
	"net/http"
)

type Env struct {
	CookieDomain string `env:"COOKIE_DOMAIN" required:"true"`
}

var Module = fx.Module("app",
	fx.Provide(config.LoadEnv[Env]()),
	repos.NewFxMongoRepo[*domain.User]("users", "usr", domain.UserIndexes),
	repos.NewFxMongoRepo[*domain.AccessToken]("access_tokens", "tkn", domain.AccessTokenIndexes),
	cache.NewFxRepo[*domain.VerifyToken](),
	cache.NewFxRepo[*domain.ResetPasswordToken](),
	fx.Provide(fxMessenger),
	fx.Invoke(func(
		server *http.ServeMux,
		d domain.Domain,
		env *Env,
		cacheClient cache.Client,
	) {
		schema := generated.NewExecutableSchema(
			generated.Config{Resolvers: graph.NewResolver(d)},
		)
		httpServer.SetupGQLServer(
			server,
			schema,
			cache.NewSessionRepo[*common.AuthSession](
				cacheClient,
				common.CookieName,
				env.CookieDomain,
				common.CacheSessionPrefix,
			),
		)
	}),
	domain.Module,
)
