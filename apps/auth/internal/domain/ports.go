package domain

import (
	"context"

	"github.com/google/go-github/v43/github"
	"github.com/xanzy/go-gitlab"
	"golang.org/x/oauth2"
)

type Github interface {
	Authorize(ctx context.Context, state string) (string, error)
	Callback(ctx context.Context, code, state string) (*github.User, *oauth2.Token, error)
	GetOAuthToken(ctx context.Context, token *oauth2.Token) (*oauth2.Token, error)
}

type Gitlab interface {
	Authorize(ctx context.Context, state string) (string, error)
	Callback(ctx context.Context, code, state string) (*gitlab.User, *oauth2.Token, error)
	GetOAuthToken(ctx context.Context, token *oauth2.Token) (*oauth2.Token, error)
}

// google oauth user struct
type GoogleUser struct {
	Email     string  `json:"email"`
	AvatarURL *string `json:"picture"`
	Name      string  `json:"name"`
}

type Google interface {
	Authorize(ctx context.Context, state string) (string, error)
	Callback(ctx context.Context, code, state string) (*GoogleUser, *oauth2.Token, error)
}
