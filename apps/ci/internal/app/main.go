package app

import (
	"encoding/json"

	"github.com/gofiber/fiber/v2"
	"go.uber.org/fx"
	"google.golang.org/grpc"
	"kloudlite.io/apps/ci/internal/app/graph"
	"kloudlite.io/apps/ci/internal/app/graph/generated"
	"kloudlite.io/apps/ci/internal/domain"
	"kloudlite.io/common"
	"kloudlite.io/grpc-interfaces/kloudlite.io/rpc/auth"
	"kloudlite.io/grpc-interfaces/kloudlite.io/rpc/console"
	"kloudlite.io/pkg/cache"
	"kloudlite.io/pkg/config"
	"kloudlite.io/pkg/errors"
	"kloudlite.io/pkg/harbor"
	httpServer "kloudlite.io/pkg/http-server"
	"kloudlite.io/pkg/logging"
	"kloudlite.io/pkg/redpanda"
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
	GithubScopes       string `env:"GITHUB_SCOPES" required:"true"`

	GitlabClientId     string `env:"GITLAB_CLIENT_ID" required:"true"`
	GitlabClientSecret string `env:"GITLAB_CLIENT_SECRET" required:"true"`
	GitlabCallbackUrl  string `env:"GITLAB_CALLBACK_URL" required:"true"`
	GitlabScopes       string `env:"GITLAB_SCOPES" required:"true"`

	GoogleClientId     string `env:"GOOGLE_CLIENT_ID" required:"true"`
	GoogleClientSecret string `env:"GOOGLE_CLIENT_SECRET" required:"true"`
	GoogleCallbackUrl  string `env:"GOOGLE_CALLBACK_URL" required:"true"`
	GoogleScopes       string `env:"GOOGLE_SCOPES" required:"true"`

	KafkaGitWebhooksTopic      string `env:"KAFKA_GIT_WEBHOOKS_TOPIC" required:"true"`
	KafkaGitWebhooksConsumerId string `env:"KAFKA_GIT_WEBHOOKS_CONSUMER_ID" required:"true"`
	KafkaApplyYamlTopic        string `env:"KAFKA_APPLY_YAML_TOPIC" required:"true"`
	KafkaBrokers               string `env:"KAFKA_BROKERS" required:"true"`

	// KAFKA_GIT_WEBHOOKS_TOPIC="kl-git-webhooks"
	// KAFKA_BROKERS="redpanda.kl-init-redpanda.svc.cluster.local"

	HarborAdminUsername string `env:"HARBOR_ADMIN_USERNAME" required:"true"`
	HarborAdminPassword string `env:"HARBOR_ADMIN_PASSWORD" required:"true"`
	HarborRegistryHost  string `env:"HARBOR_REGISTRY_HOST" required:"true"`

	GithubWebhookAuthzSecret string `env:"GITHUB_WEBHOOK_AUTHZ_SECRET" required:"true"`
	GitlabWebhookAuthzSecret string `env:"GITLAB_WEBHOOK_AUTHZ_SECRET" required:"true"`
	KlHookTriggerAuthzSecret string `env:"KL_HOOK_TRIGGER_AUTHZ_SECRET" required:"true"`
}

