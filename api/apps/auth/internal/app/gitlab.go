package app

import (
	"context"
	"fmt"
	"strings"

	"github.com/kloudlite/api/apps/auth/internal/domain"
	"github.com/kloudlite/api/apps/auth/internal/env"
	"github.com/kloudlite/api/pkg/errors"
	"github.com/xanzy/go-gitlab"
	"golang.org/x/oauth2"
	oauthGitlab "golang.org/x/oauth2/gitlab"
)

type gitlabI struct {
	enabled bool
	cfg     *oauth2.Config
}

func (gl *gitlabI) GetOAuthToken(ctx context.Context, token *oauth2.Token) (*oauth2.Token, error) {
	if !gl.enabled {
		return nil, fmt.Errorf("gitlab oauth is disabled")
	}
	return gl.cfg.TokenSource(ctx, token).Token()
}

func (gl *gitlabI) Authorize(_ context.Context, state string) (string, error) {
	if !gl.enabled {
		return "", fmt.Errorf("gitlab oauth is disabled")
	}
	return gl.cfg.AuthCodeURL(state), nil
}

func (gl *gitlabI) Callback(ctx context.Context, code string, state string) (*gitlab.User, *oauth2.Token, error) {
	if !gl.enabled {
		return nil, nil, fmt.Errorf("gitlab oauth is disabled")
	}
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

func fxGitlab(ev *env.Env) domain.Gitlab {
	if !ev.OAuth2Enabled || !ev.OAuth2GitlabEnabled {
		return &gitlabI{enabled: false, cfg: nil}
	}

	cfg := oauth2.Config{
		ClientID:     ev.GitlabClientId,
		ClientSecret: ev.GitlabClientSecret,
		Endpoint:     oauthGitlab.Endpoint,
		RedirectURL:  ev.GitlabCallbackUrl,
		Scopes:       strings.Split(ev.GitlabScopes, ","),
	}

	return &gitlabI{enabled: true, cfg: &cfg}
}
