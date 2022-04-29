package app

import (
	"context"
	"kloudlite.io/apps/ci/internal/domain"
	fn "kloudlite.io/pkg/functions"

	"github.com/xanzy/go-gitlab"
	"golang.org/x/oauth2"
	oauthGitlab "golang.org/x/oauth2/gitlab"
	"kloudlite.io/pkg/errors"
)

type gitlabI struct {
	cfg        *oauth2.Config
	webhookUrl string
}

func (gl *gitlabI) getToken(ctx context.Context, token *domain.AccessToken) (*oauth2.Token, error) {
	t, err := gl.cfg.TokenSource(ctx, token.Token).Token()
	if err != nil {
		return nil, errors.NewEf(err, "could not get token from token source")
	}
	return t, nil
}

func (gl *gitlabI) getClient(ctx context.Context, token *domain.AccessToken) (*gitlab.Client, error) {
	t, err := gl.getToken(ctx, token)
	if err != nil {
		return nil, err
	}
	client, err := gitlab.NewOAuthClient(t.AccessToken)
	if err != nil {
		return nil, errors.NewEf(err, "could not build gitlab oauth client")
	}
	return client, nil
}

func (gl *gitlabI) ListGroups(ctx context.Context, token *domain.AccessToken, query string, page int, size int) ([]*gitlab.Group, error) {
	client, err := gl.getClient(ctx, token)
	groups, _, err := client.Groups.ListGroups(&gitlab.ListGroupsOptions{
		Search: &query,
		ListOptions: gitlab.ListOptions{
			Page:    page,
			PerPage: size,
		},
		TopLevelOnly: fn.NewBool(true),
	})
	if err != nil {
		return nil, nil
	}
	return groups, nil
}

func (gl *gitlabI) ListRepos(ctx context.Context, token *domain.AccessToken, gid string, query string) ([]*gitlab.Project, error) {
	client, err := gl.getClient(ctx, token)
	if err != nil {
		return nil, err
	}
	projects, _, err := client.Groups.ListGroupProjects(gid, &gitlab.ListGroupProjectsOptions{
		ListOptions:      gitlab.ListOptions{},
		IncludeSubGroups: fn.NewBool(true),
		Search:           &query,
		Simple:           fn.NewBool(true),
	})
	if err != nil {
		return nil, err
	}
	return projects, nil
}

func (gl *gitlabI) ListBranches(ctx context.Context, token *domain.AccessToken, projectId string, page int, size int) ([]*gitlab.Branch, error) {
	client, err := gl.getClient(ctx, token)
	if err != nil {
		return nil, err
	}
	branches, _, err := client.Branches.ListBranches(projectId, &gitlab.ListBranchesOptions{
		ListOptions: gitlab.ListOptions{
			Page:    page,
			PerPage: size,
		},
		Search: nil,
	})

	if err != nil {
		return nil, errors.NewEf(err, "could not list branches")
	}

	return branches, nil
}

func (gl *gitlabI) AddWebhook(ctx context.Context, token *domain.AccessToken) (*gitlab.ProjectHook, error) {
	client, err := gl.getClient(ctx, token)
	if err != nil {
		return nil, err
	}
	id, err := fn.CleanerNanoid(32)
	if err != nil {
		return nil, err
	}
	hook, _, err := client.Projects.AddProjectHook(ctx, &gitlab.AddProjectHookOptions{
		PushEvents:    fn.NewBool(true),
		TagPushEvents: fn.NewBool(true),
		Token:         &id,
		URL:           &gl.webhookUrl,
	})
	return hook, nil
}

func (gl *gitlabI) RemoveWebhook(ctx context.Context, token *domain.AccessToken) error {
	//TODO implement me
	panic("implement me")
}

func (gl *gitlabI) RepoToken(ctx context.Context, token *domain.AccessToken) (*oauth2.Token, error) {
	return gl.getToken(ctx, token)
}

func (gl *gitlabI) Callback(ctx context.Context, code string, state string) (*gitlab.User, *oauth2.Token, error) {
	token, err := gl.cfg.Exchange(ctx, code)
	if err != nil {
		return nil, nil, errors.NewEf(err, "could not exchange the token")
	}

	c2, err := gitlab.NewOAuthClient(token.AccessToken)
	if err != nil {
		return nil, nil, errors.NewEf(err, "could not create gitlab oauth client")
	}
	u, _, err := c2.Users.CurrentUser()
	if err != nil {
		return nil, nil, errors.NewEf(err, "could not get gitlab user")
	}
	return u, token, nil
}

type GitlabOAuth interface {
	GitlabConfig() (clientId, clientSecret, callbackUrl string)
}

func fxGitlab(env *Env) domain.Gitlab {
	clientId, clientSecret, callbackUrl := env.GitlabConfig()
	cfg := oauth2.Config{
		ClientID:     clientId,
		ClientSecret: clientSecret,
		Endpoint:     oauthGitlab.Endpoint,
		RedirectURL:  callbackUrl,
		Scopes:       []string{"api"},
	}

	return &gitlabI{cfg: &cfg, webhookUrl: env.GitlabWebhookUrl}
}