func (env *Env) GetBrokerHosts() string {
	return env.KafkaBrokers
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

func (env *Env) GetSubscriptionTopics() []string {
	return []string{env.KafkaGitWebhooksTopic}
}

func (env *Env) GetConsumerGroupId() string {
	return env.KafkaGitWebhooksConsumerId
}

type AuthCacheClient cache.Client
type CacheClient cache.Client

type AuthGRPCClient *grpc.ClientConn
type ConsoleGRPCClient *grpc.ClientConn

var Module = fx.Module(
	"app",

	fx.Provide(config.LoadEnv[Env]()),
	// Mongo Repos
	repos.NewFxMongoRepo[*domain.Pipeline]("pipelines", "pip", domain.PipelineIndexes),
	repos.NewFxMongoRepo[*domain.GitRepositoryHook]("git_repo_hooks", "grh", domain.GitRepositoryHookIndices),
	// repos.NewFxMongoRepo[*domain.HarborAccount]("harbor-accounts", "harbor_acc", []repos.IndexField{}),

	redpanda.NewConsumerFx[*Env](),
	redpanda.NewProducerFx[*Env](),

	fx.Provide(
		func(conn AuthGRPCClient) auth.AuthClient {
			return auth.NewAuthClient((*grpc.ClientConn)(conn))
		},
	),

	fx.Provide(
		func(conn ConsoleGRPCClient) console.ConsoleClient {
			return console.NewConsoleClient((*grpc.ClientConn)(conn))
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
					pipeline, err := d.GetPipeline(ctx.Context(), repos.ID(pipelineId))
					if err != nil {
						return err
					}
					if provider == common.ProviderGitlab {
						token, err := d.GitlabPullToken(ctx.Context(), pipeline.AccessTokenId)
						if err != nil {
							return errors.NewEf(err, "while getting gitlab pull token")
						}
						return ctx.JSON(token)
					}

					if provider == common.ProviderGithub {
						token, err := d.GithubInstallationToken(ctx.Context(), pipeline.GitRepoUrl)
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

	// Webhook
	fx.Invoke(
		func(app *fiber.App, d domain.Domain) error {
			app.Post(
				"/hooks/:gitProvider", func(ctx *fiber.Ctx) error {
					return nil
				},
			)
			return nil
		},
	),

	// Tekton Interceptor
	fx.Invoke(
		func(app *fiber.App, d domain.Domain, logger logging.Logger, consoleCli console.ConsoleClient) error {
			app.Post(
				"/tekton/interceptor/:gitProvider", func(ctx *fiber.Ctx) error {
					gitProvider := ctx.Params("gitProvider")

					var req tekton.Request
					err := json.Unmarshal(ctx.Body(), &req)
					if err != nil {
						return err
					}

					switch gitProvider {
					case common.ProviderGithub:
						{
							tkVars, _, err := d.TektonInterceptorGithub(ctx.Context(), &req)
							if err != nil {
								logger.Errorf(err, "tekton interceptor github")
								response := tekton.NewResponse(&req).Err(err)
								jsonBody, err := response.ToJson()
								if err != nil {
									return ctx.JSON(err)
								}
								logger.Debugf("ERR Response: %s", jsonBody)
								return ctx.Send(jsonBody)
							}

							if err != nil {
								return ctx.JSON(err)
							}
							tkVarsJson, err := tkVars.ToJson()
							if err != nil {
								logger.Debugf("ERR %s", err)
								return ctx.JSON(err)
							}
							responseBody, err := tekton.NewResponse(&req).Extend(tkVarsJson).Ok().ToJson()
							if err != nil {
								logger.Debugf("ERR %s", err)
								return ctx.JSON(err)
							}
							logger.Debugf("responseBody: %s\n", responseBody)
							return ctx.Send(responseBody)
						}
					case common.ProviderGitlab:
						{
							tkVars, _, err := d.TektonInterceptorGitlab(ctx.Context(), &req)
							if err != nil {
								response := tekton.NewResponse(&req).Err(err)
								jsonBody, err := response.ToJson()
								if err != nil {
									logger.Infof("ERR Response: %v", err)
									return ctx.JSON(err)
								}

								logger.Infof("ERR Response: %s", jsonBody)
								return ctx.Send(jsonBody)
							}
							tkVarsJson, err := tkVars.ToJson()
							if err != nil {
								logger.Debugf("Err Response: %v", err)
								return ctx.JSON(err)
							}
							responseBody, err := tekton.NewResponse(&req).Extend(tkVarsJson).Ok().ToJson()
							if err != nil {
								logger.Debugf("Err Response: %v", err)
								return ctx.JSON(err)
							}
							logger.Debugf("responseBody: %s\n", responseBody)
							return ctx.Send(responseBody)
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

	fx.Provide(fxGitlab),
	fx.Provide(fxGithub),

	fx.Provide(
		func(env *Env) domain.HarborHost {
			return domain.HarborHost(env.HarborRegistryHost)
		},
	),

	fx.Provide(
		func(env *Env) (*harbor.Client, error) {
			return harbor.NewClient(
				harbor.Args{
					HarborAdminUsername: env.HarborAdminUsername,
					HarborAdminPassword: env.HarborAdminPassword,
					HarborRegistryHost:  env.HarborRegistryHost,
				},
			)
		},
	),

	fx.Invoke(ProcessWebhooks),

	domain.Module,
)
