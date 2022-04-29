package app

import (
	"context"
	"kloudlite.io/apps/ci/internal/domain"
	"kloudlite.io/grpc-interfaces/kloudlite.io/rpc/ci"
	"kloudlite.io/pkg/harbor"
	"kloudlite.io/pkg/repos"
)

type server struct {
	ci.UnimplementedCIServer
	harborCli harbor.Harbor
	d         domain.Domain
}

func (s *server) CreatePipeline(ctx context.Context, in *ci.PipelineIn) (*ci.PipelineOutput, error) {
	i := int(in.GithubInstallationId)
	ba := make(map[string]interface{}, 0)
	if in.BuildArgs != nil {
		for k, v := range in.BuildArgs {
			ba[k] = v
		}
	}
	pipeline, err := s.d.CretePipeline(ctx, repos.ID(in.UserId), domain.Pipeline{
		Name:                 in.Name,
		ImageName:            in.ImageName,
		PipelineEnv:          in.PipelineEnv,
		GitProvider:          in.GitProvider,
		GitRepoUrl:           in.GitRepoUrl,
		DockerFile:           &in.DockerFile,
		ContextDir:           &in.ContextDir,
		GithubInstallationId: &i,
		GitlabTokenId:        in.GitlabTokenId,
		BuildArgs:            ba,
	})
	if err != nil {
		return nil, err
	}
	return &ci.PipelineOutput{PipelineId: string(pipeline.Id)}, err
}

func (s *server) CreateHarborProject(ctx context.Context, in *ci.HarborProjectIn) (*ci.HarborProjectOut, error) {
	if err := s.harborCli.CreateProject(ctx, in.Name); err != nil {
		return nil, err
	}
	return &ci.HarborProjectOut{Status: true}, nil
}

func (s *server) DeleteHarborProject(ctx context.Context, in *ci.HarborProjectIn) (*ci.HarborProjectOut, error) {
	if err := s.harborCli.DeleteProject(ctx, in.Name); err != nil {
		return nil, err
	}
	return &ci.HarborProjectOut{Status: true}, nil
}

func fxCiServer(env *Env, harborCli harbor.Harbor, d domain.Domain) ci.CIServer {
	return &server{
		harborCli: harborCli,
		d:         d,
	}
}
