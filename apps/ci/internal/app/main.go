package app

import (
	"context"
	"github.com/gofiber/fiber/v2"
	"go.uber.org/fx"
	"google.golang.org/grpc"
	"kloudlite.io/apps/ci/internal/domain"
	"kloudlite.io/grpc-interfaces/kloudlite.io/rpc/ci"
	"kloudlite.io/pkg/config"
	"kloudlite.io/pkg/repos"
)

type Env struct {
	CookieDomain string `env:"COOKIE_DOMAIN" required:"true"`
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

var Module = fx.Module("app",
	fx.Provide(config.LoadEnv[Env]()),
	repos.NewFxMongoRepo[*domain.Pipeline]("pipelines", "acc", domain.PipelineIndexes),
	fx.Provide(fxCiServer),
	fx.Invoke(func(server *fiber.App, d domain.Domain) {
		server.Get("/pipeline/:id", func(ctx *fiber.Ctx) error {
			pipeline, err := d.GetPipeline(ctx.Context(), repos.ID(ctx.Params("id")))
			if err != nil {
				return err
			}
			return ctx.JSON(pipeline)
		})
	}),

	fx.Invoke(func(server *grpc.Server, ciServer ci.CIServer) {
		ci.RegisterCIServer(server, ciServer)
	}),
	domain.Module,
)
