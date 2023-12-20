package app

import (
	"context"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/bradleyfalzon/ghinstallation/v2"
	"github.com/google/go-github/v43/github"
	"github.com/kloudlite/api/apps/auth/internal/domain"
	"github.com/kloudlite/api/apps/auth/internal/env"
	"github.com/kloudlite/api/pkg/errors"
	fn "github.com/kloudlite/api/pkg/functions"
	"golang.org/x/oauth2"
	oauthGithub "golang.org/x/oauth2/github"
)

type githubI struct {
	enabled      bool
	cfg          *oauth2.Config
	ghCli        *github.Client
	ghCliForUser func(ctx context.Context, token *oauth2.Token) *github.Client
	webhookUrl   string
}

func (gh *githubI) GetOAuthToken(ctx context.Context, token *oauth2.Token) (*oauth2.Token, error) {
	if !gh.enabled {
		return nil, errors.Newf("github oauth is disabled")
	}
	return gh.cfg.TokenSource(ctx, token).Token()
}

func (gh *githubI) Authorize(_ context.Context, state string) (string, error) {
	if !gh.enabled {
		return "", errors.Newf("github oauth is disabled")
	}
	csrfToken := fn.Must(fn.CleanerNanoid(32))
	b64state, err := fn.ToBase64UrlFromJson(map[string]string{"csrf": csrfToken, "state": state})
	if err != nil {
		return "", errors.NewEf(err, "could not JSON marshal oauth State into []byte")
	}
	return gh.cfg.AuthCodeURL(
		b64state,
		oauth2.SetAuthURLParam("allow_signup", strconv.FormatBool(true)),
	), nil
}

func (gh *githubI) Callback(ctx context.Context, code, state string) (*github.User, *oauth2.Token, error) {
	if !gh.enabled {
		return nil, nil, errors.Newf("github oauth is disabled")
	}

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

func (gh *githubI) GetPrimaryEmail(ctx context.Context, token *oauth2.Token) (string, error) {
	if !gh.enabled {
		return "", errors.Newf("github oauth is disabled")
	}

	emails, _, err := gh.ghCliForUser(ctx, token).Users.ListEmails(
		ctx, &github.ListOptions{
			Page:    1,
			PerPage: 100, // max per page, assuming user would not have 100 emails added to his github
		},
	)
	if err != nil {
		return "", err
	}

	for i := range emails {
		if emails[i].GetPrimary() {
			return emails[i].GetEmail(), nil
		}
	}

	return "", errors.Newf("no primary email could be found for this user, among first 100 emails provided by github")
}

func fxGithub(ev *env.Env) domain.Github {
	if !ev.OAuth2Enabled || !ev.OAuth2GithubEnabled {
		return &githubI{enabled: false}
	}

	cfg := oauth2.Config{
		ClientID:     ev.GithubClientId,
		ClientSecret: ev.GithubClientSecret,
		Endpoint:     oauthGithub.Endpoint,
		RedirectURL:  ev.GithubCallbackUrl,
		Scopes:       strings.Split(ev.GithubScopes, ","),
	}
	privatePem, err := os.ReadFile(ev.GithubAppPKFile)
	if err != nil {
		panic(errors.NewEf(err, "reading github app PK file"))
	}

	appId, _ := strconv.ParseInt(ev.GithubAppId, 10, 64)
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
		enabled:      true,
		cfg:          &cfg,
		ghCli:        ghCli,
		ghCliForUser: ghCliForUser,
		webhookUrl:   ev.GithubWebhookUrl,
	}
}
