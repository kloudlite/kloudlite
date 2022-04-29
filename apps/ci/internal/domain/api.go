package domain

import (
	"context"
	"github.com/xanzy/go-gitlab"
	"kloudlite.io/pkg/types"

	"kloudlite.io/pkg/repos"
)

type Domain interface {
	GetPipeline(ctx context.Context, pipelineId repos.ID) (*Pipeline, error)
	CretePipeline(ctx context.Context, userId repos.ID, pipeline Pipeline) (*Pipeline, error)

	GithubInstallationToken(ctx context.Context, repoUrl string, instId int64) (string, error)
	GithubListInstallations(ctx context.Context, userId repos.ID) (any, error)
	GithubListRepos(ctx context.Context, userId repos.ID, installationId int64, page, size int) (any, error)
	GithubSearchRepos(ctx context.Context, userId repos.ID, q string, org string, page, size int) (any, error)
	GithubListBranches(ctx context.Context, userId repos.ID, repoUrl string, page, size int) (any, error)
	GithubAddWebhook(ctx context.Context, userId repos.ID, refId string, repoUrl string) error

	GitlabListGroups(ctx context.Context, userId repos.ID, query *string, pagination *types.Pagination) (any, error)
	GitlabListRepos(ctx context.Context, userId repos.ID, gid string, query *string, pagination *types.Pagination) (any, error)
	GitlabListBranches(ctx context.Context, userId repos.ID, repoId string, query *string, pagination *types.Pagination) (any, error)
	GitlabAddWebhook(ctx context.Context, userId repos.ID, repoId string, pipelineId string) (*gitlab.ProjectHook, error)
}

type Harbor interface {
	SaveUserAcc(ctx context.Context, acc *HarborAccount) error
}
