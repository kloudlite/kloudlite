package domain

import (
	"context"
	"fmt"
	"strings"
	"time"

	"kloudlite.io/common"
	"kloudlite.io/pkg/cache"
	"kloudlite.io/pkg/errors"
	"kloudlite.io/pkg/functions"
	"kloudlite.io/pkg/repos"
)

func generateId(prefix string) string {
	id, e := functions.CleanerNanoid(28)
	if e != nil {
		panic(fmt.Errorf("could not get cleanerNanoid()"))
	}
	return fmt.Sprintf("%s-%s", prefix, strings.ToLower(id))
}

type domainI struct {
	userRepo        repos.DbRepo[*User]
	accessTokenRepo repos.DbRepo[*AccessToken]
	messenger       Messenger
	verifyTokenRepo cache.Repo[*VerifyToken]
}

func (d *domainI) OauthAddLogin(ctx context.Context, id repos.ID, provider string, state string, code string) (bool, error) {
	//TODO implement me
	panic("implement me")
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
	session := common.NewSession(
		string(matched.Id),
		matched.Email,
		matched.Verified,
		"email/password",
	)
	cache.SetSession(ctx, session)
	return session, nil
}

func (d *domainI) InviteUser(ctx context.Context, email string, name string) (repos.ID, error) {
	//TODO implement me
	panic("implement me")
}

func (d *domainI) generateAndSendVerificationToken(ctx context.Context, user *User) error {
	verificationToken := generateId("invite")
	err := d.verifyTokenRepo.Set(ctx, verificationToken, &VerifyToken{
		Token:  verificationToken,
		UserId: string(user.Id),
	})
	if err != nil {
		return err
	}
	err = d.messenger.SendVerificationEmail(ctx, verificationToken, user)
	if err != nil {
		return err
	}
	return nil
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
	err = d.generateAndSendVerificationToken(ctx, create)
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
	user, err := d.userRepo.FindById(ctx, userId)
	if err != nil {
		return nil, err
	}
	user.Metadata = metadata
	updated, err := d.userRepo.UpdateById(ctx, userId, user)
	if err != nil {
		return nil, err
	}
	return updated, nil
}

func (d *domainI) ClearUserMetadata(ctx context.Context, userId repos.ID) (*User, error) {
	user, err := d.userRepo.FindById(ctx, userId)
	if err != nil {
		return nil, err
	}
	user.Metadata = nil
	updated, err := d.userRepo.UpdateById(ctx, userId, user)
	if err != nil {
		return nil, err
	}
	return updated, nil
}

func (d *domainI) VerifyEmail(ctx context.Context, token string) (*common.AuthSession, error) {
	v, err := d.verifyTokenRepo.Get(ctx, token)
	if err != nil {
		return nil, err
	}
	user, err := d.userRepo.FindById(ctx, repos.ID(v.UserId))
	if err != nil {
		return nil, err
	}
	user.Verified = true
	u, err := d.userRepo.UpdateById(ctx, repos.ID(v.UserId), user)
	if err != nil {
		return nil, err
	}
	return common.NewSession(string(u.Id), u.Email, u.Verified, "email/verify"), nil
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

func fxDomain(
	userRepo repos.DbRepo[*User],
	accessTokenRepo repos.DbRepo[*AccessToken],
	verifyTokenRepo cache.Repo[*VerifyToken],
	messenger Messenger,
) Domain {
	return &domainI{
		userRepo,
		accessTokenRepo,
		messenger,
		verifyTokenRepo,
	}
}
