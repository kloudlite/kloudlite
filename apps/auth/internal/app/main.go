package app

import (
	"net/http"

	"go.uber.org/fx"
	"kloudlite.io/apps/auth/internal/app/graph"
	"kloudlite.io/apps/auth/internal/app/graph/generated"
	"kloudlite.io/apps/auth/internal/domain"
	"kloudlite.io/common"
	"kloudlite.io/pkg/cache"
	"kloudlite.io/pkg/config"
	httpServer "kloudlite.io/pkg/http-server"
	"kloudlite.io/pkg/repos"
)

type Env struct {
	CookieDomain string `env:"COOKIE_DOMAIN" required:"true"`

	GithubClientId     string `env:"GITHUB_CLIENT_ID" required:"true"`
	GithubClientSecret string `env:"GITHUB_CLIENT_SECRET" required:"true"`
	GithubCallbackUrl  string `env:"GITHUB_CALLBACK_URL" required:"true"`
	GithubAppId        string `env:"GITHUB_APP_ID" required:"true"`
	GithubAppPKFile    string `env:"GITHUB_APP_PK_FILE" required:"true"`

	GitlabClientId     string `env:"GITLAB_CLIENT_ID" required:"true"`
	GitlabClientSecret string `env:"GITLAB_CLIENT_SECRET" required:"true"`
	GitlabCallbackUrl  string `env:"GITLAB_CALLBACK_URL" required:"true"`

	GoogleClientId     string `env:"GOOGLE_CLIENT_ID" required:"true"`
	GoogleClientSecret string `env:"GOOGLE_CLIENT_SECRET" required:"true"`
	GoogleCallbackUrl  string `env:"GOOGLE_CALLBACK_URL" required:"true"`
}

func (cfg *Env) GithubConfig() (clientId string, clientSecret string, callbackUrl string) {
	return cfg.GithubClientId, cfg.GithubClientSecret, cfg.GithubCallbackUrl
}

var Module = fx.Module("app",
	config.EnvFx[Env](),
	repos.NewFxMongoRepo[*domain.User]("users", "usr", domain.UserIndexes),
	repos.NewFxMongoRepo[*domain.AccessToken]("access_tokens", "tkn", domain.AccessTokenIndexes),
	cache.NewFxRepo[*domain.VerifyToken](),
	cache.NewFxRepo[*domain.ResetPasswordToken](),

	fx.Provide(fxGithub),

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
