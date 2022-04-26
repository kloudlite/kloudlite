package graph

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.

import (
	"context"
	"fmt"

	"kloudlite.io/apps/ci/internal/app/graph/generated"
	"kloudlite.io/apps/ci/internal/app/graph/model"
	"kloudlite.io/pkg/repos"
)

func (r *mutationResolver) CiDeleteGitPipeline(ctx context.Context, pipelineID repos.ID) (bool, error) {
	panic(fmt.Errorf("not implemented"))
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

func (r *queryResolver) CiGithubInstallations(ctx context.Context) ([]map[string]interface{}, error) {
	panic(fmt.Errorf("not implemented"))
}

func (r *queryResolver) CiGithubRepos(ctx context.Context, installationID string, limit *int, page *int) ([]map[string]interface{}, error) {
	panic(fmt.Errorf("not implemented"))
}

func (r *queryResolver) CiGithubRepoBranches(ctx context.Context, repoURL string, limit *int, page *int) ([]map[string]interface{}, error) {
	panic(fmt.Errorf("not implemented"))
}

func (r *queryResolver) CiSearchGithubRepos(ctx context.Context, search *string, org string, limit *int, page *int) ([]map[string]interface{}, error) {
	panic(fmt.Errorf("not implemented"))
}

func (r *queryResolver) CiGitPipelines(ctx context.Context, projectID repos.ID, query map[string]interface{}) ([]*model.GitPipeline, error) {
	panic(fmt.Errorf("not implemented"))
}

func (r *queryResolver) CiGitPipeline(ctx context.Context, pipelineID repos.ID) (*model.GitPipeline, error) {
	panic(fmt.Errorf("not implemented"))
}

// Mutation returns generated.MutationResolver implementation.
func (r *Resolver) Mutation() generated.MutationResolver { return &mutationResolver{r} }

// Query returns generated.QueryResolver implementation.
func (r *Resolver) Query() generated.QueryResolver { return &queryResolver{r} }

type mutationResolver struct{ *Resolver }
type queryResolver struct{ *Resolver }
