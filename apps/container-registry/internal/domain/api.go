package domain

import (
	"context"

	"kloudlite.io/apps/container-registry/internal/domain/entities"
	"kloudlite.io/pkg/repos"
	"kloudlite.io/pkg/types"
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
	ProcessEvents(ctx context.Context, events []entities.Event) error

	CheckUserNameAvailability(ctx RegistryContext, username string) (*CheckNameAvailabilityOutput, error)
	// registry
	ListRepositories(ctx RegistryContext, search map[string]repos.MatchFilter, pagination repos.CursorPagination) (*repos.PaginatedRecord[*entities.Repository], error)
	CreateRepository(ctx RegistryContext, repoName string) (*entities.Repository, error)
	DeleteRepository(ctx RegistryContext, repoName string) error

	// tags
	ListRepositoryTags(ctx RegistryContext, repoName string, search map[string]repos.MatchFilter, pagination repos.CursorPagination) (*repos.PaginatedRecord[*entities.Tag], error)
	DeleteRepositoryTag(ctx RegistryContext, repoName string, digest string) error

	// credential
	ListCredentials(ctx RegistryContext, search map[string]repos.MatchFilter, pagination repos.CursorPagination) (*repos.PaginatedRecord[*entities.Credential], error)
	CreateCredential(ctx RegistryContext, credential entities.Credential) (*entities.Credential, error)
	DeleteCredential(ctx RegistryContext, userName string) error

	GetToken(ctx RegistryContext, username string) (string, error)
	GetTokenKey(ctx context.Context, username string, accountname string) (string, error)

	AddBuild(ctx RegistryContext, build entities.Build) (*entities.Build, error)
	UpdateBuild(ctx RegistryContext, id repos.ID, build entities.Build) (*entities.Build, error)
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
	GithubListBranches(ctx context.Context, userId repos.ID, repoUrl string, pagination *types.Pagination) ([]*entities.GithubBranch, error)
	GithubAddWebhook(ctx context.Context, userId repos.ID, repoUrl string) (repos.ID, error)

	GitlabListGroups(ctx context.Context, userId repos.ID, query *string, pagination *types.Pagination) (any, error)
	GitlabListRepos(ctx context.Context, userId repos.ID, gid string, query *string, pagination *types.Pagination) (any, error)
	GitlabListBranches(ctx context.Context, userId repos.ID, repoId string, query *string, pagination *types.Pagination) (any, error)
	GitlabAddWebhook(ctx context.Context, userId repos.ID, repoId string, pipelineId repos.ID) (repos.ID, error)
	GitlabPullToken(ctx context.Context, tokenId repos.ID) (string, error)
}
