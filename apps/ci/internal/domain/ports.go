package domain

import (
	"context"
	"github.com/google/go-github/v43/github"
	"github.com/xanzy/go-gitlab"
	"golang.org/x/oauth2"
)

type Github interface {
	Callback(ctx context.Context, code, state string) (*github.User, *oauth2.Token, error)
	GetToken(ctx context.Context, token *oauth2.Token) (*oauth2.Token, error)
	GetInstallationToken(ctx context.Context, repoUrl string, instId int64) (string, error)
	ListInstallations(ctx context.Context, accToken *AccessToken) ([]*github.Installation, error)
	ListRepos(ctx context.Context, accToken *AccessToken, instId int64, page, size int) (*github.ListRepositories, error)
	SearchRepos(ctx context.Context, accToken *AccessToken, q string, org string, page, size int) (*github.RepositoriesSearchResult, error)
	ListBranches(ctx context.Context, accToken *AccessToken, repoUrl string, page, size int) ([]*github.Branch, error)
	AddWebhook(ctx context.Context, accToken *AccessToken, refId string, repoUrl string) error
	GetAppToken()
	GetRepoToken()
}

type Gitlab interface {
	Callback(ctx context.Context, code, state string) (*gitlab.User, *oauth2.Token, error)
	ListGroups(ctx context.Context, token *AccessToken, query string, page int, size int) ([]*gitlab.Group, error)
	ListRepos(ctx context.Context, token *AccessToken, gid string, query string) ([]*gitlab.Project, error)
	ListBranches(ctx context.Context, token *AccessToken, projectId string, page int, size int) ([]*gitlab.Branch, error)
	AddWebhook(ctx context.Context, token *AccessToken) (*gitlab.ProjectHook, error)
	RemoveWebhook(ctx context.Context, token *AccessToken) error
	RepoToken(ctx context.Context, token *AccessToken) (*oauth2.Token, error)
}
