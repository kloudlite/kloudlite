package domain

import (
	"context"
	"kloudlite.io/pkg/repos"
)

type Domain interface {
	GetUserById(ctx context.Context, id repos.ID) (*User, error)
	GetUserByEmail(ctx context.Context, email string) (*User, error)
	GetLoginDetails(ctx context.Context, provider string, state *string) (string, error)
	Login(ctx context.Context, email string, password string) (*Session, error)
	InviteUser(ctx context.Context, email string, name string) (repos.ID, error)
	SignUp(ctx context.Context, name string, email string, password string) (*Session, error)
	Logout(ctx context.Context, userId repos.ID) (bool, error)
	SetUserMetadata(ctx context.Context, userId repos.ID, metadata UserMetadata) (*User, error)
	ClearUserMetadata(ctx context.Context, id repos.ID) (*User, error)
	VerifyEmail(ctx context.Context, token string) (*Session, error)
	ResetPassword(ctx context.Context, token string, password string) (bool, error)
	RequestResetPassword(ctx context.Context, email string) (bool, error)
	LoginWithInviteToken(ctx context.Context, token string) (*Session, error)
	ChangeEmail(ctx context.Context, id repos.ID, email string) (bool, error)
	ResendVerificationEmail(ctx context.Context, email string) (bool, error)
	VerifyChangeEmail(ctx context.Context, token string) (bool, error)
	ChangePassword(ctx context.Context, id repos.ID, currentPassword string, newPassword string) (bool, error)
	OauthLogin(ctx context.Context, provider string, state string, code string) (*Session, error)
	OauthAddLogin(ctx context.Context, id repos.ID, provider string, state string, code string) (bool, error)
}
