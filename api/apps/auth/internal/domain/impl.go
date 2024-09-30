package domain

import (
	recaptchaenterprise "cloud.google.com/go/recaptchaenterprise/v2/apiv1"
	recaptchapb "cloud.google.com/go/recaptchaenterprise/v2/apiv1/recaptchaenterprisepb"
	"context"
	"crypto/md5"
	b64 "encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/kloudlite/api/apps/auth/internal/entities"
	"github.com/kloudlite/api/apps/auth/internal/env"

	"github.com/kloudlite/api/constants"

	"github.com/kloudlite/api/grpc-interfaces/kloudlite.io/rpc/comms"

	"github.com/kloudlite/api/common"
	"github.com/kloudlite/api/pkg/errors"
	"github.com/kloudlite/api/pkg/functions"
	"github.com/kloudlite/api/pkg/kv"
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
	userRepo        repos.DbRepo[*entities.User]
	accessTokenRepo repos.DbRepo[*entities.AccessToken]
	commsClient     comms.CommsClient
	verifyTokenRepo kv.Repo[*entities.VerifyToken]
	resetTokenRepo  kv.Repo[*entities.ResetPasswordToken]
	inviteCodeRepo  repos.DbRepo[*entities.InviteCode]
	logger          logging.Logger
	github          Github
	gitlab          Gitlab
	google          Google
	remoteLoginRepo repos.DbRepo[*entities.RemoteLogin]
	recaptchaClient *recaptchaenterprise.Client

	envVars *env.Env
}

func (d *domainI) SetRemoteLoginAuthHeader(ctx context.Context, loginId repos.ID, authHeader string) error {
	remoteLogin, err := d.remoteLoginRepo.FindById(ctx, loginId)
	if err != nil {
		return errors.NewEf(err, "could not find remote login")
	}
	remoteLogin.AuthHeader = authHeader
	remoteLogin.LoginStatus = entities.LoginStatusSucceeded
	_, err = d.remoteLoginRepo.UpdateById(ctx, loginId, remoteLogin)
	if err != nil {
		return errors.NewEf(err, "could not update remote login")
	}
	return nil
}

func (d *domainI) GetRemoteLogin(ctx context.Context, loginId repos.ID, secret string) (*entities.RemoteLogin, error) {
	id, err := d.remoteLoginRepo.FindById(ctx, loginId)
	if err != nil {
		return nil, errors.NewEf(err, "could not find remote login")
	}
	if id.Secret != secret {
		return nil, errors.New("invalid secret")
	}
	return id, errors.NewE(err)
}

func (d *domainI) CreateRemoteLogin(ctx context.Context, secret string) (repos.ID, error) {
	create, err := d.remoteLoginRepo.Create(
		ctx, &entities.RemoteLogin{
			LoginStatus: entities.LoginStatusPending,
			Secret:      secret,
		},
	)
	if err != nil {
		return "", errors.NewE(err)
	}
	return create.Id, nil
}

func (d *domainI) EnsureUserByEmail(ctx context.Context, email string) (*entities.User, error) {
	u, err := d.userRepo.Create(
		ctx, &entities.User{
			Email: email,
		},
	)
	if err != nil {
		return nil, errors.NewE(err)
	}
	return u, errors.NewE(err)
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
				return false, errors.NewE(err)
			}
			return true, errors.NewE(err)
		}

	case constants.ProviderGitlab:
		{
			u, t, err := d.gitlab.Callback(ctx, code, state)
			if err != nil {
				return false, errors.NewEf(err, "could not login to gitlab")
			}
			_, err = d.afterOAuthLogin(ctx, provider, t, user, &u.AvatarURL)
			if err != nil {
				return false, errors.NewE(err)
			}
			return true, errors.NewE(err)
		}

	default:
		{
			return false, errors.Newf("unknown provider=%s, aborting request", provider)
		}
	}
}

func (d *domainI) GetUserById(ctx context.Context, id repos.ID) (*entities.User, error) {
	return d.userRepo.FindById(ctx, id)
}

