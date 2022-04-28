package app

import (
	"context"
	"kloudlite.io/apps/ci/internal/domain"
	"kloudlite.io/grpc-interfaces/kloudlite.io/rpc/ci"
	"kloudlite.io/pkg/repos"
)

type ciServerImpl struct {
	ci.UnimplementedCIServer
	d domain.Domain
}

func (c *ciServerImpl) CreatePipeline(ctx context.Context, in *ci.PipelineIn) (*ci.PipelineOutput, error) {
	i := int(in.GithubInstallationId)
	ba := make(map[string]interface{}, 0)
	if in.BuildArgs != nil {
		for k, v := range in.BuildArgs {
			ba[k] = v
		}
	}
	pipeline, err := c.d.CretePipeline(ctx, repos.ID(in.UserId), domain.Pipeline{
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
