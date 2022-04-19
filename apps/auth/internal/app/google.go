package app

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"kloudlite.io/apps/auth/internal/domain"
	"kloudlite.io/pkg/errors"
)

type googleI struct {
	cfg *oauth2.Config
}

func (g *googleI) Authorize(ctx context.Context, state string) (string, error) {
	return g.cfg.AuthCodeURL(state), nil
}

func (g *googleI) Callback(ctx context.Context, code string, state string) (*domain.GoogleUser, *oauth2.Token, error) {
	nCode, err := url.PathUnescape(code)
	if err != nil {
		return nil, nil, errors.NewEf(err, "could not UnEscape string code %q", code)
	}
	t, err := g.cfg.Exchange(ctx, nCode)
	if err != nil {
		return nil, nil, errors.NewEf(err, "could not exchange oauth code for token")
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://www.googleapis.com/oauth2/v3/userinfo", nil)
	if err != nil {
		return nil, nil, errors.NewEf(err, "could not build http.Request")
	}
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %v", t.AccessToken))
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, nil, errors.NewEf(err, "making http call for user to google")
	}
	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, nil, errors.NewEf(err, "reading response buffer")
	}
	var gu domain.GoogleUser
	err = json.Unmarshal(b, &gu)
	if err != nil {
		return nil, nil, errors.NewEf(err, "marshalling bytes to struct GoogleUser")
	}
	return &gu, t, nil
}

type GoogleOAuth interface {
	GoogleConfig() (clientId, clientSecret, callbackUrl string)
}

func fxGoogle(env *Env) domain.Google {
	clientId, clientSecret, callbackUrl := env.GoogleConfig()
	cfg := &oauth2.Config{
		ClientID:     clientId,
		ClientSecret: clientSecret,
		Endpoint:     google.Endpoint,
		RedirectURL:  callbackUrl,
		Scopes: []string{
			"https://www.googleapis.com/auth/userinfo.profile",
			"https://www.googleapis.com/auth/userinfo.email",
		},
	}
	return &googleI{cfg}
}