func (d *domainI) GetUserByEmail(ctx context.Context, email string) (*entities.User, error) {
	return d.userRepo.FindOne(ctx, repos.Filter{"email": email})
}

func (d *domainI) Login(ctx context.Context, email string, password string) (*common.AuthSession, error) {
	user, err := d.userRepo.FindOne(ctx, repos.Filter{"email": email})
	if err != nil {
		return nil, errors.NewE(err)
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

func (d *domainI) verifyCaptcha(ctx context.Context, token string) (bool, error) {
	req := &recaptchapb.CreateAssessmentRequest{
		Parent: fmt.Sprintf("projects/%s", d.envVars.GoogleCloudProjectId), // Project path in the format 'projects/{project-id}'
		Assessment: &recaptchapb.Assessment{
			Event: &recaptchapb.Event{
				Token:   token,
				SiteKey: d.envVars.RecaptchaSiteKey,
			},
		},
	}

	resp, err := d.recaptchaClient.CreateAssessment(ctx, req)
	if err != nil {
		return false, errors.NewE(err)
	}

	if !resp.TokenProperties.Valid {
		return false, errors.Newf("CAPTCHA token is invalid: %s", resp.TokenProperties.InvalidReason)
	}
	return true, nil
}

func (d *domainI) SignUp(ctx context.Context, name string, email string, password string, captchaToken string) (*common.AuthSession, error) {
	isValidCaptcha, err := d.verifyCaptcha(ctx, captchaToken)
	if err != nil {
		return nil, errors.Newf("failed to verify CAPTCHA: %v", err)
	}

	if !isValidCaptcha {
		return nil, errors.New("CAPTCHA verification failed")
	}

	matched, err := d.userRepo.FindOne(ctx, repos.Filter{"email": email})
	if err != nil {
		if matched != nil {
			return nil, errors.NewE(err)
		}
	}

	if matched != nil && matched.Email == email {
		return nil, errors.Newf("user(email=%q) already exists", email)
	}

	salt := generateId("salt")
	sum := md5.Sum([]byte(password + salt))
	user, err := d.userRepo.Create(
		ctx, &entities.User{
			Name:         name,
			Email:        email,
			Password:     hex.EncodeToString(sum[:]),
			Verified:     !d.envVars.UserEmailVerifactionEnabled,
			Approved:     false,
			Metadata:     nil,
			Joined:       time.Now(),
			PasswordSalt: salt,
		},
	)
	if err != nil {
		return nil, errors.NewE(err)
	}

	if d.envVars.UserEmailVerifactionEnabled {
		err = d.generateAndSendVerificationToken(ctx, user)
		if err != nil {
			return nil, errors.NewE(err)
		}
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

func (d *domainI) SetUserMetadata(ctx context.Context, userId repos.ID, metadata entities.UserMetadata) (*entities.User, error) {
	user, err := d.userRepo.FindById(ctx, userId)
	if err != nil {
		return nil, errors.NewE(err)
	}
	user.Metadata = metadata
	updated, err := d.userRepo.UpdateById(ctx, userId, user)
	if err != nil {
		return nil, errors.NewE(err)
	}
	return updated, nil
}

func (d *domainI) ClearUserMetadata(ctx context.Context, userId repos.ID) (*entities.User, error) {
	user, err := d.userRepo.FindById(ctx, userId)
	if err != nil {
		return nil, errors.NewE(err)
	}
	user.Metadata = nil
	updated, err := d.userRepo.UpdateById(ctx, userId, user)
	if err != nil {
		return nil, errors.NewE(err)
	}
	return updated, nil
}

func (d *domainI) VerifyEmail(ctx context.Context, token string) (*common.AuthSession, error) {
	v, err := d.verifyTokenRepo.Get(ctx, token)
	if err != nil {
		return nil, errors.NewE(err)
	}
	user, err := d.userRepo.FindById(ctx, v.UserId)
	if err != nil {
		return nil, errors.NewE(err)
	}
	user.Verified = true
	u, err := d.userRepo.UpdateById(ctx, v.UserId, user)
	if err != nil {
		return nil, errors.NewE(err)
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
		return false, errors.NewE(err)
	}

	err = d.resetTokenRepo.Drop(ctx, token)
	if err != nil {
		// TODO silent fail
		d.logger.Errorf(err, "could not delete resetPassword token")
		return false, nil
	}
	return true, nil
}

func (d *domainI) RequestResetPassword(ctx context.Context, email string, captchaToken string) (bool, error) {
	isValidCaptcha, err := d.verifyCaptcha(ctx, captchaToken)
	if err != nil {
		return false, errors.Newf("failed to verify CAPTCHA: %v", err)
	}

	if !isValidCaptcha {
		return false, errors.New("CAPTCHA verification failed")
	}

	resetToken := generateId("reset")
	one, err := d.userRepo.FindOne(ctx, repos.Filter{"email": email})
	if err != nil {
		return false, errors.NewE(err)
	}
	if one == nil {
		return false, errors.New("no account present with provided email, register your account first.")
	}
	err = d.resetTokenRepo.SetWithExpiry(
		ctx,
		resetToken,
		&entities.ResetPasswordToken{Token: resetToken, UserId: one.Id},
		time.Second*24*60*60,
	)
	if err != nil {
		return false, errors.NewE(err)
	}
	err = d.sendResetPasswordEmail(ctx, resetToken, one)
	if err != nil {
		return false, errors.NewE(err)
	}
	return true, nil
}

func (d *domainI) ChangeEmail(ctx context.Context, id repos.ID, email string) (bool, error) {
	user, err := d.userRepo.FindById(ctx, id)
	if err != nil {
		return false, errors.NewE(err)
	}
	user.Email = email
	updated, err := d.userRepo.UpdateById(ctx, id, user)
	if err != nil {
		return false, errors.NewE(err)
	}
	err = d.generateAndSendVerificationToken(ctx, updated)
	if err != nil {
		return false, errors.NewE(err)
	}
	return true, nil
}

func (d *domainI) ResendVerificationEmail(ctx context.Context, userId repos.ID) (bool, error) {
	user, err := d.userRepo.FindById(ctx, userId)
	if err != nil {
		return false, errors.NewE(err)
	}
	err = d.generateAndSendVerificationToken(ctx, user)
	if err != nil {
		return false, errors.NewE(err)
	}
	return true, errors.NewE(err)
}

func (d *domainI) ChangePassword(ctx context.Context, id repos.ID, currentPassword string, newPassword string) (bool, error) {
	user, err := d.userRepo.FindById(ctx, id)
	if err != nil {
		return false, errors.NewE(err)
	}
	sum := md5.Sum([]byte(currentPassword + user.PasswordSalt))
	if user.Password == hex.EncodeToString(sum[:]) {
		salt := generateId("salt")
		user.PasswordSalt = salt
		newSum := md5.Sum([]byte(newPassword + user.PasswordSalt))
		user.Password = hex.EncodeToString(newSum[:])
		_, err := d.userRepo.UpdateById(ctx, id, user)
		if err != nil {
			return false, errors.NewE(err)
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

func (d *domainI) addOAuthLogin(ctx context.Context, provider string, token *oauth2.Token, u *entities.User, avatarUrl *string) (*entities.User, error) {
	user, err := d.userRepo.FindOne(ctx, repos.Filter{"email": u.Email})
	if err != nil {
		return nil, errors.NewEf(err, "could not find user")
	}
	if user == nil {
		user = u
		user.Joined = time.Now()
		user, err = d.userRepo.Create(ctx, user)
		//if _, err := d.commsClient.SendWelcomeEmail(
		//	ctx, &comms.WelcomeEmailInput{
		//		Email: user.Email,
		//		Name:  user.Name,
		//	},
		//); err != nil {
		//	d.logger.Errorf(err)
		//}
		if _, err := d.commsClient.SendWaitingEmail(
			ctx, &comms.WelcomeEmailInput{
				Email: user.Email,
				Name:  user.Name,
			},
		); err != nil {
			d.logger.Errorf(err)
		}
		if err != nil {
			return nil, errors.NewEf(err, "could not create user (email=%s)", user.Email)
		}
	}
	t, err := d.accessTokenRepo.Upsert(
		ctx, repos.Filter{"email": user.Email, "provider": provider}, &entities.AccessToken{
			UserId:   user.Id,
			Email:    user.Email,
			Provider: provider,
			Token:    token,
		},
	)
	if err != nil {
		return nil, errors.NewEf(err, "could not store access token in repo")
	}

	p := &entities.ProviderDetail{TokenId: t.Id, Avatar: avatarUrl}

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

func (d *domainI) afterOAuthLogin(ctx context.Context, provider string, token *oauth2.Token, newUser *entities.User, avatarUrl *string) (*common.AuthSession, error) {
	user, err := d.addOAuthLogin(ctx, provider, token, newUser, avatarUrl)
	if err != nil {
		return nil, errors.NewE(err)
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
					return "", errors.NewE(err)
				}
				return pEmail, nil
			}()
			if err != nil {
				return nil, errors.NewE(err)
			}

			name := func() string {
				if u.Name != nil {
					return *u.Name
				}
				return u.GetLogin()
			}()

			user := &entities.User{
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

			user := &entities.User{
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

			user := &entities.User{
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
		return "", errors.NewE(err)
	}
	return b64.StdEncoding.EncodeToString(b), nil
}

func (d *domainI) GetAccessToken(ctx context.Context, provider string, userId string, tokenId string) (*entities.AccessToken, error) {
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
		return nil, errors.NewE(err)
	}

	if provider == "github" {
		token, err := d.github.GetOAuthToken(ctx, accToken.Token)
		if err != nil {
			return nil, errors.NewEf(err, "could not get oauth token")
		}
		hash2, err := d.Hash(token)
		if err != nil {
			return nil, errors.NewE(err)
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
			return nil, errors.NewE(err)
		}
		if hash != hash2 {
			accToken.Token = token
		}
	}

	_, err = d.accessTokenRepo.UpdateById(ctx, accToken.Id, accToken)
	if err != nil {
		return nil, errors.NewEf(err, "could not update access token")
	}
	return accToken, nil
}

func (d *domainI) sendResetPasswordEmail(ctx context.Context, token string, user *entities.User) error {
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

func (d *domainI) sendVerificationEmail(ctx context.Context, token string, user *entities.User) error {
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

func (d *domainI) generateAndSendVerificationToken(ctx context.Context, user *entities.User) error {
	verificationToken := generateId("invite")
	err := d.verifyTokenRepo.SetWithExpiry(
		ctx, verificationToken, &entities.VerifyToken{
			Token:  verificationToken,
			UserId: user.Id,
		}, time.Second*24*60*60,
	)
	if err != nil {
		return errors.NewE(err)
	}
	err = d.sendVerificationEmail(ctx, verificationToken, user)
	if err != nil {
		return errors.NewE(err)
	}
	return nil
}

func fxDomain(
	userRepo repos.DbRepo[*entities.User],
	accessTokenRepo repos.DbRepo[*entities.AccessToken],
	remoteLoginRepo repos.DbRepo[*entities.RemoteLogin],
	verifyTokenRepo kv.Repo[*entities.VerifyToken],
	resetTokenRepo kv.Repo[*entities.ResetPasswordToken],
	inviteCodeRepo repos.DbRepo[*entities.InviteCode],
	github Github,
	gitlab Gitlab,
	google Google,
	logger logging.Logger,
	commsClient comms.CommsClient,
	recaptchaClient *recaptchaenterprise.Client,
	ev *env.Env,
) Domain {
	return &domainI{
		remoteLoginRepo: remoteLoginRepo,
		commsClient:     commsClient,
		userRepo:        userRepo,
		accessTokenRepo: accessTokenRepo,
		verifyTokenRepo: verifyTokenRepo,
		resetTokenRepo:  resetTokenRepo,
		inviteCodeRepo:  inviteCodeRepo,
		github:          github,
		gitlab:          gitlab,
		google:          google,
		logger:          logger,
		recaptchaClient: recaptchaClient,
		envVars:         ev,
	}
}
