package app

import (
	"context"
	"fmt"
	// "io/ioutil"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/bradleyfalzon/ghinstallation/v2"
	"github.com/google/go-github/v45/github"
	"go.uber.org/fx"
	"golang.org/x/oauth2"
	oauthGithub "golang.org/x/oauth2/github"
	"kloudlite.io/apps/container-registry/internal/domain"
	"kloudlite.io/apps/container-registry/internal/domain/entities"

	"kloudlite.io/pkg/errors"
	// fn "kloudlite.io/pkg/functions"
	"kloudlite.io/pkg/types"
)

type githubOptions interface {
	GithubConfig() (clientId, clientSecret, callbackUrl, ghAppId, ghAppPKFile string)
	GithubScopes() string
	GithubWebhookAuthzSecret() string
}

type githubI struct {
	cfg          *oauth2.Config
	ghCli        *github.Client
	ghCliForUser func(ctx context.Context, token *oauth2.Token) *github.Client
	env          githubOptions
}

type GithubOAuth interface {
	GithubConfig() (clientId, clientSecret, callbackUrl, githubAppId, githubAppPKFile string)
}

func (gh *githubI) getOwnerAndRepo(repoUrl string) (owner, repo string, err error) {
	re := regexp.MustCompile("https://(.+)/(.+)/(.+)")

	if !re.MatchString(repoUrl) {
		return "", "", fmt.Errorf("invalid repository url")
	}

	matches := re.FindStringSubmatch(repoUrl)
	return matches[2], strings.Split(matches[3], ".git")[0], nil
}

func (gh *githubI) buildListOptions(p *types.Pagination) github.ListOptions {
	if p == nil {
		return github.ListOptions{
			PerPage: 200,
		}
	}
	return github.ListOptions{
		Page:    p.Page,
		PerPage: p.PerPage,
	}
}

// maybe not needed
func (gh *githubI) AddRepoWebhook(ctx context.Context, accToken *entities.AccessToken, repoUrl, webhookUrl string) (*entities.GithubWebhookId, error) {
	// owner, repo := gh.getOwnerAndRepo(repoUrl)
	// hookName := "kloudlite-pipeline"
	//
	// hook, res, err := gh.ghCliForUser(ctx, accToken.Token).Repositories.CreateHook(
	// 	ctx, owner, repo, &github.Hook{
	// 		Config: map[string]any{
	// 			"url":          webhookUrl,
	// 			"content_type": "json",
	// 			"secret":       gh.env.GithubWebhookAuthzSecret,
	// 		},
	// 		Events: []string{"push"},
	// 		Active: fn.NewBool(true),
	// 		Name:   &hookName,
	// 	},
	// )
	// if err != nil {
	// 	// ASSERT: GitHub returns 422 only if hook already exists on the repository
	// 	if res.StatusCode == 422 {
	// 		return nil, nil
	// 	}
	// 	return nil, errors.NewEf(err, "could not create github webhook")
	// }
	//
	// return fn.New(entities.GithubWebhookId(*hook.ID)), nil

	return nil, fmt.Errorf("not implemented")
}

func (gh *githubI) AddWebhook(ctx context.Context, accToken *entities.AccessToken, repoUrl string, webhookUrl string) (*entities.GithubWebhookId, error) {
	// TODO:: we migrated to GitHub app webhook, which allows us to skip creating github repository webhooks, now
	return nil, fmt.Errorf("not implemented")
}

// Callback implements domain.Github.
func (gh *githubI) Callback(ctx context.Context, code string, state string) (*github.User, *oauth2.Token, error) {
	token, err := gh.cfg.Exchange(ctx, code)
	if err != nil {
		return nil, nil, errors.NewEf(err, "could not exchange the token")
	}
	c := gh.cfg.Client(ctx, token)
	c2 := github.NewClient(c)
	u, _, err := c2.Users.Get(ctx, "")
	if err != nil {
		return nil, nil, errors.NewEf(err, "could nog get authenticated user from github")
	}
	return u, token, nil

}

// CheckWebhookExists implements domain.Github.
func (gh *githubI) CheckWebhookExists(ctx context.Context, token *entities.AccessToken, repoUrl string, webhookId *entities.GithubWebhookId) (bool, error) {
	if webhookId == nil {
		return false, nil
	}

	owner, repo, err := gh.getOwnerAndRepo(repoUrl)
	if err != nil {
		return false, err
	}

	hook, _, err := gh.ghCliForUser(ctx, token.Token).Repositories.GetHook(ctx, owner, repo, int64(*webhookId))
	if err != nil {
		return false, err
	}

	return hook != nil, nil
}

// DeleteWebhook implements domain.Github.
func (gh *githubI) DeleteWebhook(ctx context.Context, accToken *entities.AccessToken, repoUrl string, hookId entities.GithubWebhookId) error {

	owner, repo, err := gh.getOwnerAndRepo(repoUrl)

	if err != nil {
		return err
	}

	resp, err := gh.ghCliForUser(ctx, accToken.Token).Repositories.DeleteHook(ctx, owner, repo, int64(hookId))
	if err != nil && resp.StatusCode == http.StatusNotFound {
		return nil
	}
	return err

}

// GetInstallationToken implements domain.Github.
func (gh *githubI) GetInstallationToken(ctx context.Context, repoUrl string) (string, error) {
	owner, repo, err := gh.getOwnerAndRepo(repoUrl)

	if err != nil {
		return "", err
	}

	inst, _, err := gh.ghCli.Apps.FindRepositoryInstallation(ctx, owner, repo)
	if err != nil {
		return "", errors.NewEf(err, "could not fetch repository installation")
	}
	installationId := *inst.ID
	it, _, err := gh.ghCli.Apps.CreateInstallationToken(ctx, installationId, &github.InstallationTokenOptions{})
	if err != nil {
		return "", errors.NewEf(err, "failed to get installation token")
	}
	return it.GetToken(), err

}

