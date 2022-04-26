package app

import (
	"context"
	gqlHandler "github.com/99designs/gqlgen/graphql/handler"
	"github.com/gofiber/adaptor/v2"
	"github.com/gofiber/fiber/v2"
	"go.uber.org/fx"
	"google.golang.org/grpc"
	"kloudlite.io/apps/ci/internal/app/graph"
	"kloudlite.io/apps/ci/internal/app/graph/generated"
	"kloudlite.io/apps/ci/internal/domain"
	"kloudlite.io/grpc-interfaces/kloudlite.io/rpc/auth"
	"kloudlite.io/grpc-interfaces/kloudlite.io/rpc/ci"
	"kloudlite.io/pkg/config"
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

type ciServerImpl struct {
	ci.UnimplementedCIServer
	d domain.Domain
}

func (c *ciServerImpl) CreatePipeline(ctx context.Context, in *ci.PipelineIn) (*ci.PipelineOutput, error) {
	pipeline, err := c.d.CretePipeline(ctx, domain.Pipeline{
		Name:                 in.Name,
		ImageName:            in.ImageName,
		PipelineEnv:          in.PipelineEnv,
		GitProvider:          in.GitProvider,
		GitRepoUrl:           in.GitRepoUrl,
		DockerFile:           in.DockerFile,
		ContextDir:           in.ContextDir,
		GithubInstallationId: in.GithubInstallationId,
		GitlabTokenId:        in.GitlabTokenId,
		BuildArgs:            in.BuildArgs,
	})
	if err != nil {
		return nil, err
	}
	return &ci.PipelineOutput{PipelineId: string(pipeline.Id)}, err
}

func fxCiServer(d domain.Domain) ci.CIServer {
	return &ciServerImpl{
		d: d,
	}
}

type AuthClientConnection *grpc.ClientConn

var Module = fx.Module("app",
	fx.Provide(config.LoadEnv[Env]()),
	repos.NewFxMongoRepo[*domain.Pipeline]("pipelines", "acc", domain.PipelineIndexes),
	fx.Provide(fxCiServer),
	fx.Provide(func(conn AuthClientConnection) auth.AuthClient {
		return auth.NewAuthClient((*grpc.ClientConn)(conn))
	}),
	fx.Invoke(func(server *fiber.App, d domain.Domain) {
		server.Get("/pipeline/:id", func(ctx *fiber.Ctx) error {
			pipeline, err := d.GetPipeline(ctx.Context(), repos.ID(ctx.Params("id")))
			if err != nil {
				return err
			}
			return ctx.JSON(pipeline)
		})

		schema := generated.NewExecutableSchema(
			generated.Config{Resolvers: &graph.Resolver{Domain: d}},
		)
		gqlServer := gqlHandler.NewDefaultServer(schema)
		server.Get("/", adaptor.HTTPHandlerFunc(gqlServer.ServeHTTP))
	}),
	fx.Invoke(func(server *grpc.Server, ciServer ci.CIServer) {
		ci.RegisterCIServer(server, ciServer)
	}),
	domain.Module,
)
