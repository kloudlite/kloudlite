package app

import (
	"context"
	"strings"

	"github.com/xanzy/go-gitlab"
	"golang.org/x/oauth2"
	oauthGitlab "golang.org/x/oauth2/gitlab"
	"kloudlite.io/apps/auth/internal/domain"
	"kloudlite.io/pkg/errors"
)

type gitlabI struct {
	cfg *oauth2.Config
}

func (gl *gitlabI) GetOAuthToken(ctx context.Context, token *oauth2.Token) (*oauth2.Token, error) {
	return gl.cfg.TokenSource(ctx, token).Token()
}

func (gl *gitlabI) Authorize(_ context.Context, state string) (string, error) {
	return gl.cfg.AuthCodeURL(state), nil
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

	return &gitlabI{cfg: &cfg}
}
