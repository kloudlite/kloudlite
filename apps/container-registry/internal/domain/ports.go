package domain

import (
	"context"

	"github.com/google/go-github/v45/github"
	"github.com/xanzy/go-gitlab"
	"golang.org/x/oauth2"
	"kloudlite.io/apps/container-registry/internal/domain/entities"
	"kloudlite.io/pkg/types"
)

type Github interface {
	Callback(ctx context.Context, code, state string) (*github.User, *oauth2.Token, error)
	GetToken(ctx context.Context, token *oauth2.Token) (*oauth2.Token, error)
	GetInstallationToken(ctx context.Context, repoUrl string) (string, error)

	ListInstallations(ctx context.Context, accToken *entities.AccessToken, pagination *types.Pagination) ([]*github.Installation, error)
	ListRepos(ctx context.Context, accToken *entities.AccessToken, instId int64, pagination *types.Pagination) (*github.ListRepositories, error)
	SearchRepos(ctx context.Context, accToken *entities.AccessToken, q, org string, pagination *types.Pagination) (*github.RepositoriesSearchResult, error)
	ListBranches(ctx context.Context, accToken *entities.AccessToken, repoUrl string, pagination *types.Pagination) ([]*github.Branch, error)
	CheckWebhookExists(ctx context.Context, token *entities.AccessToken, repoUrl string, webhookId *entities.GithubWebhookId) (bool, error)
	AddWebhook(ctx context.Context, accToken *entities.AccessToken, repoUrl string, webhookUrl string) (*entities.GithubWebhookId, error)
	DeleteWebhook(ctx context.Context, accToken *entities.AccessToken, repoUrl string, hookId entities.GithubWebhookId) error
	GetLatestCommit(ctx context.Context, accToken *entities.AccessToken, repoUrl string, branchName string) (string, error)
}

type Gitlab interface {
	Callback(ctx context.Context, code, state string) (*gitlab.User, *oauth2.Token, error)
	ListGroups(ctx context.Context, token *entities.AccessToken, query *string, pagination *types.Pagination) ([]entities.GitlabGroup, error)
	ListRepos(ctx context.Context, token *entities.AccessToken, gid string, query *string, pagination *types.Pagination) ([]*gitlab.Project, error)
	ListBranches(ctx context.Context, token *entities.AccessToken, repoId string, query *string, pagination *types.Pagination) ([]*gitlab.Branch, error)
	CheckWebhookExists(ctx context.Context, token *entities.AccessToken, repoId string, webhookId *entities.GitlabWebhookId) (bool, error)
	AddWebhook(ctx context.Context, token *entities.AccessToken, repoId string, pipelineId string) (*entities.GitlabWebhookId, error)
	DeleteWebhook(ctx context.Context, token *entities.AccessToken, repoUrl string, hookId entities.GitlabWebhookId) error
	RepoToken(ctx context.Context, token *entities.AccessToken) (*oauth2.Token, error)
	GetRepoId(repoUrl string) string
	GetLatestCommit(ctx context.Context, token *entities.AccessToken, repoUrl string, branchName string) (string, error)
	GetTriggerWebhookUrl() string
}
