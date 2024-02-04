package domain

import (
	"context"

	"github.com/kloudlite/api/apps/container-registry/internal/domain/entities"
	"github.com/kloudlite/api/pkg/logging"
	"github.com/kloudlite/api/pkg/repos"
	"github.com/kloudlite/api/pkg/types"
)

func NewRegistryContext(parent context.Context, userId repos.ID, accountName string) RegistryContext {
	return RegistryContext{
		Context:     parent,
		UserId:      userId,
		AccountName: accountName,
	}
}

type CheckNameAvailabilityOutput struct {
	Result         bool     `json:"result"`
	SuggestedNames []string `json:"suggestedNames,omitempty"`
}

type Domain interface {
	ProcessRegistryEvents(ctx context.Context, events []entities.Event, logger logging.Logger) error

	CheckUserNameAvailability(ctx RegistryContext, username string) (*CheckNameAvailabilityOutput, error)
	// registry
	ListRepositories(ctx RegistryContext, search map[string]repos.MatchFilter, pagination repos.CursorPagination) (*repos.PaginatedRecord[*entities.Repository], error)
	CreateRepository(ctx RegistryContext, repoName string) (*entities.Repository, error)
	DeleteRepository(ctx RegistryContext, repoName string) error

	// tags
	ListRepositoryDigests(ctx RegistryContext, repoName string, search map[string]repos.MatchFilter, pagination repos.CursorPagination) (*repos.PaginatedRecord[*entities.Digest], error)
	DeleteRepositoryDigest(ctx RegistryContext, repoName string, digest string) error

	// credential
	ListCredentials(ctx RegistryContext, search map[string]repos.MatchFilter, pagination repos.CursorPagination) (*repos.PaginatedRecord[*entities.Credential], error)
	CreateCredential(ctx RegistryContext, credential entities.Credential) (*entities.Credential, error)
	DeleteCredential(ctx RegistryContext, userName string) error

	GetToken(ctx RegistryContext, username string) (string, error)
	GetTokenKey(ctx context.Context, username string, accountname string) (string, error)

	AddBuild(ctx RegistryContext, build entities.Build) (*entities.Build, error)
	UpdateBuild(ctx RegistryContext, id repos.ID, build entities.Build) (*entities.Build, error)
	UpdateBuildInternal(ctx context.Context, build *entities.Build) (*entities.Build, error)
	ListBuilds(ctx RegistryContext, repoName string, search map[string]repos.MatchFilter, pagination repos.CursorPagination) (*repos.PaginatedRecord[*entities.Build], error)
	GetBuild(ctx RegistryContext, buildId repos.ID) (*entities.Build, error)
	DeleteBuild(ctx RegistryContext, buildId repos.ID) error
	TriggerBuild(ctx RegistryContext, buildId repos.ID) error

	// webhook
	ParseGithubHook(eventType string, hookBody []byte) (*GitWebhookPayload, error)
	ParseGitlabHook(eventType string, hookBody []byte) (*GitWebhookPayload, error)

	GithubInstallationToken(ctx context.Context, repoUrl string) (string, error)
	GithubListInstallations(ctx context.Context, userId repos.ID, pagination *types.Pagination) ([]*entities.GithubInstallation, error)
	GithubListRepos(ctx context.Context, userId repos.ID, installationId int64, pagination *types.Pagination) (*entities.GithubListRepository, error)
	GithubSearchRepos(ctx context.Context, userId repos.ID, q, org string, pagination *types.Pagination) (*entities.GithubSearchRepository, error)
	GithubListBranches(ctx context.Context, userId repos.ID, repoUrl string, pagination *types.Pagination) ([]*entities.GitBranch, error)
	GithubAddWebhook(ctx context.Context, userId repos.ID, repoUrl string) (repos.ID, error)

	GitlabListGroups(ctx context.Context, userId repos.ID, query *string, pagination *types.Pagination) ([]*entities.GitlabGroup, error)
	GitlabListRepos(ctx context.Context, userId repos.ID, gid string, query *string, pagination *types.Pagination) ([]*entities.GitlabProject, error)
	GitlabListBranches(ctx context.Context, userId repos.ID, repoId string, query *string, pagination *types.Pagination) ([]*entities.GitBranch, error)
	GitlabAddWebhook(ctx context.Context, userId repos.ID, repoId string) (*int, error)
	GitlabPullToken(ctx context.Context, userId repos.ID) (string, error)

	GetBuildTemplate(obj BuildJobTemplateData) ([]byte, error)

	ListBuildsByGit(ctx context.Context, repoUrl, branch, provider string) ([]*entities.Build, error)

	AddBuildCache(ctx RegistryContext, buildCache entities.BuildCacheKey) (*entities.BuildCacheKey, error)
	UpdateBuildCache(ctx RegistryContext, id repos.ID, buildCache entities.BuildCacheKey) (*entities.BuildCacheKey, error)
	DeleteBuildCache(ctx RegistryContext, id repos.ID) error
	ListBuildCaches(ctx RegistryContext, search map[string]repos.MatchFilter, pagination repos.CursorPagination) (*repos.PaginatedRecord[*entities.BuildCacheKey], error)

	ListBuildRuns(ctx RegistryContext, repoName string, search map[string]repos.MatchFilter, pagination repos.CursorPagination) (*repos.PaginatedRecord[*entities.BuildRun], error)
	GetBuildRun(ctx RegistryContext, repoName string, runName string) (*entities.BuildRun, error)
	OnBuildRunUpdateMessage(ctx RegistryContext, buildRun entities.BuildRun) error
	OnBuildRunDeleteMessage(ctx RegistryContext, buildRun entities.BuildRun) error
	OnBuildRunApplyErrorMessage(ctx RegistryContext, clusterName string,name string, errorMsg string) error
	ListBuildsByCache(ctx RegistryContext, cacheId repos.ID, pagination repos.CursorPagination) (*repos.PaginatedRecord[*entities.Build], error)
	CreateBuildRun(ctx RegistryContext, build *entities.Build, hook *GitWebhookPayload, pullToken string) error
}
