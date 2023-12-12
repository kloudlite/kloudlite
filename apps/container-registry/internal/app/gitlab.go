package app

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"github.com/kloudlite/api/apps/container-registry/internal/domain"
	"github.com/kloudlite/api/apps/container-registry/internal/domain/entities"
	"github.com/kloudlite/api/pkg/errors"
	fn "github.com/kloudlite/api/pkg/functions"
	"github.com/kloudlite/api/pkg/types"
	gitlab "github.com/xanzy/go-gitlab"
	"go.uber.org/fx"
	"golang.org/x/oauth2"
	oauthGitlab "golang.org/x/oauth2/gitlab"
)

type gitlabOptions interface {
	GitlabConfig() (clientId, clientSecret, callbackUrl string)
	GitlabScopes() string
	GitlabWebhookAuthzSecret() *string
	GitlabWebhookUrl() *string
}

type gitlabI struct {
	cfg *oauth2.Config
	// webhookUrl string
	env gitlabOptions
}

// AddWebhook implements domain.Gitlab.
func (gl *gitlabI) AddWebhook(ctx context.Context, token *entities.AccessToken, repoId string) (*int, error) {
	client, err := gl.getClient(ctx, token)
	if err != nil {
		return nil, err
	}
	// webhookUrl := fmt.Sprintf("%s?pipelineId=%s", gl.webhookUrl, pipelineId)

	hook, _, err := client.Projects.AddProjectHook(
		repoId, &gitlab.AddProjectHookOptions{
			PushEvents:    fn.New(true),
			TagPushEvents: fn.New(true),
			Token:         gl.env.GitlabWebhookAuthzSecret(),
			URL:           gl.env.GitlabWebhookUrl(),
		},
	)
	if err != nil {
		return nil, errors.NewEf(err, "could not add gitlab webhook")
	}
	return &hook.ID, nil
}

// Callback implements domain.Gitlab.
func (*gitlabI) Callback(ctx context.Context, code string, state string) (*gitlab.User, *oauth2.Token, error) {
	panic("unimplemented")
}

// CheckWebhookExists implements domain.Gitlab.
func (*gitlabI) CheckWebhookExists(ctx context.Context, token *entities.AccessToken, repoId string, webhookId *entities.GitlabWebhookId) (bool, error) {
	panic("unimplemented")
}

func (gl *gitlabI) getRepoId(repoUrl string) string {
	re := regexp.MustCompile("https://(.*?)/(.*)")
	// re := regexp.MustCompile("https://(.*?)/(.*)(.git)?")
	matches := re.FindStringSubmatch(repoUrl)
	return strings.Split(matches[2], ".git")[0]
}

// DeleteWebhook implements domain.Gitlab.
func (gl *gitlabI) DeleteWebhook(ctx context.Context, token *entities.AccessToken, repoUrl string, hookId entities.GitlabWebhookId) error {

	client, err := gl.getClient(ctx, token)
	if err != nil {
		return err
	}
	_, err = client.Projects.DeleteProjectHook(gl.getRepoId(repoUrl), int(hookId))
	return err
}

// GetLatestCommit implements domain.Gitlab.
func (*gitlabI) GetLatestCommit(ctx context.Context, token *entities.AccessToken, repoUrl string, branchName string) (string, error) {
	panic("unimplemented")
}

// GetRepoId implements domain.Gitlab.
func (*gitlabI) GetRepoId(repoUrl string) string {
	re := regexp.MustCompile("https://(.*?)/(.*)")
	// re := regexp.MustCompile("https://(.*?)/(.*)(.git)?")
	matches := re.FindStringSubmatch(repoUrl)
	return strings.Split(matches[2], ".git")[0]

}

// GetTriggerWebhookUrl implements domain.Gitlab.
func (*gitlabI) GetTriggerWebhookUrl() string {
	panic("unimplemented")
}

// ListBranches implements domain.Gitlab.
func (gl *gitlabI) ListBranches(ctx context.Context, token *entities.AccessToken, repoId string, query *string, pagination *types.Pagination) ([]*gitlab.Branch, error) {

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

	return branches, err
}

func (gl *gitlabI) getToken(_ context.Context, token *entities.AccessToken) (*oauth2.Token, error) {
	if token == nil {
		return nil, errors.New("token is nil")
	}
	return token.Token, nil
}

func (gl *gitlabI) getClient(ctx context.Context, token *entities.AccessToken) (*gitlab.Client, error) {
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

func buildListOptions(p *types.Pagination) gitlab.ListOptions {
	if p == nil {
		return gitlab.ListOptions{}
	}
	return gitlab.ListOptions{
		Page:    p.Page,
		PerPage: p.PerPage,
	}
}

// ListGroups implements domain.Gitlab.
func (gl *gitlabI) ListGroups(ctx context.Context, token *entities.AccessToken, query *string, pagination *types.Pagination) ([]*entities.GitlabGroup, error) {
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

	grs := make([]*entities.GitlabGroup, 0, len(groups)+1)

	user, _, err := client.Users.CurrentUser()
	if err != nil {
		return nil, err
	}

	grs = append(grs, &entities.GitlabGroup{Id: fmt.Sprintf("%d", user.ID), FullName: user.Name, AvatarUrl: user.AvatarURL})
	for i := range groups {
		grs = append(
			grs, &entities.GitlabGroup{
				Id:        fmt.Sprintf("%d", groups[i].ID),
				FullName:  groups[i].FullName,
				AvatarUrl: groups[i].AvatarURL,
			},
		)
	}

	return grs, nil
}

// ListRepos implements domain.Gitlab.
func (gl *gitlabI) ListRepos(ctx context.Context, token *entities.AccessToken, gid string, query *string, pagination *types.Pagination) ([]*gitlab.Project, error) {
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

// RepoToken implements domain.Gitlab.
func (*gitlabI) RepoToken(ctx context.Context, token *entities.AccessToken) (*oauth2.Token, error) {
	panic("unimplemented")
}

func fxGitlab[T gitlabOptions]() fx.Option {
	return fx.Module("gitlab-fx", fx.Provide(
		func(env T) (domain.Gitlab, error) {

			clientId, clientSecret, callbackUrl := env.GitlabConfig()
			cfg := oauth2.Config{
				ClientID:     clientId,
				ClientSecret: clientSecret,
				Endpoint:     oauthGitlab.Endpoint,
				RedirectURL:  callbackUrl,
				Scopes:       strings.Split(env.GitlabScopes(), ","),
			}

			return &gitlabI{
				env: env,
				cfg: &cfg,
			}, nil
		},
	))
}
