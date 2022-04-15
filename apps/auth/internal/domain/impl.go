package domain

import (
	"context"
	"kloudlite.io/common"
	"kloudlite.io/pkg/errors"
	"kloudlite.io/pkg/repos"
	"time"
)

type domainI struct {
	userRepo        repos.DbRepo[*User]
	accessTokenRepo repos.DbRepo[*AccessToken]
}

func (d *domainI) GetUserById(ctx context.Context, id repos.ID) (*User, error) {
	return d.userRepo.FindById(ctx, id)
}

func (d *domainI) GetUserByEmail(ctx context.Context, email string) (*User, error) {
	//TODO implement me
	panic("implement me")
}

func (d *domainI) GetLoginDetails(ctx context.Context, provider string, state *string) (string, error) {
	//TODO implement me
	panic("implement me")
}

func (d *domainI) Login(ctx context.Context, email string, password string) (*common.AuthSession, error) {
	matched, err := d.userRepo.FindOne(ctx, repos.Query{
		Filter: repos.Filter{
			"email":    email,
			"password": password,
		},
	})
	if err != nil {
		return nil, err
	}
	return common.NewSession(
		string(matched.Id),
		matched.Email,
		matched.Verified,
		"email/password",
	), nil
}

func (d *domainI) InviteUser(ctx context.Context, email string, name string) (repos.ID, error) {
	//TODO implement me
	panic("implement me")
}

func (d *domainI) SignUp(ctx context.Context, name string, email string, password string) (*common.AuthSession, error) {
	matched, err := d.userRepo.FindOne(ctx, repos.Query{
		Filter: repos.Filter{
			"email": email,
		},
	})
	if matched != nil {
		return nil, errors.New("User Already exist")
	}
	create, err := d.userRepo.Create(ctx, &User{
		Name:     name,
		Email:    email,
		Password: password,
		Verified: false,
		Metadata: nil,
		Joined:   time.Now(),
	})
	if err != nil {
		return nil, err
	}
	return common.NewSession(
		string(create.Id),
		create.Email,
		create.Verified,
		"email/password",
	), nil
}

func (d *domainI) SetUserMetadata(ctx context.Context, userId repos.ID, metadata UserMetadata) (*User, error) {
	//TODO implement me
	panic("implement me")
}

func (d *domainI) ClearUserMetadata(ctx context.Context, id repos.ID) (*User, error) {
	//TODO implement me
	panic("implement me")
}

func (d *domainI) VerifyEmail(ctx context.Context, token string) (*common.AuthSession, error) {
	//TODO implement me
	panic("implement me")
}

func (d *domainI) ResetPassword(ctx context.Context, token string, password string) (bool, error) {
	//TODO implement me
	panic("implement me")
}

func (d *domainI) RequestResetPassword(ctx context.Context, email string) (bool, error) {
	//TODO implement me
	panic("implement me")
}

func (d *domainI) LoginWithInviteToken(ctx context.Context, token string) (*common.AuthSession, error) {
	//TODO implement me
	panic("implement me")
}

func (d *domainI) ChangeEmail(ctx context.Context, id repos.ID, email string) (bool, error) {
	//TODO implement me
	panic("implement me")
}

func (d *domainI) ResendVerificationEmail(ctx context.Context, email string) (bool, error) {
	//TODO implement me
	panic("implement me")
}

func (d *domainI) VerifyChangeEmail(ctx context.Context, token string) (bool, error) {
	//TODO implement me
	panic("implement me")
}

func (d *domainI) ChangePassword(ctx context.Context, id repos.ID, currentPassword string, newPassword string) (bool, error) {
	//TODO implement me
	panic("implement me")
}

func (d *domainI) OauthLogin(ctx context.Context, provider string, state string, code string) (*common.AuthSession, error) {
	//TODO implement me
	panic("implement me")
}

func (d *domainI) OauthAddLogin(ctx context.Context, id repos.ID, provider string, state string, code string) (bool, error) {
	//TODO implement me
	panic("implement me")
}

func fxDomain(
	userRepo repos.DbRepo[*User],
	accessTokenRepo repos.DbRepo[*AccessToken],
) Domain {
	return &domainI{userRepo, accessTokenRepo}
}
