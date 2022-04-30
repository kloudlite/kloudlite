package app

import (
	"context"
	"fmt"
	"kloudlite.io/apps/ci/internal/domain"
	"kloudlite.io/grpc-interfaces/kloudlite.io/rpc/ci"
	"kloudlite.io/pkg/errors"
	"kloudlite.io/pkg/harbor"
	"kloudlite.io/pkg/repos"
)

type server struct {
	ci.UnimplementedCIServer
	harborCli harbor.Harbor
	d         domain.Domain
	dh        domain.Harbor
}

func (s *server) CreatePipeline(ctx context.Context, in *ci.PipelineIn) (*ci.PipelineOutput, error) {
	i := int(in.GithubInstallationId)
	ba := make(map[string]interface{}, 0)
	if in.BuildArgs != nil {
		for k, v := range in.BuildArgs {
			ba[k] = v
		}
	}
	fmt.Println(in)
	pipeline, err := s.d.CretePipeline(ctx, repos.ID(in.UserId), domain.Pipeline{
		ProjectId:            in.ProjectId,
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
	userAcc, err := s.harborCli.CreateUserAccount(ctx, in.Name)
	if err != nil {
		return nil, err
	}
	fmt.Println("useracc:", userAcc)
	if err := s.dh.SaveUserAcc(ctx, &domain.HarborAccount{
		BaseEntity: repos.BaseEntity{
			Id: repos.ID(fmt.Sprintf("%d", userAcc.Id)),
		},
		ProjectName: in.Name,
		Username:    userAcc.Name,
		Password:    userAcc.Secret,
	}); err != nil {
		return nil, errors.NewEf(err, "could not save harbor user account into DB")
	}
	return &ci.HarborProjectOut{Status: true}, nil
}

func (s *server) DeleteHarborProject(ctx context.Context, in *ci.HarborProjectIn) (*ci.HarborProjectOut, error) {
	if err := s.harborCli.DeleteProject(ctx, in.Name); err != nil {
		return nil, err
	}
	return &ci.HarborProjectOut{Status: true}, nil
}

func fxCiServer(harborCli harbor.Harbor, d domain.Domain, dh domain.Harbor) ci.CIServer {
	return &server{
		harborCli: harborCli,
		d:         d,
		dh:        dh,
	}
}
