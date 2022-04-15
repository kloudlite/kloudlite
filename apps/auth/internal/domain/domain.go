package domain

import (
	"context"
	"kloudlite.io/common"
	"kloudlite.io/pkg/repos"
)

type Domain interface {
	GetUserById(ctx context.Context, id repos.ID) (*User, error)
	GetUserByEmail(ctx context.Context, email string) (*User, error)
	GetLoginDetails(ctx context.Context, provider string, state *string) (string, error)
	Login(ctx context.Context, email string, password string) (*common.AuthSession, error)
	InviteUser(ctx context.Context, email string, name string) (repos.ID, error)
	SignUp(ctx context.Context, name string, email string, password string) (*common.AuthSession, error)
	SetUserMetadata(ctx context.Context, userId repos.ID, metadata UserMetadata) (*User, error)
	ClearUserMetadata(ctx context.Context, id repos.ID) (*User, error)
	VerifyEmail(ctx context.Context, token string) (*common.AuthSession, error)
	ResetPassword(ctx context.Context, token string, password string) (bool, error)
	RequestResetPassword(ctx context.Context, email string) (bool, error)
	LoginWithInviteToken(ctx context.Context, token string) (*common.AuthSession, error)
	ChangeEmail(ctx context.Context, id repos.ID, email string) (bool, error)
	ResendVerificationEmail(ctx context.Context, email string) (bool, error)
	VerifyChangeEmail(ctx context.Context, token string) (bool, error)
	ChangePassword(ctx context.Context, id repos.ID, currentPassword string, newPassword string) (bool, error)
	OauthLogin(ctx context.Context, provider string, state string, code string) (*common.AuthSession, error)
	OauthAddLogin(ctx context.Context, id repos.ID, provider string, state string, code string) (bool, error)
}
