package app

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"

	"github.com/bradleyfalzon/ghinstallation/v2"
	"github.com/google/go-github/v43/github"
	"golang.org/x/oauth2"
	oauthGithub "golang.org/x/oauth2/github"
	"kloudlite.io/apps/auth/internal/domain"
	"kloudlite.io/pkg/errors"
	fn "kloudlite.io/pkg/functions"
)

type githubI struct {
	cfg         *oauth2.Config
	githubAppId string
	githubAppPK []byte
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

func (gh *githubI) InstallationToken(ctx context.Context, accToken *domain.AccessToken, repoUrl string) (string, error) {
	fmt.Println(accToken)
	c := gh.cfg.Client(ctx, accToken.Token)
	c2 := github.NewClient(c)

	inst, _, err := c2.Apps.FindRepositoryInstallation(ctx, "nxtcoder17", "sample")
	if err != nil {
		return "", errors.NewEf(err, "could not fetch repository installation")
	}
	it, _, err := c2.Apps.CreateInstallationToken(ctx, *inst.ID, &github.InstallationTokenOptions{})
	fmt.Println(it)
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
	ghinstallation.NewAppsTransport(http.DefaultTransport, 10, privatePem)

	if err != nil {
		panic(errors.NewEf(err, "could not read github app PK file"))
	}
	return &githubI{
		cfg:         &cfg,
		githubAppId: ghAppId,
		githubAppPK: privatePem,
	}
}
