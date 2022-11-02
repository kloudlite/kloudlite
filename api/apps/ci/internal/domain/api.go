package domain

import (
	"context"

	"kloudlite.io/pkg/harbor"
	"kloudlite.io/pkg/types"

	"kloudlite.io/pkg/repos"
)

type Domain interface {
	GetPipeline(ctx context.Context, pipelineId repos.ID) (*Pipeline, error)

	StartPipeline(ctx context.Context, pipelineId repos.ID, pipelineRunId repos.ID) error
	FinishPipeline(ctx context.Context, pipelineId repos.ID) error
	EndPipelineWithError(ctx context.Context, pipelineId repos.ID, err error) error

	GetPipelines(ctx context.Context, projectId repos.ID) ([]*Pipeline, error)
	GetTektonRunParams(ctx context.Context, gitProvider string, gitRepoUrl string, gitBranch string) ([]*TektonVars, error)
	GetAppPipelines(ctx context.Context, appId repos.ID) ([]*Pipeline, error)
	CreatePipeline(ctx context.Context, userId repos.ID, pipeline Pipeline) (*Pipeline, error)
	DeletePipeline(ctx context.Context, pipelineId repos.ID) error
	TriggerPipeline(ctx context.Context, userId repos.ID, pipelineId repos.ID) error

	CreateNewPipelineRun(ctx context.Context, pipelineId repos.ID) (*PipelineRun, error)
	UpdatePipelineRunStatus(ctx context.Context, pStatus PipelineRunStatus) error
	ListPipelineRuns(ctx context.Context, pipelineId repos.ID) ([]*PipelineRun, error)
	GetPipelineRun(ctx context.Context, pipelineRunId repos.ID) (*PipelineRun, error)

	ParseGithubHook(eventType string, hookBody []byte) (*GitWebhookPayload, error)
	ParseGitlabHook(eventType string, hookBody []byte) (*GitWebhookPayload, error)

	GithubInstallationToken(ctx context.Context, repoUrl string) (string, error)
	GithubListInstallations(ctx context.Context, userId repos.ID, pagination *types.Pagination) (any, error)
	GithubListRepos(ctx context.Context, userId repos.ID, installationId int64, pagination *types.Pagination) (any, error)
	GithubSearchRepos(ctx context.Context, userId repos.ID, q, org string, pagination *types.Pagination) (any, error)
	GithubListBranches(ctx context.Context, userId repos.ID, repoUrl string, pagination *types.Pagination) (any, error)
	GithubAddWebhook(ctx context.Context, userId repos.ID, repoUrl string) (repos.ID, error)

	GitlabListGroups(ctx context.Context, userId repos.ID, query *string, pagination *types.Pagination) (any, error)
	GitlabListRepos(ctx context.Context, userId repos.ID, gid string, query *string, pagination *types.Pagination) (any, error)
	GitlabListBranches(ctx context.Context, userId repos.ID, repoId string, query *string, pagination *types.Pagination) (any, error)
	GitlabAddWebhook(ctx context.Context, userId repos.ID, repoId string, pipelineId repos.ID) (repos.ID, error)
	GitlabPullToken(ctx context.Context, tokenId repos.ID) (string, error)

	// harbor
	HarborImageSearch(ctx context.Context, accountId repos.ID, q string, pagination *types.Pagination) ([]harbor.Repository, error)
	HarborImageTags(ctx context.Context, imageName string, pagination *types.Pagination) ([]harbor.ImageTag, error)
}

type Harbor interface {
	SaveUserAcc(ctx context.Context, acc *HarborAccount) error
}
