package domain

import (
	"context"

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
}
