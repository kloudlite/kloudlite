package domain

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"kloudlite.io/pkg/messaging"
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
	resetTokenRepo  cache.Repo[*ResetPasswordToken]
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
			"email": email,
		},
	})
	if err != nil {
		return nil, err
	}
	bytes := md5.Sum([]byte(password + matched.PasswordSalt))
	if matched.Password != hex.EncodeToString(bytes[:]) {
		return nil, errors.New("not valid credentials")
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

func (d *domainI) SignUp(ctx context.Context, name string, email string, password string) (*common.AuthSession, error) {
	matched, err := d.userRepo.FindOne(ctx, repos.Query{
		Filter: repos.Filter{
			"email": email,
		},
	})
	if matched != nil {
		return nil, errors.New("User Already exist")
	}
	salt := generateId("salt")
	sum := md5.Sum([]byte(password + salt))
	create, err := d.userRepo.Create(ctx, &User{
		Name:         name,
		Email:        email,
		Password:     hex.EncodeToString(sum[:]),
		Verified:     false,
		Metadata:     nil,
		Joined:       time.Now(),
		PasswordSalt: salt,
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
	return common.NewSession(
		string(u.Id),
		u.Email,
		u.Verified,
		"email/verify",
	), nil
}

func (d *domainI) ResetPassword(ctx context.Context, token string, password string) (bool, error) {
	get, err := d.resetTokenRepo.Get(ctx, token)
	if err != nil {
		return false, err
	}
	if get == nil {
		return false, errors.New("unable to verify link")
	}
	user, err := d.userRepo.FindById(ctx, repos.ID(get.UserId))
	if err != nil {
		return false, errors.New("unable to find user")
	}
	salt := generateId("salt")
	sum := md5.Sum([]byte(password + salt))
	user.Password = hex.EncodeToString(sum[:])
	user.PasswordSalt = salt
	fmt.Println(user)
	_, err = d.userRepo.UpdateById(ctx, repos.ID(get.UserId), user)
	if err != nil {
		return false, err
	}
	err = d.resetTokenRepo.Drop(ctx, token)
	if err != nil {
		// TODO silent fail
		fmt.Println("[ERROR]", err)
		return false, nil
	}
	return true, nil
}

func (d *domainI) RequestResetPassword(ctx context.Context, email string) (bool, error) {
	resetToken := generateId("reset")
	one, err := d.userRepo.FindOne(ctx, repos.Query{Filter: repos.Filter{
		"email": email,
	}})
	if err != nil {
		return false, err
	}
	err = d.resetTokenRepo.SetWithExpiry(ctx, resetToken, &ResetPasswordToken{
		Token:  resetToken,
		UserId: string(one.Id),
	}, time.Second*24*60*60)
	if err != nil {
		return false, err
	}
	err = d.sendResetPasswordEmail(ctx, resetToken, one)
	if err != nil {
		return false, err
	}
	return true, nil
}

func (d *domainI) LoginWithInviteToken(ctx context.Context, token string) (*common.AuthSession, error) {
	//TODO implement me
	panic("implement me")
}

func (d *domainI) ChangeEmail(ctx context.Context, id repos.ID, email string) (bool, error) {
	user, err := d.userRepo.FindById(ctx, id)
	if err != nil {
		return false, err
	}
	user.Email = email
	updated, err := d.userRepo.UpdateById(ctx, id, user)
	if err != nil {
		return false, err
	}
	err = d.generateAndSendVerificationToken(ctx, updated)
	if err != nil {
		return false, err
	}
	return true, nil
}

func (d *domainI) ResendVerificationEmail(ctx context.Context, userId repos.ID) (bool, error) {
	user, err := d.userRepo.FindById(ctx, userId)
	if err != nil {
		return false, err
	}
	err = d.generateAndSendVerificationToken(ctx, user)
	if err != nil {
		return false, err
	}
	return true, err
}

func (d *domainI) ChangePassword(ctx context.Context, id repos.ID, currentPassword string, newPassword string) (bool, error) {
	user, err := d.userRepo.FindById(ctx, id)
	if err != nil {
		return false, err
	}
	sum := md5.Sum([]byte(currentPassword + user.PasswordSalt))
	if user.Password == hex.EncodeToString(sum[:]) {
		salt := generateId("salt")
		user.PasswordSalt = salt
		newSum := md5.Sum([]byte(newPassword + user.PasswordSalt))
		user.Password = hex.EncodeToString(newSum[:])
		_, err := d.userRepo.UpdateById(ctx, id, user)
		if err != nil {
			return false, err
		}
		// TODO send comm
		return true, nil
	}
	return false, errors.New("invalid credentials")
}

func (d *domainI) OauthLogin(ctx context.Context, provider string, state string, code string) (*common.AuthSession, error) {
	//TODO implement me
	panic("implement me")
}

func (d *domainI) sendResetPasswordEmail(ctx context.Context, token string, user *User) error {
	return d.messenger.SendEmail(ctx, "reset-password", messaging.Json{
		"token":    token,
		"userName": user.Name,
		"userId":   user.Id,
	})
}

func (d *domainI) sendVerificationEmail(ctx context.Context, token string, user *User) error {
	return d.messenger.SendEmail(ctx, "verify-email", messaging.Json{
		"token":    token,
		"userName": user.Name,
		"userId":   user.Id,
	})
}

func (d *domainI) generateAndSendVerificationToken(ctx context.Context, user *User) error {
	verificationToken := generateId("invite")
	err := d.verifyTokenRepo.SetWithExpiry(ctx, verificationToken, &VerifyToken{
		Token:  verificationToken,
		UserId: string(user.Id),
	}, time.Second*24*60*60)
	if err != nil {
		return err
	}
	err = d.sendVerificationEmail(ctx, verificationToken, user)
	if err != nil {
		return err
	}
	return nil
}

func fxDomain(
	userRepo repos.DbRepo[*User],
	accessTokenRepo repos.DbRepo[*AccessToken],
	verifyTokenRepo cache.Repo[*VerifyToken],
	resetTokenRepo cache.Repo[*ResetPasswordToken],
	messenger Messenger,
) Domain {
	return &domainI{
		userRepo,
		accessTokenRepo,
		messenger,
		verifyTokenRepo,
		resetTokenRepo,
	}
}
