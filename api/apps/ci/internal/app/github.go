package app

import (
	"context"
	"fmt"
	"io/ioutil"
	"kloudlite.io/apps/ci/internal/domain"
	"kloudlite.io/pkg/types"
	"net/http"
	"regexp"
	"strconv"
	"time"

	"github.com/bradleyfalzon/ghinstallation/v2"
	"github.com/google/go-github/v43/github"
	"golang.org/x/oauth2"
	oauthGithub "golang.org/x/oauth2/github"
	"kloudlite.io/pkg/errors"
	fn "kloudlite.io/pkg/functions"
)

type githubI struct {
	cfg          *oauth2.Config
	ghCli        *github.Client
	ghCliForUser func(ctx context.Context, token *oauth2.Token) *github.Client
	webhookUrl   string
}

func (gh *githubI) getOwnerAndRepo(repoUrl string) (owner, repo string) {
	re := regexp.MustCompile("https://(.*?)/(.*)/(.*?)([.]git)?")
	matches := re.FindStringSubmatch(repoUrl)
	return matches[2], matches[3]
}

func (gh *githubI) buildListOptions(p *types.Pagination) github.ListOptions {
	if p == nil {
		return github.ListOptions{}
	}
	return github.ListOptions{
		Page:    p.Page,
		PerPage: p.PerPage,
	}
}

func (gh *githubI) ListBranches(ctx context.Context, accToken *domain.AccessToken, repoUrl string, pagination *types.Pagination) ([]*github.Branch, error) {
	owner, repo := gh.getOwnerAndRepo(repoUrl)
	branches, _, err := gh.ghCliForUser(ctx, accToken.Token).Repositories.ListBranches(
		ctx, owner, repo, &github.BranchListOptions{
			ListOptions: gh.buildListOptions(pagination),
		},
	)
	if err != nil {
		return nil, errors.NewEf(err, "could not list branches")
	}
	return branches, nil
}

func (gh *githubI) SearchRepos(ctx context.Context, accToken *domain.AccessToken, q, org string, pagination *types.Pagination) (*github.RepositoriesSearchResult, error) {
	rsr, _, err := gh.ghCliForUser(ctx, accToken.Token).Search.Repositories(
		ctx, fmt.Sprintf("%s org:%s", q, org), &github.SearchOptions{
			ListOptions: gh.buildListOptions(pagination),
		},
	)
	if err != nil {
		return nil, errors.NewEf(err, "could not search repositories")
	}
	return rsr, nil
}

func (gh *githubI) ListInstallations(ctx context.Context, accToken *domain.AccessToken, pagination *types.Pagination) ([]*github.Installation, error) {
	opts := gh.buildListOptions(pagination)
	i, _, err := gh.ghCliForUser(ctx, accToken.Token).Apps.ListUserInstallations(ctx, &opts)
	if err != nil {
		return nil, errors.NewEf(err, "could not list user installations")
	}
	return i, nil
}

func (gh *githubI) ListRepos(ctx context.Context, accToken *domain.AccessToken, instId int64, pagination *types.Pagination) (*github.ListRepositories, error) {
	opts := gh.buildListOptions(pagination)
	fmt.Println("opts: ", opts)
	repos, _, err := gh.ghCliForUser(ctx, accToken.Token).Apps.ListUserRepos(ctx, instId, &opts)
	// repos, _, err := gh.ghCli.Apps.ListUserRepos(ctx, instId, &opts)
	if err != nil {
		return nil, errors.NewEf(err, "could not list user repositories")
	}
	return repos, nil
}

func (gh *githubI) GetLatestCommit(ctx context.Context, repoUrl string, branchName string) (string, error) {
	owner, repo := gh.getOwnerAndRepo(repoUrl)
	branch, _, err := gh.ghCli.Repositories.GetBranch(ctx, owner, repo, branchName, true)
	if err != nil {
		return "", err
	}
	return *branch.GetCommit().SHA, nil
}

func (gh *githubI) AddWebhook(ctx context.Context, accToken *domain.AccessToken, pipelineId string, repoUrl string) error {
	owner, repo := gh.getOwnerAndRepo(repoUrl)
	hookUrl := fmt.Sprintf("%s?pipelineId=%s", gh.webhookUrl, pipelineId)
	hookName := "kloudlite-pipeline"

	hook, res, err := gh.ghCliForUser(ctx, accToken.Token).Repositories.CreateHook(
		ctx, owner, repo, &github.Hook{
			Config: map[string]interface{}{
				"url":          hookUrl,
				"content_type": "json",
			},
			Events: []string{"push"},
			Active: fn.NewBool(true),
			Name:   &hookName,
		},
	)
	if err != nil {
		fmt.Println(res.Status, res.Body)
		// ASSERT: github returns 422 only if hook already exists on the repository
		if res.StatusCode == 422 {
			fmt.Printf("Hook: %+v\n", hook)
			return nil
		}
		return errors.NewEf(err, "could not create github webhook")
	}
	fmt.Printf("Hook: %+v\n", hook)

	return nil
}

func (gh *githubI) Callback(ctx context.Context, code, state string) (*github.User, *oauth2.Token, error) {
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

func (gh *githubI) GetToken(ctx context.Context, token *oauth2.Token) (*oauth2.Token, error) {
	return gh.cfg.TokenSource(ctx, token).Token()
}

func (gh *githubI) GetInstallationToken(ctx context.Context, repoUrl string) (string, error) {
	owner, repo := gh.getOwnerAndRepo(repoUrl)
	inst, _, err := gh.ghCli.Apps.FindRepositoryInstallation(ctx, owner, repo)
	if err != nil {
		return "", errors.NewEf(err, "could not fetch repository installation")
	}
	installationId := *inst.ID
	it, _, err := gh.ghCli.Apps.CreateInstallationToken(ctx, installationId, &github.InstallationTokenOptions{})
	fmt.Println(it)
	if err != nil {
		return "", errors.NewEf(err, "failed to get installation token")
	}
	return it.GetToken(), err
}

type GithubOAuth interface {
	GithubConfig() (clientId, clientSecret, callbackUrl, githubAppId, githubAppPKFile string)
}

func fxGithub(env *Env) domain.Github {
	clientId, clientSecret, callbackUrl, ghAppId, ghAppPKFile := env.GithubConfig()
	cfg := oauth2.Config{
		ClientID:     clientId,
		ClientSecret: clientSecret,
		Endpoint:     oauthGithub.Endpoint,
		RedirectURL:  callbackUrl,
		Scopes:       []string{"user:email", "admin:org"},
	}
	privatePem, err := ioutil.ReadFile(ghAppPKFile)
	if err != nil {
		panic(errors.NewEf(err, "reading github app PK file"))
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

	ghCli := github.NewClient(&http.Client{Transport: itr, Timeout: time.Second * 30})

	return &githubI{
		cfg:          &cfg,
		ghCli:        ghCli,
		ghCliForUser: ghCliForUser,
		webhookUrl:   env.GithubWebhookUrl,
	}
}
