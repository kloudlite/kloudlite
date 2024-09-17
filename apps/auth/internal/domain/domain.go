package domain

import (
	"context"

	"github.com/kloudlite/api/apps/auth/internal/entities"
	"github.com/kloudlite/api/common"
	"github.com/kloudlite/api/pkg/repos"
)

type Domain interface {
	SetRemoteLoginAuthHeader(ctx context.Context, loginId repos.ID, authHeader string) error
	GetRemoteLogin(ctx context.Context, loginId repos.ID, secret string) (*entities.RemoteLogin, error)
	CreateRemoteLogin(ctx context.Context, secret string) (repos.ID, error)

	Login(ctx context.Context, email string, password string) (*common.AuthSession, error)
	SignUp(ctx context.Context, name string, email string, password string) (*common.AuthSession, error)
	EnsureUserByEmail(ctx context.Context, email string) (*entities.User, error)
	GetUserById(ctx context.Context, id repos.ID) (*entities.User, error)
	GetUserByEmail(ctx context.Context, email string) (*entities.User, error)
	SetUserMetadata(ctx context.Context, userId repos.ID, metadata entities.UserMetadata) (*entities.User, error)
	ClearUserMetadata(ctx context.Context, id repos.ID) (*entities.User, error)
	VerifyEmail(ctx context.Context, token string) (*common.AuthSession, error)
	ResetPassword(ctx context.Context, token string, password string) (bool, error)
	RequestResetPassword(ctx context.Context, email string) (bool, error)
	ChangeEmail(ctx context.Context, id repos.ID, email string) (bool, error)
	ResendVerificationEmail(ctx context.Context, userId repos.ID) (bool, error)
	ChangePassword(ctx context.Context, id repos.ID, currentPassword string, newPassword string) (bool, error)
	GetAccessToken(ctx context.Context, provider string, userId string, tokenId string) (*entities.AccessToken, error)
	GetLoginDetails(ctx context.Context, provider string, state *string) (string, error)
	InviteUser(ctx context.Context, email string, name string) (repos.ID, error)
	OauthRequestLogin(ctx context.Context, provider string, state string) (string, error)
	OauthLogin(ctx context.Context, provider string, state string, code string) (*common.AuthSession, error)
	OauthAddLogin(ctx context.Context, userId repos.ID, provider string, state string, code string) (bool, error)

	/// Invite code
	//ListInviteCodes(ctx context.Context) ([]*entities.InviteCode, error)
	//GetInviteCode(ctx context.Context, name string) (*entities.InviteCode, error)

	CreateInviteCode(ctx context.Context, name string, inviteCode string) (*entities.InviteCode, error)
	DeleteInviteCode(ctx context.Context, invCodeId string) error
	// UpdateInviteCode(ctx context.Context, invCode entities.InviteCode) (*entities.InviteCode, error)

	VerifyInviteCode(ctx context.Context, userId repos.ID, invitationCode string) (bool, error)
}

type Messenger interface {
	SendEmail(ctx context.Context, template string, payload map[string]any) error
}
