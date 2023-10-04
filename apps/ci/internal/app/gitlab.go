package app

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"kloudlite.io/apps/ci/internal/domain"
	fn "kloudlite.io/pkg/functions"
	"kloudlite.io/pkg/types"

	"github.com/xanzy/go-gitlab"
	"golang.org/x/oauth2"
	oauthGitlab "golang.org/x/oauth2/gitlab"
	"kloudlite.io/pkg/errors"
)

type gitlabI struct {
	cfg        *oauth2.Config
	webhookUrl string
	env        *Env
}

func (gl *gitlabI) GetTriggerWebhookUrl() string {
	return gl.webhookUrl
}

func (gl *gitlabI) getToken(_ context.Context, token *domain.AccessToken) (*oauth2.Token, error) {
	if token == nil {
		return nil, errors.New("token is nil")
	}
	return token.Token, nil
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

func (gl *gitlabI) ListGroups(ctx context.Context, token *domain.AccessToken, query *string, pagination *types.Pagination) ([]domain.GitlabGroup, error) {
	client, err := gl.getClient(ctx, token)
	if err != nil {
		return nil, err
	}
	groups, _, err := client.Groups.ListGroups(
		&gitlab.ListGroupsOptions{
			ListOptions:          buildListOptions(pagination),
			Search:               query,
			TopLevelOnly:         fn.New(true),
			WithCustomAttributes: nil,
		},
	)
	if err != nil {
		return nil, err
	}

	grs := make([]domain.GitlabGroup, 0, len(groups)+1)

	user, _, err := client.Users.CurrentUser()
	if err != nil {
		return nil, err
	}

	grs = append(grs, domain.GitlabGroup{Id: fmt.Sprintf("%d", user.ID), FullName: user.Name, AvatarUrl: user.AvatarURL})
	for i := range groups {
		grs = append(
			grs, domain.GitlabGroup{
				Id:        fmt.Sprintf("%d", groups[i].ID),
				FullName:  groups[i].FullName,
				AvatarUrl: groups[i].AvatarURL,
			},
		)
	}

	return grs, nil
}

func buildListOptions(p *types.Pagination) gitlab.ListOptions {
	if p == nil {
		return gitlab.ListOptions{}
	}
	return gitlab.ListOptions{
		Page:    p.Page,
		PerPage: p.PerPage,
	}
}

func (gl *gitlabI) ListRepos(ctx context.Context, token *domain.AccessToken, gid string, query *string, pagination *types.Pagination) ([]*gitlab.Project, error) {
	client, err := gl.getClient(ctx, token)
	if err != nil {
		return nil, err
	}

	user, _, err := client.Users.CurrentUser()
	if err != nil {
		return nil, errors.NewEf(err, "could not get current gitlab user")
	}

	if fmt.Sprintf("%d", user.ID) == gid {
		projects, _, err := client.Projects.ListUserProjects(
			user.ID, &gitlab.ListProjectsOptions{
				ListOptions: buildListOptions(pagination),
				Search:      query,
				Simple:      fn.New(true),
			},
		)
		return projects, err
	}

	projects, _, err := client.Groups.ListGroupProjects(
		gid, &gitlab.ListGroupProjectsOptions{
			IncludeSubGroups: fn.New(true),
			ListOptions:      buildListOptions(pagination),
			Search:           query,
			Simple:           fn.New(true),
		},
	)
	return projects, err
}

func (gl *gitlabI) ListBranches(ctx context.Context, token *domain.AccessToken, repoId string, query *string, pagination *types.Pagination) ([]*gitlab.Branch, error) {
	client, err := gl.getClient(ctx, token)
	if err != nil {
		return nil, err
	}
	branches, _, err := client.Branches.ListBranches(
		repoId, &gitlab.ListBranchesOptions{
			ListOptions: buildListOptions(pagination),
			Search:      query,
		},
	)

	if err != nil {
		return nil, errors.NewEf(err, "could not list branches")
	}

	return branches, nil
}

func (gl *gitlabI) getRepoId(repoUrl string) string {
	re := regexp.MustCompile("https://(.*?)/(.*)")
	// re := regexp.MustCompile("https://(.*?)/(.*)(.git)?")
	matches := re.FindStringSubmatch(repoUrl)
	return strings.Split(matches[2], ".git")[0]
}

func (gl *gitlabI) GetRepoId(repoUrl string) string {
	return gl.getRepoId(repoUrl)
}

func (gl *gitlabI) CheckWebhookExists(ctx context.Context, token *domain.AccessToken, repoId string, webhookId *domain.GitlabWebhookId) (bool, error) {
	if webhookId == nil {
		return false, nil
	}
	client, err := gl.getClient(ctx, token)
	if err != nil {
		return false, err
	}

	hook, _, err := client.Projects.GetProjectHook(repoId, int(*webhookId))
	if err != nil {
		return false, nil
	}
	return hook != nil, nil
}

func (gl *gitlabI) AddWebhook(ctx context.Context, token *domain.AccessToken, repoId string, pipelineId string) (*domain.GitlabWebhookId, error) {
	client, err := gl.getClient(ctx, token)
	if err != nil {
		return nil, err
	}
	// webhookUrl := fmt.Sprintf("%s?pipelineId=%s", gl.webhookUrl, pipelineId)

	hook, _, err := client.Projects.AddProjectHook(
		repoId, &gitlab.AddProjectHookOptions{
			PushEvents:    fn.New(true),
			TagPushEvents: fn.New(true),
			Token:         &gl.env.GitlabWebhookAuthzSecret,
			URL:           &gl.webhookUrl,
		},
	)
	if err != nil {
		return nil, errors.NewEf(err, "could not add gitlab webhook")
	}
	return fn.New(domain.GitlabWebhookId(hook.ID)), nil
}

func (gl *gitlabI) DeleteWebhook(ctx context.Context, token *domain.AccessToken, repoUrl string, hookId domain.GitlabWebhookId) error {
	client, err := gl.getClient(ctx, token)
	if err != nil {
		return err
	}
	_, err = client.Projects.DeleteProjectHook(gl.getRepoId(repoUrl), int(hookId))
	return err
}

func (gl *gitlabI) GetLatestCommit(ctx context.Context, token *domain.AccessToken, repoUrl string, branchName string) (string, error) {
	client, err := gl.getClient(ctx, token)
	if err != nil {
		return "", err
	}
	repoId := gl.getRepoId(repoUrl)
	branch, _, err := client.Branches.GetBranch(repoId, branchName)
	if err != nil {
		return "", err
	}
	return branch.Commit.ShortID, nil
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
		Scopes:       strings.Split(env.GitlabScopes, ","),
	}

	return &gitlabI{cfg: &cfg, webhookUrl: env.GitlabWebhookUrl, env: env}
}
