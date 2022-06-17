package domain

import (
	"context"
	"github.com/google/go-github/v43/github"
	"github.com/xanzy/go-gitlab"
	"golang.org/x/oauth2"
	"kloudlite.io/pkg/types"
)

type Github interface {
	Callback(ctx context.Context, code, state string) (*github.User, *oauth2.Token, error)
	GetToken(ctx context.Context, token *oauth2.Token) (*oauth2.Token, error)
	GetInstallationToken(ctx context.Context, repoUrl string) (string, error)
	ListInstallations(ctx context.Context, accToken *AccessToken, pagination *types.Pagination) ([]*github.Installation, error)
	ListRepos(ctx context.Context, accToken *AccessToken, instId int64, pagination *types.Pagination) (*github.ListRepositories, error)
	SearchRepos(ctx context.Context, accToken *AccessToken, q, org string, pagination *types.Pagination) (*github.RepositoriesSearchResult, error)
	ListBranches(ctx context.Context, accToken *AccessToken, repoUrl string, pagination *types.Pagination) ([]*github.Branch, error)
	AddWebhook(ctx context.Context, accToken *AccessToken, pipelineId string, repoUrl string) error
}

type Gitlab interface {
	Callback(ctx context.Context, code, state string) (*gitlab.User, *oauth2.Token, error)
	ListGroups(ctx context.Context, token *AccessToken, query *string, pagination *types.Pagination) ([]*gitlab.Group, error)
	ListRepos(ctx context.Context, token *AccessToken, gid string, query *string, pagination *types.Pagination) ([]*gitlab.Project, error)
	ListBranches(ctx context.Context, token *AccessToken, repoId string, query *string, pagination *types.Pagination) ([]*gitlab.Branch, error)
	AddWebhook(ctx context.Context, token *AccessToken, repoId string, pipelineId string) (*gitlab.ProjectHook, error)
	RemoveWebhook(ctx context.Context, token *AccessToken) error
	RepoToken(ctx context.Context, token *AccessToken) (*oauth2.Token, error)
	GetRepoId(repoUrl string) string
}
