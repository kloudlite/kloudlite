package app

import (
	"encoding/json"
	"fmt"
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
	"kloudlite.io/pkg/errors"
	"kloudlite.io/pkg/harbor"
	httpServer "kloudlite.io/pkg/http-server"
	"kloudlite.io/pkg/repos"
	"kloudlite.io/pkg/tekton"
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

func (env *Env) GetHarborConfig() (username, password, registryUrl string) {
	return env.HarborUsername, env.HarborPassword, env.HarborUrl
}

type AuthCacheClient cache.Client
type CacheClient cache.Client

type AuthClientConnection *grpc.ClientConn

var Module = fx.Module(
	"app",
	fx.Provide(config.LoadEnv[Env]()),
	// Mongo Repos
	repos.NewFxMongoRepo[*domain.Pipeline]("pipelines", "pip", domain.PipelineIndexes),
	repos.NewFxMongoRepo[*domain.HarborAccount]("harbor-accounts", "harbor_acc", []repos.IndexField{}),

	fx.Provide(fxCiServer),
	fx.Provide(
		func(conn AuthClientConnection) auth.AuthClient {
			return auth.NewAuthClient((*grpc.ClientConn)(conn))
		},
	),

	fx.Provide(
		func(env *Env) (harbor.Harbor, error) {
			return harbor.NewClient(env)
		},
	),

	// FiberApp
	fx.Invoke(
		func(app *fiber.App, d domain.Domain, gitlab domain.Gitlab) {
			app.Get(
				"/healthy", func(ctx *fiber.Ctx) error {
					return ctx.JSON("hello world")
				},
			)
			app.Get(
				"/pipelines/:pipeline", func(ctx *fiber.Ctx) error {
					pipeline, err := d.GetPipeline(ctx.Context(), repos.ID(ctx.Params("pipeline")))
					if err != nil {
						return err
					}
					return ctx.JSON(pipeline)
				},
			)

			app.Get(
				"/access-token/:provider/:pipelineId", func(ctx *fiber.Ctx) error {
					provider := ctx.Params("provider")
					pipelineId := ctx.Params("pipelineId")
					if provider == "gitlab" {
						token, err := d.GitlabPullToken(ctx.Context(), repos.ID(pipelineId))
						if err != nil {
							return errors.NewEf(err, "while getting gitlab pull token")
						}
						return ctx.JSON(token)
					}

					if provider == "github" {
						token, err := d.GithubInstallationToken(ctx.Context(), repos.ID(pipelineId))
						if err != nil {
							return errors.NewEf(err, "while getting gitlab pull token")
						}
						return ctx.JSON(token)
					}
					return errors.Newf("unknown (provider=%s) not one of [github,gitlab]", provider)
				},
			)
		},
	),

	// Tekton Interceptor
	fx.Invoke(
		func(app *fiber.App, d domain.Domain) error {
			app.Get(
				"/tekton/interceptor", func(ctx *fiber.Ctx) error {
					return ctx.JSON(map[string]string{"hello": "world"})
				},
			)
			app.Post(
				"/tekton/interceptor/:gitProvider", func(ctx *fiber.Ctx) error {
					fmt.Println("HERE, req received ....")

					gitProvider := ctx.Params("gitProvider")

					var req tekton.Request
					err := json.Unmarshal(ctx.Body(), &req)
					if err != nil {
						return err
					}

					switch gitProvider {
					case common.ProviderGithub:
						{
							resp := d.TektonInterceptorGithub(ctx.Context(), &req)
							jsonBody, err := resp.ToJson()
							if err != nil {
								return ctx.JSON(err)
							}

							fmt.Printf("jsonBody: %s\n", jsonBody)
							return ctx.Send(jsonBody)
						}
					case common.ProviderGitlab:
						{
							resp := d.TektonInterceptorGithub(ctx.Context(), &req)
							return ctx.JSON(resp)
						}
					}

					return nil
				},
			)
			return nil
		},
	),

	// GraphQL App
	fx.Invoke(
		func(
			server *fiber.App,
			d domain.Domain,
			env *Env,
			cacheClient AuthCacheClient,
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
		},
	),

	fx.Invoke(
		func(server *grpc.Server, ciServer ci.CIServer) {
			ci.RegisterCIServer(server, ciServer)
		},
	),
	fx.Provide(fxGitlab),
	fx.Provide(fxGithub),
	domain.Module,
)
