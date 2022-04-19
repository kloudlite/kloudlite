package app

import (
	"context"
	"strconv"

	"github.com/google/go-github/v43/github"
	"golang.org/x/oauth2"
	oauthGithub "golang.org/x/oauth2/github"
	"kloudlite.io/apps/auth/internal/domain"
	"kloudlite.io/pkg/cache"
	"kloudlite.io/pkg/errors"
	fn "kloudlite.io/pkg/functions"
	"kloudlite.io/pkg/repos"
)

type githubI struct {
	cfg             *oauth2.Config
	githubAppId     string
	githubAppPKFile string
	accTokenRepo    repos.DbRepo[*domain.AccessToken]
}

func (gh *githubI) Authorize(_ context.Context, state string) (string, error) {
	csrfToken := fn.Must(fn.CleanerNanoid(32))
	b64state, err := fn.Json.ToB64Url(map[string]string{"csrf": csrfToken, "state": state})
	if err != nil {
		return "", errors.NewEf(err, "could not JSON marshal oauth State into []byte")
	}
	return gh.cfg.AuthCodeURL(
		b64state,
		oauth2.SetAuthURLParam("allow_signup", strconv.FormatBool(true)),
	), nil
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

func (gh *githubI) getAccessToken(ctx context.Context, provider string) (*domain.AccessToken, error) {
	session := cache.GetSession[*domain.AccessToken](ctx)
	if session == nil {
		return nil, errors.UnAuthorized()
	}
	accToken, err := gh.accTokenRepo.FindOne(ctx, repos.Filter{"user_id": session.UserId, "provider": provider})
	if err != nil {
		return nil, errors.NewEf(err, "finding access token")
	}
	return accToken, nil
}

func (gh *githubI) InstallationToken(ctx context.Context, accToken *domain.AccessToken, installationId int64) (string, error) {
	// accToken, err := gh.getAccessToken(ctx, common.ProviderGithub)
	c := gh.cfg.Client(ctx, accToken.Token)
	c2 := github.NewClient(c)
	it, _, err := c2.Apps.CreateInstallationToken(ctx, installationId, &github.InstallationTokenOptions{})
	if err != nil {
		return "", errors.NewEf(err, "failed to get installation token")
	}
	return it.GetToken(), err
}

func (gh *githubI) GetAppToken() {
}

func (gh *githubI) GetRepoToken() {
	panic("not implemented") // TODO: Implement
}

type GithubOAuth interface {
	GithubConfig() (clientId, clientSecret, callbackUrl, githubAppId, githubAppPKFile string)
}

func fxGithub(env *Env, accTokenRepo repos.DbRepo[*domain.AccessToken]) domain.Github {
	clientId, clientSecret, callbackUrl, ghAppId, ghAppPKFile := env.GithubConfig()
	cfg := oauth2.Config{
		ClientID:     clientId,
		ClientSecret: clientSecret,
		Endpoint:     oauthGithub.Endpoint,
		RedirectURL:  callbackUrl,
		Scopes:       []string{"user:email", "admin:org"},
	}
	return &githubI{
		cfg:             &cfg,
		githubAppId:     ghAppId,
		githubAppPKFile: ghAppPKFile,
		accTokenRepo:    accTokenRepo,
	}
}
