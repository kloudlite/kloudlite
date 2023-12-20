package domain

import (
	"context"
	"crypto/md5"
	b64 "encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/kloudlite/api/constants"

	"github.com/kloudlite/api/grpc-interfaces/kloudlite.io/rpc/comms"

	"github.com/kloudlite/api/common"
	"github.com/kloudlite/api/pkg/cache"
	"github.com/kloudlite/api/pkg/errors"
	"github.com/kloudlite/api/pkg/functions"
	"github.com/kloudlite/api/pkg/logging"
	"github.com/kloudlite/api/pkg/repos"
	"golang.org/x/oauth2"
)

func generateId(prefix string) string {
	id := functions.CleanerNanoidOrDie(28)
	return fmt.Sprintf("%s-%s", prefix, strings.ToLower(id))
}

func newAuthSession(userId repos.ID, userEmail string, userName string, userVerified bool, loginMethod string) *common.AuthSession {
	sessionId := generateId("sess")
	s := &common.AuthSession{
		UserId:       userId,
		UserEmail:    userEmail,
		UserVerified: userVerified,
		UserName:     userName,
		LoginMethod:  loginMethod,
	}
	s.SetId(repos.ID(sessionId))
	return s
}

type domainI struct {
	userRepo        repos.DbRepo[*User]
	accessTokenRepo repos.DbRepo[*AccessToken]
	commsClient     comms.CommsClient
	verifyTokenRepo cache.Repo[*VerifyToken]
	resetTokenRepo  cache.Repo[*ResetPasswordToken]
	logger          logging.Logger
	github          Github
	gitlab          Gitlab
	google          Google
	remoteLoginRepo repos.DbRepo[*RemoteLogin]
}

func (d *domainI) SetRemoteLoginAuthHeader(ctx context.Context, loginId repos.ID, authHeader string) error {
	remoteLogin, err := d.remoteLoginRepo.FindById(ctx, loginId)
	if err != nil {
		return errors.NewEf(err, "could not find remote login")
	}
	remoteLogin.AuthHeader = authHeader
	remoteLogin.LoginStatus = LoginStatusSucceeded
	_, err = d.remoteLoginRepo.UpdateById(ctx, loginId, remoteLogin)
	if err != nil {
		return errors.NewEf(err, "could not update remote login")
	}
	return nil
}

func (d *domainI) GetRemoteLogin(ctx context.Context, loginId repos.ID, secret string) (*RemoteLogin, error) {
	id, err := d.remoteLoginRepo.FindById(ctx, loginId)
	if err != nil {
		return nil, errors.NewEf(err, "could not find remote login")
	}
	if id.Secret != secret {
		return nil, errors.New("invalid secret")
	}
	return id, err
}

func (d *domainI) CreateRemoteLogin(ctx context.Context, secret string) (repos.ID, error) {
	create, err := d.remoteLoginRepo.Create(
		ctx, &RemoteLogin{
			LoginStatus: LoginStatusPending,
			Secret:      secret,
		},
	)
	if err != nil {
		return "", err
	}
	return create.Id, nil
}

func (d *domainI) EnsureUserByEmail(ctx context.Context, email string) (*User, error) {
	u, err := d.userRepo.Create(
		ctx, &User{
			Email: email,
		},
	)
	if err != nil {
		return nil, err
	}
	return u, err
}

func (d *domainI) OauthAddLogin(ctx context.Context, userId repos.ID, provider string, state string, code string) (bool, error) {
	user, err := d.userRepo.FindById(ctx, userId)
	if err != nil {
		return false, errors.NewEf(err, "could not find user")
	}
	switch provider {
	case constants.ProviderGithub:
		{
			u, t, err := d.github.Callback(ctx, code, state)
			if err != nil {
				return false, errors.NewEf(err, "could not login to github")
			}
			_, err = d.addOAuthLogin(ctx, provider, t, user, u.AvatarURL)
			if err != nil {
				return false, err
			}
			return true, err
		}

	case constants.ProviderGitlab:
		{
			u, t, err := d.gitlab.Callback(ctx, code, state)
			if err != nil {
				return false, errors.NewEf(err, "could not login to gitlab")
			}
			_, err = d.afterOAuthLogin(ctx, provider, t, user, &u.AvatarURL)
			if err != nil {
				return false, err
			}
			return true, err
		}

	default:
		{
			return false, errors.Newf("unknown provider=%s, aborting request", provider)
		}
	}

}

