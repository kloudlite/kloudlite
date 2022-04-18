package domain

import (
	"context"

	"golang.org/x/oauth2"
)

type Github interface {
	Authorize(ctx context.Context, state string) (string, error)
	Callback(ctx context.Context, code, state string) (*User, *oauth2.Token, error)
	GetToken(ctx context.Context, token *oauth2.Token) (*oauth2.Token, error)
	RefreshToken()
	GetAppToken()
	GetRepoToken()
}
