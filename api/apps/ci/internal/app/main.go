package app

import (
	"github.com/gofiber/fiber/v2"
	"go.uber.org/fx"
	"google.golang.org/grpc"
	"kloudlite.io/apps/ci/internal/app/graph"
	"kloudlite.io/apps/ci/internal/app/graph/generated"
	"kloudlite.io/apps/ci/internal/domain"
	"kloudlite.io/common"
	"kloudlite.io/grpc-interfaces/kloudlite.io/rpc/auth"
	"kloudlite.io/grpc-interfaces/kloudlite.io/rpc/ci"
	"kloudlite.io/pkg/cache"
	"kloudlite.io/pkg/config"
	httpServer "kloudlite.io/pkg/http-server"
	"kloudlite.io/pkg/repos"
)

type Env struct {
	CookieDomain     string `env:"COOKIE_DOMAIN" required:"true"`
	GithubWebhookUrl string `env:"GITHUB_WEBHOOK_URL" required:"true"`
	GitlabWebhookUrl string `env:"GITLAB_WEBHOOK_URL" required:"true"`

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

	HarborUsername string `env:"HARBOR_USERNAME" required:"true"`
	HarborPassword string `env:"HARBOR_PASSWORD" required:"true"`
	HarborUrl      string `env:"HARBOR_URL" required:"true"`
}

func (env *Env) GoogleConfig() (clientId string, clientSecret string, callbackUrl string) {
	return env.GoogleClientId, env.GoogleClientSecret, env.GoogleCallbackUrl
}

func (env *Env) GitlabConfig() (clientId string, clientSecret string, callbackUrl string) {
	return env.GitlabClientId, env.GitlabClientSecret, env.GitlabCallbackUrl
}

func (env *Env) GithubConfig() (clientId, clientSecret, callbackUrl, githubAppId, githubAppPKFile string) {
	return env.GithubClientId, env.GithubClientSecret, env.GithubCallbackUrl, env.GithubAppId, env.GithubAppPKFile
}

type AuthClientConnection *grpc.ClientConn

var Module = fx.Module("app",
	fx.Provide(config.LoadEnv[Env]()),
	// Mongo Repos
	repos.NewFxMongoRepo[*domain.Pipeline]("pipelines", "pip", domain.PipelineIndexes),

	fx.Provide(fxCiServer),
	fx.Provide(func(conn AuthClientConnection) auth.AuthClient {
		return auth.NewAuthClient((*grpc.ClientConn)(conn))
	}),

	// FiberApp
	fx.Invoke(func(app *fiber.App, d domain.Domain, github domain.Github) {
		app.Get("/pipelines/:pipeline", func(ctx *fiber.Ctx) error {
			pipeline, err := d.GetPipeline(ctx.Context(), repos.ID(ctx.Params("pipeline")))
			if err != nil {
				return err
			}
			return ctx.JSON(pipeline)
		})

		app.Get("/access-repo-token/:installation_id", func(ctx *fiber.Ctx) error {
			paramsInt, err := ctx.ParamsInt("installation_id")
			if err != nil {
				return err
			}
			token, err := github.GetInstallationToken(ctx.Context(), "", int64(paramsInt))
			return ctx.JSON(token)
		})
	}),

	// GraphQL App
	fx.Invoke(func(
		server *fiber.App,
		d domain.Domain,
		env *Env,
		cacheClient cache.Client,
	) {
		schema := generated.NewExecutableSchema(
			generated.Config{Resolvers: &graph.Resolver{Domain: d}},
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
	}),
	fx.Invoke(func(server *grpc.Server, ciServer ci.CIServer) {
		ci.RegisterCIServer(server, ciServer)
	}),
	fx.Provide(fxGitlab),
	fx.Provide(fxGithub),
	domain.Module,
)