func (d *domainI) GetUserById(ctx context.Context, id repos.ID) (*User, error) {
	return d.userRepo.FindById(ctx, id)
}

func (d *domainI) GetUserByEmail(ctx context.Context, email string) (*User, error) {
	return d.userRepo.FindOne(ctx, repos.Filter{"email": email})
}

func (d *domainI) Login(ctx context.Context, email string, password string) (*common.AuthSession, error) {
	user, err := d.userRepo.FindOne(ctx, repos.Filter{"email": email})
	if err != nil {
		return nil, err
	}

	if user == nil {
		d.logger.Warnf("user not found for email=%s", email)
		return nil, errors.Newf("not valid credentials")
	}

	bytes := md5.Sum([]byte(password + user.PasswordSalt))
	// TODO (nxtcoder17): use crypto/subtle to compare hashes, to avoid timing attacks, also does not work now
	if user.Password != hex.EncodeToString(bytes[:]) {
		return nil, errors.New("not valid credentials")
	}
	session := newAuthSession(user.Id, user.Email, user.Name, user.Verified, "email/password")
	return session, nil
}

func (d *domainI) SignUp(ctx context.Context, name string, email string, password string) (*common.AuthSession, error) {
	matched, err := d.userRepo.FindOne(ctx, repos.Filter{"email": email})

	if err != nil {
		if matched != nil {
			return nil, err
		}
	}

	if matched != nil {
		if matched.Email == email {
			return nil, errors.Newf("user(email=%s) already exists", email)
		}
	}

	salt := generateId("salt")
	sum := md5.Sum([]byte(password + salt))
	user, err := d.userRepo.Create(
		ctx, &User{
			Name:         name,
			Email:        email,
			Password:     hex.EncodeToString(sum[:]),
			Verified:     false,
			Metadata:     nil,
			Joined:       time.Now(),
			PasswordSalt: salt,
		},
	)

	if err != nil {
		return nil, err
	}

	err = d.generateAndSendVerificationToken(ctx, user)
	if err != nil {
		return nil, err
	}

	return newAuthSession(user.Id, user.Email, user.Name, user.Verified, "email/password"), nil
}

func (d *domainI) GetLoginDetails(ctx context.Context, provider string, state *string) (string, error) {
	// TODO implement me
	panic("implement me")
}

func (d *domainI) InviteUser(ctx context.Context, email string, name string) (repos.ID, error) {
	// TODO implement me
	panic("implement me")
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
	user, err := d.userRepo.FindById(ctx, v.UserId)
	if err != nil {
		return nil, err
	}
	user.Verified = true
	u, err := d.userRepo.UpdateById(ctx, v.UserId, user)
	if err != nil {
		return nil, err
	}
	if _, err := d.commsClient.SendWelcomeEmail(
		ctx, &comms.WelcomeEmailInput{
			Email: user.Email,
			Name:  user.Name,
		},
	); err != nil {
		d.logger.Errorf(err)
	}

	return newAuthSession(u.Id, u.Email, u.Name, u.Verified, "email/verify"), nil
}

func (d *domainI) ResetPassword(ctx context.Context, token string, password string) (bool, error) {
	get, err := d.resetTokenRepo.Get(ctx, token)
	if err != nil || get == nil {
		return false, errors.NewEf(err, "failed to verify reset password token")
	}

	user, err := d.userRepo.FindById(ctx, get.UserId)
	if err != nil {
		return false, errors.NewEf(err, "unable to find user")
	}
	salt := generateId("salt")
	sum := md5.Sum([]byte(password + salt))
	user.Password = hex.EncodeToString(sum[:])
	user.PasswordSalt = salt
	_, err = d.userRepo.UpdateById(ctx, repos.ID(get.UserId), user)
	if err != nil {
		return false, err
	}

	err = d.resetTokenRepo.Drop(ctx, token)
	if err != nil {
		// TODO silent fail
		d.logger.Errorf(err, "could not delete resetPassword token")
		return false, nil
	}
	return true, nil
}

