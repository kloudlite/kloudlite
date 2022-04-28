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
	"kloudlite.io/pkg/cache"
	"kloudlite.io/pkg/repos"
)

func (r *mutationResolver) CiDeleteGitPipeline(ctx context.Context, pipelineID repos.ID) (bool, error) {
	panic(fmt.Errorf("not implemented"))
}

func (r *mutationResolver) CiCreatePipeline(ctx context.Context, in model.GitPipelineIn) (map[string]interface{}, error) {
	session := cache.GetSession[*common.AuthSession](ctx)
	if session == nil {
		return nil, errors.New("not authorized")
	}
	pipeline, err := r.Domain.CretePipeline(ctx, session.UserId, domain.Pipeline{
		Name:                 in.Name,
		ImageName:            in.ImageName,
		GitProvider:          in.GitProvider,
		GitRepoUrl:           in.GitRepoURL,
		DockerFile:           in.DockerFile,
		ContextDir:           in.ContextDir,
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

func (r *queryResolver) CiGitlabRepos(ctx context.Context, groupID repos.ID, search *string, limit *int, page *int) ([]map[string]interface{}, error) {
	panic(fmt.Errorf("not implemented"))
}

func (r *queryResolver) CiGitlabGroups(ctx context.Context, search *string, limit *int, page *int) ([]map[string]interface{}, error) {
	panic(fmt.Errorf("not implemented"))
}

func (r *queryResolver) CiGitlabRepoBranches(ctx context.Context, repoURL string, search *string) ([]map[string]interface{}, error) {
	panic(fmt.Errorf("not implemented"))
}

func (r *queryResolver) CiGithubInstallations(ctx context.Context) (interface{}, error) {
	session := cache.GetSession[*common.AuthSession](ctx)
	if session == nil {
		return nil, errors.New("not authenticated")
	}
	return r.Domain.GithubListInstallations(ctx, session.UserId)
}

func (r *queryResolver) CiGithubInstallationToken(ctx context.Context, repoURL *string, instID *int) (interface{}, error) {
	if instID == nil {
		return r.Domain.GithubInstallationToken(ctx, *repoURL, 0)
	}
	return r.Domain.GithubInstallationToken(ctx, "", int64(*instID))
}

func (r *queryResolver) CiGithubRepos(ctx context.Context, installationID int, limit *int, page *int) (interface{}, error) {
	session := cache.GetSession[*common.AuthSession](ctx)
	if session == nil {
		return nil, errors.New("not authenticated")
	}
	p, l := 1, 20
	if page != nil {
		p = *page
	}
	if limit != nil {
		p = *limit
	}
	return r.Domain.GithubListRepos(ctx, session.UserId, int64(installationID), p, l)
}

func (r *queryResolver) CiGithubRepoBranches(ctx context.Context, repoURL string, limit *int, page *int) (interface{}, error) {
	session := cache.GetSession[*common.AuthSession](ctx)
	if session == nil {
		return nil, errors.New("not authenticated")
	}
	p, l := 1, 20
	if page != nil {
		p = *page
	}
	if limit != nil {
		p = *limit
	}
	branches, err := r.Domain.GithubListBranches(ctx, session.UserId, repoURL, p, l)
	return branches, err
}

func (r *queryResolver) CiSearchGithubRepos(ctx context.Context, search *string, org string, limit *int, page *int) (interface{}, error) {
	session := cache.GetSession[*common.AuthSession](ctx)
	if session == nil {
		return nil, errors.New("not authenticated")
	}
	p, l := 1, 20
	if page != nil {
		p = *page
	}
	if limit != nil {
		p = *limit
	}
	return r.Domain.GithubSearchRepos(ctx, session.UserId, *search, org, p, l)
}

func (r *queryResolver) CiGetPipelines(ctx context.Context, projectID repos.ID) ([]*model.GitPipeline, error) {
	panic(fmt.Errorf("not implemented"))
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
	}, nil
}

// Mutation returns generated.MutationResolver implementation.
func (r *Resolver) Mutation() generated.MutationResolver { return &mutationResolver{r} }

// Query returns generated.QueryResolver implementation.
func (r *Resolver) Query() generated.QueryResolver { return &queryResolver{r} }

type mutationResolver struct{ *Resolver }
type queryResolver struct{ *Resolver }
