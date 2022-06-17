package domain

import (
	"context"
	"github.com/xanzy/go-gitlab"
	"kloudlite.io/pkg/tekton"
	"kloudlite.io/pkg/types"

	"kloudlite.io/pkg/repos"
)

type Domain interface {
	GetPipeline(ctx context.Context, pipelineId repos.ID) (*Pipeline, error)
	GetPipelines(ctx context.Context, projectId repos.ID) ([]*Pipeline, error)
	CreatePipeline(ctx context.Context, userId repos.ID, pipeline Pipeline) (*Pipeline, error)

	GithubInstallationToken(ctx context.Context, pipelineId repos.ID) (string, error)
	GithubListInstallations(ctx context.Context, userId repos.ID, pagination *types.Pagination) (any, error)
	GithubListRepos(ctx context.Context, userId repos.ID, installationId int64, pagination *types.Pagination) (any, error)
	GithubSearchRepos(ctx context.Context, userId repos.ID, q, org string, pagination *types.Pagination) (any, error)
	GithubListBranches(ctx context.Context, userId repos.ID, repoUrl string, pagination *types.Pagination) (any, error)
	GithubAddWebhook(ctx context.Context, userId repos.ID, refId string, repoUrl string) error

	GitlabListGroups(ctx context.Context, userId repos.ID, query *string, pagination *types.Pagination) (any, error)
	GitlabListRepos(ctx context.Context, userId repos.ID, gid string, query *string, pagination *types.Pagination) (any, error)
	GitlabListBranches(ctx context.Context, userId repos.ID, repoId string, query *string, pagination *types.Pagination) (any, error)
	GitlabAddWebhook(ctx context.Context, userId repos.ID, repoId string, pipelineId string) (*gitlab.ProjectHook, error)
	GitlabPullToken(ctx context.Context, pipelineId repos.ID) (string, error)

	// tekton interceptor
	TektonInterceptorGithub(ctx context.Context, req *tekton.Request) *tekton.Response
	TektonInterceptorGitlab(ctx context.Context, req *tekton.Request) *tekton.Response
}

type Harbor interface {
	SaveUserAcc(ctx context.Context, acc *HarborAccount) error
}