func (d *domainI) RequestResetPassword(ctx context.Context, email string) (bool, error) {
	resetToken := generateId("reset")
	one, err := d.userRepo.FindOne(ctx, repos.Filter{"email": email})
	if err != nil {
		return false, err
	}
	if one == nil {
		return false, errors.New("no account present with provided email, register your account first.")
	}
	err = d.resetTokenRepo.SetWithExpiry(
		ctx,
		resetToken,
		&ResetPasswordToken{Token: resetToken, UserId: one.Id},
		time.Second*24*60*60,
	)
	if err != nil {
		return false, err
	}
	err = d.sendResetPasswordEmail(ctx, resetToken, one)
	if err != nil {
		return false, err
	}
	return true, nil
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

func (d *domainI) OauthRequestLogin(ctx context.Context, provider string, state string) (string, error) {
	if provider == constants.ProviderGithub {
		return d.github.Authorize(ctx, state)
	}

	if provider == constants.ProviderGitlab {
		return d.gitlab.Authorize(ctx, state)
	}

	if provider == constants.ProviderGoogle {
		return d.google.Authorize(ctx, state)
	}

	return "", errors.Newf("Unsupported provider (%v)", provider)
}

func (d *domainI) addOAuthLogin(ctx context.Context, provider string, token *oauth2.Token, u *User, avatarUrl *string) (*User, error) {
	user, err := d.userRepo.FindOne(ctx, repos.Filter{"email": u.Email})
	if err != nil {
		return nil, errors.NewEf(err, "could not find user")
	}
	if user == nil {
		user = u
		user.Joined = time.Now()
		user, err = d.userRepo.Create(ctx, user)
		d.commsClient.SendWelcomeEmail(
			ctx, &comms.WelcomeEmailInput{
				Email: user.Email,
				Name:  user.Name,
			},
		)
		if err != nil {
			return nil, errors.NewEf(err, "could not create user (email=%s)", user.Email)
		}
	}
	t, err := d.accessTokenRepo.Upsert(
		ctx, repos.Filter{"email": user.Email, "provider": provider}, &AccessToken{
			UserId:   user.Id,
			Email:    user.Email,
			Provider: provider,
			Token:    token,
		},
	)

	if err != nil {
		return nil, errors.NewEf(err, "could not store access token in repo")
	}

	p := &ProviderDetail{TokenId: t.Id, Avatar: avatarUrl}

	if provider == constants.ProviderGithub {
		user.ProviderGithub = p
	}

	if provider == constants.ProviderGitlab {
		user.ProviderGitlab = p
	}

	if provider == constants.ProviderGoogle {
		user.ProviderGoogle = p
	}

	user.Verified = true
	user, err = d.userRepo.UpdateById(ctx, user.Id, user)
	if err != nil {
		return nil, errors.NewEf(err, "could not update user")
	}
	return user, nil
}

func (d *domainI) afterOAuthLogin(ctx context.Context, provider string, token *oauth2.Token, newUser *User, avatarUrl *string) (*common.AuthSession, error) {
	user, err := d.addOAuthLogin(ctx, provider, token, newUser, avatarUrl)
	if err != nil {
		return nil, err
	}
	session := newAuthSession(user.Id, user.Email, user.Name, user.Verified, fmt.Sprintf("oauth2/%s", provider))
	return session, nil
}

func (d *domainI) OauthLogin(ctx context.Context, provider string, state string, code string) (*common.AuthSession, error) {
	switch provider {
	case constants.ProviderGithub:
		{
			u, t, err := d.github.Callback(ctx, code, state)
			if err != nil {
				return nil, errors.NewEf(err, "could not login to github")
			}

			email, err := func() (string, error) {
				if u.Email != nil {
					return *u.Email, nil
				}
				d.logger.Infof("user has no public email, trying to get his protected email")
				pEmail, err := d.github.GetPrimaryEmail(ctx, t)
				if err != nil {
					return "", err
				}
				return pEmail, nil
			}()
			if err != nil {
				return nil, err
			}

			name := func() string {
				if u.Name != nil {
					return *u.Name
				}
				return u.GetLogin()
			}()

			user := &User{
				Name:   name,
				Avatar: u.AvatarURL,
				Email:  email,
			}
			return d.afterOAuthLogin(ctx, provider, t, user, u.AvatarURL)
		}

	case constants.ProviderGitlab:
		{
			u, t, err := d.gitlab.Callback(ctx, code, state)
			if err != nil {
				return nil, errors.NewEf(err, "could not login to gitlab")
			}

			user := &User{
				Name:   u.Name,
				Avatar: &u.AvatarURL,
				Email:  u.Email,
			}

			return d.afterOAuthLogin(ctx, provider, t, user, &u.AvatarURL)
		}

	case constants.ProviderGoogle:
		{
			u, t, err := d.google.Callback(ctx, code, state)
			if err != nil {
				return nil, errors.NewEf(err, "could not login to google")
			}

			user := &User{
				Name:   u.Name,
				Avatar: u.AvatarURL,
				Email:  u.Email,
			}

			return d.afterOAuthLogin(ctx, provider, t, user, u.AvatarURL)
		}
	default:
		{
			return nil, errors.Newf("unknown provider=%s, aborting request", provider)
		}
	}
}

func (gl *domainI) Hash(t *oauth2.Token) (string, error) {
	b, err := json.Marshal(t)
	if err != nil {
		return "", err
	}
	return b64.StdEncoding.EncodeToString(b), nil
}

func (d *domainI) GetAccessToken(ctx context.Context, provider string, userId string, tokenId string) (*AccessToken, error) {
	if tokenId == "" && (provider == "" || userId == "") {
		return nil, errors.Newf("bad params: [tokenId, (provider, userId)]")
	}
	q := repos.Filter{"id": tokenId}
	if tokenId == "" {
		q = repos.Filter{"user_id": userId, "provider": provider}
	}
	accToken, err := d.accessTokenRepo.FindOne(ctx, q)
	if err != nil {
		return nil, errors.NewEf(err, "finding access token")
	}
	if accToken == nil {
		return nil, errors.Newf("no token found  for (tokenId=%s, provider=%s, userId=%s)", tokenId, provider, userId)
	}

	hash, err := d.Hash(accToken.Token)
	if err != nil {
		return nil, err
	}

	if provider == "github" {
		token, err := d.github.GetOAuthToken(ctx, accToken.Token)
		if err != nil {
			return nil, errors.NewEf(err, "could not get oauth token")
		}
		hash2, err := d.Hash(token)
		if err != nil {
			return nil, err
		}
		if hash != hash2 {
			accToken.Token = token
		}
	}

	if provider == "gitlab" {
		token, err := d.gitlab.GetOAuthToken(ctx, accToken.Token)
		if err != nil {
			return nil, errors.NewEf(err, "could not get oauth token")
		}
		hash2, err := d.Hash(token)
		if err != nil {
			return nil, err
		}
		if hash != hash2 {
			accToken.Token = token
		}
	}

	_, err = d.accessTokenRepo.UpdateById(ctx, accToken.Id, accToken)
	if err != nil {
		return nil, errors.NewEf(err, "could not update access token")
	}
	// fmt.Println("accToken: ", accToken)
	return accToken, nil
}

func (d *domainI) sendResetPasswordEmail(ctx context.Context, token string, user *User) error {
	_, err := d.commsClient.SendPasswordResetEmail(
		ctx, &comms.PasswordResetEmailInput{
			Email:      user.Email,
			Name:       user.Name,
			ResetToken: token,
		},
	)
	if err != nil {
		return errors.NewEf(err, "could not send password reset email")
	}
	return nil
}

func (d *domainI) sendVerificationEmail(ctx context.Context, token string, user *User) error {
	_, err := d.commsClient.SendVerificationEmail(
		ctx, &comms.VerificationEmailInput{
			Email:             user.Email,
			Name:              user.Name,
			VerificationToken: token,
		},
	)
	if err != nil {
		return errors.NewEf(err, "could not send verification email")
	}
	return nil
}

func (d *domainI) generateAndSendVerificationToken(ctx context.Context, user *User) error {
	verificationToken := generateId("invite")
	err := d.verifyTokenRepo.SetWithExpiry(
		ctx, verificationToken, &VerifyToken{
			Token:  verificationToken,
			UserId: user.Id,
		}, time.Second*24*60*60,
	)
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
	remoteLoginRepo repos.DbRepo[*RemoteLogin],
	verifyTokenRepo cache.Repo[*VerifyToken],
	resetTokenRepo cache.Repo[*ResetPasswordToken],
	github Github,
	gitlab Gitlab,
	google Google,
	logger logging.Logger,
	commsClient comms.CommsClient,
) Domain {
	return &domainI{
		remoteLoginRepo: remoteLoginRepo,
		commsClient:     commsClient,
		userRepo:        userRepo,
		accessTokenRepo: accessTokenRepo,
		verifyTokenRepo: verifyTokenRepo,
		resetTokenRepo:  resetTokenRepo,
		github:          github,
		gitlab:          gitlab,
		google:          google,
		logger:          logger,
	}
}
