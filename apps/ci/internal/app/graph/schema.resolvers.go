package graph

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"kloudlite.io/apps/ci/internal/app/graph/generated"
	"kloudlite.io/apps/ci/internal/app/graph/model"
	"kloudlite.io/apps/ci/internal/domain"
	"kloudlite.io/common"
	httpServer "kloudlite.io/pkg/http-server"
	"kloudlite.io/pkg/repos"
	"kloudlite.io/pkg/types"
)

func (r *mutationResolver) CiDeleteGitPipeline(ctx context.Context, pipelineID repos.ID) (bool, error) {
	panic(fmt.Errorf("not implemented"))
}

func (r *mutationResolver) CiCreatePipeline(ctx context.Context, in model.GitPipelineIn) (map[string]interface{}, error) {
	session := httpServer.GetSession[*common.AuthSession](ctx)
	if session == nil {
		return nil, errors.New("not authorized")
	}
	pipeline, err := r.Domain.CreatePipeline(ctx, session.UserId, domain.Pipeline{
		Name:                 in.Name,
		ImageName:            in.ImageName,
		GitProvider:          in.GitProvider,
		GitBranch:            in.GitBranch,
		GitRepoUrl:           in.GitRepoURL,
		DockerFile:           in.DockerFile,
		ContextDir:           in.ContextDir,
		GitlabRepoId:         in.GitlabRepoID,
		GithubInstallationId: in.GithubInstallationID,
		BuildArgs:            in.BuildArgs,
	})
	marshal, err := json.Marshal(pipeline)
	if err != nil {
		return nil, err
	}
	x := make(map[string]any)
	err = json.Unmarshal(marshal, &x)
	if err != nil {
		return nil, err
	}
	return x, err
}

func (r *queryResolver) CiGithubInstallations(ctx context.Context, pagination *types.Pagination) (interface{}, error) {
	session := httpServer.GetSession[*common.AuthSession](ctx)
	if session == nil {
		return nil, errors.New("not authenticated")
	}
	return r.Domain.GithubListInstallations(ctx, session.UserId, pagination)
}

func (r *queryResolver) CiGithubInstallationToken(ctx context.Context, repoURL *string, instID *int) (interface{}, error) {
	if instID == nil {
		return r.Domain.GithubInstallationToken(ctx, "")
	}
	return r.Domain.GithubInstallationToken(ctx, "")
}

func (r *queryResolver) CiGithubRepos(ctx context.Context, installationID int, pagination *types.Pagination) (interface{}, error) {
	session := httpServer.GetSession[*common.AuthSession](ctx)
	if session == nil {
		return nil, errors.New("not authenticated")
	}
	return r.Domain.GithubListRepos(ctx, session.UserId, int64(installationID), pagination)
}

func (r *queryResolver) CiGithubRepoBranches(ctx context.Context, repoURL string, pagination *types.Pagination) (interface{}, error) {
	session := httpServer.GetSession[*common.AuthSession](ctx)
	if session == nil {
		return nil, errors.New("not authenticated")
	}
	branches, err := r.Domain.GithubListBranches(ctx, session.UserId, repoURL, pagination)
	return branches, err
}

func (r *queryResolver) CiSearchGithubRepos(ctx context.Context, search *string, org string, pagination *types.Pagination) (interface{}, error) {
	session := httpServer.GetSession[*common.AuthSession](ctx)
	if session == nil {
		return nil, errors.New("not authenticated")
	}
	return r.Domain.GithubSearchRepos(ctx, session.UserId, *search, org, pagination)
}

func (r *queryResolver) CiGitlabGroups(ctx context.Context, query *string, pagination *types.Pagination) (interface{}, error) {
	session := httpServer.GetSession[*common.AuthSession](ctx)
	return r.Domain.GitlabListGroups(ctx, session.UserId, query, pagination)
}

func (r *queryResolver) CiGitlabRepos(ctx context.Context, groupID string, search *string, pagination *types.Pagination) (interface{}, error) {
	session := httpServer.GetSession[*common.AuthSession](ctx)
	return r.Domain.GitlabListRepos(ctx, session.UserId, groupID, search, pagination)
}

func (r *queryResolver) CiGitlabRepoBranches(ctx context.Context, repoID string, search *string, pagination *types.Pagination) (interface{}, error) {
	session := httpServer.GetSession[*common.AuthSession](ctx)
	return r.Domain.GitlabListBranches(ctx, session.UserId, repoID, search, pagination)
}

func (r *queryResolver) CiGetPipelines(ctx context.Context, projectID repos.ID) ([]*model.GitPipeline, error) {
	pipelineEntities, err := r.Domain.GetPipelines(ctx, projectID)
	if err != nil {
		return nil, err
	}
	pipelines := make([]*model.GitPipeline, len(pipelineEntities))
	for i, pipelineE := range pipelineEntities {
		pipelines[i] = &model.GitPipeline{
			ID:                   pipelineE.Id,
			Name:                 pipelineE.Name,
			ImageName:            pipelineE.ImageName,
			GitProvider:          pipelineE.GitProvider,
			GitRepoURL:           pipelineE.GitRepoUrl,
			DockerFile:           pipelineE.DockerFile,
			ContextDir:           pipelineE.ContextDir,
			GithubInstallationID: pipelineE.GithubInstallationId,
			BuildArgs:            pipelineE.BuildArgs,
			Metadata:             pipelineE.Metadata,
		}
	}
	return pipelines, nil
}

func (r *queryResolver) CiGetPipeline(ctx context.Context, pipelineID repos.ID) (*model.GitPipeline, error) {
	pipelineE, err := r.Domain.GetPipeline(ctx, pipelineID)
	if err != nil {
		return nil, err
	}
	return &model.GitPipeline{
		ID:                   pipelineE.Id,
		Name:                 pipelineE.Name,
		ImageName:            pipelineE.ImageName,
		GitProvider:          pipelineE.GitProvider,
		GitRepoURL:           pipelineE.GitRepoUrl,
		DockerFile:           pipelineE.DockerFile,
		ContextDir:           pipelineE.ContextDir,
		GithubInstallationID: pipelineE.GithubInstallationId,
		BuildArgs:            pipelineE.BuildArgs,
		Metadata:             pipelineE.Metadata,
	}, nil
}

// Mutation returns generated.MutationResolver implementation.
func (r *Resolver) Mutation() generated.MutationResolver { return &mutationResolver{r} }

// Query returns generated.QueryResolver implementation.
func (r *Resolver) Query() generated.QueryResolver { return &queryResolver{r} }

type mutationResolver struct{ *Resolver }
type queryResolver struct{ *Resolver }