// GetLatestCommit implements domain.Github.
func (gh *githubI) GetLatestCommit(ctx context.Context, accToken *entities.AccessToken, repoUrl string, branchName string) (string, error) {
	owner, repo, err := gh.getOwnerAndRepo(repoUrl)
	if err != nil {
		return "", err
	}

	inst, _, err := gh.ghCli.Apps.FindRepositoryInstallation(ctx, owner, repo)
	if err != nil {
		return "", errors.NewEf(err, "could not fetch repository installation")
	}
	installationId := *inst.ID
	it, _, err := gh.ghCli.Apps.CreateInstallationToken(ctx, installationId, &github.InstallationTokenOptions{})
	if err != nil {
		return "", errors.NewEf(err, "failed to get installation token")
	}
	return it.GetToken(), err
}

// GetToken implements domain.Github.
func (gh *githubI) GetToken(ctx context.Context, token *oauth2.Token) (*oauth2.Token, error) {
	return gh.cfg.TokenSource(ctx, token).Token()
}

// ListBranches implements domain.Github.
func (gh *githubI) ListBranches(ctx context.Context, accToken *entities.AccessToken, repoUrl string, pagination *types.Pagination) ([]*github.Branch, error) {
	owner, repo, err := gh.getOwnerAndRepo(repoUrl)
	if err != nil {
		return nil, err
	}

	var branches []*github.Branch
	hasFetchedAll := false
	pageNo := 1
	for !hasFetchedAll {
		if pageNo > 5 {
			break
		}
		brcs, _, err := gh.ghCliForUser(ctx, accToken.Token).Repositories.ListBranches(
			ctx, owner, repo, &github.BranchListOptions{
				ListOptions: func() github.ListOptions {
					return github.ListOptions{Page: pageNo, PerPage: 100}
				}(),
			},
		)
		if err != nil {
			return nil, errors.NewEf(err, "could not list branches")
		}
		branches = append(branches, brcs...)
		if len(brcs) != 100 {
			hasFetchedAll = true
		}
		pageNo++
	}
	return branches, nil
}

// ListInstallations implements domain.Github.
func (gh *githubI) ListInstallations(ctx context.Context, accToken *entities.AccessToken, pagination *types.Pagination) ([]*github.Installation, error) {
	opts := gh.buildListOptions(pagination)
	i, _, err := gh.ghCliForUser(ctx, accToken.Token).Apps.ListUserInstallations(ctx, &opts)
	if err != nil {
		return nil, errors.NewEf(err, "could not list user installations")
	}
	return i, nil
}

// ListRepos implements domain.Github.
func (gh *githubI) ListRepos(ctx context.Context, accToken *entities.AccessToken, instId int64, pagination *types.Pagination) (*github.ListRepositories, error) {
	opts := gh.buildListOptions(pagination)
	repos, _, err := gh.ghCliForUser(ctx, accToken.Token).Apps.ListUserRepos(ctx, instId, &opts)
	if err != nil {
		return nil, errors.NewEf(err, "could not list user repositories")
	}
	return repos, nil

}

// SearchRepos implements domain.Github.
func (gh *githubI) SearchRepos(ctx context.Context, accToken *entities.AccessToken, q string, org string, pagination *types.Pagination) (*github.RepositoriesSearchResult, error) {
	// TODO: search repos not working at all from the API
	searchQ := fmt.Sprintf("%s org:%s", q, org)
	rsr, resp, err := gh.ghCliForUser(ctx, accToken.Token).Search.Repositories(
		context.TODO(), searchQ, &github.SearchOptions{
			ListOptions: gh.buildListOptions(pagination),
		},
	)
	fmt.Println(resp)
	if err != nil {
		return nil, errors.NewEf(err, "could not search repositories")
	}
	return rsr, nil
}

func fxGithub[T githubOptions]() fx.Option {

	return fx.Module("github-fx",
		fx.Provide(
			func(env T) (domain.Github, error) {
				clientId, clientSecret, callbackUrl, ghAppId, ghAppPKFile := env.GithubConfig()
				cfg := oauth2.Config{
					ClientID:     clientId,
					ClientSecret: clientSecret,
					Endpoint:     oauthGithub.Endpoint,
					RedirectURL:  callbackUrl,
					Scopes:       strings.Split(env.GithubScopes(), ","),
					// Scopes:       []string{"user:email", "admin:org"},
				}

				if _, err := os.Stat(ghAppPKFile); err != nil {
					return nil, fmt.Errorf("github privaate key file (path='%s') does not exist", ghAppPKFile)
				}

				// ioutil.ReadFile(name string)
				privatePem, err := os.ReadFile(ghAppPKFile)
				if err != nil {
					return nil, errors.NewEf(err, "reading github app PK file")
				}

				appId, _ := strconv.ParseInt(ghAppId, 10, 64)
				itr, err := ghinstallation.NewAppsTransport(http.DefaultTransport, appId, privatePem)
				if err != nil {
					panic(errors.NewEf(err, "creating app transport"))
				}

				ghCliForUser := func(ctx context.Context, token *oauth2.Token) *github.Client {
					ts := oauth2.StaticTokenSource(token)
					return github.NewClient(oauth2.NewClient(ctx, ts))
				}

				ghCli := github.NewClient(&http.Client{Transport: itr, Timeout: 30 * time.Second})

				return &githubI{
					cfg:          &cfg,
					ghCli:        ghCli,
					ghCliForUser: ghCliForUser,
					env:          env,
				}, nil
			}))
}
