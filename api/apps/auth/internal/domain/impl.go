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
	remoteLoginRepo repos.DbRepo[*entities.RemoteLogin]

	envVars *env.AuthEnv
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

func (d *domainI) GetUserById(ctx context.Context, id repos.ID) (*entities.User, error) {
	return d.userRepo.FindById(ctx, id)
}

func (d *domainI) GetUserByEmail(ctx context.Context, email string) (*entities.User, error) {
	return d.userRepo.FindOne(ctx, repos.Filter{"email": email})
}

func (d *domainI) LoginWithSSO(ctx context.Context, email string, name string) (*entities.User, error) {
	user, err := d.userRepo.FindOne(ctx, repos.Filter{"email": email})
	if err != nil {
		return nil, errors.NewE(err)
	}

	if user == nil {
		user, err = d.userRepo.Create(ctx, &entities.User{
			Email:    email,
			Name:     name,
			Verified: true,
		})
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
	}

	return user, nil
}

func (d *domainI) LoginWithOAuth(ctx context.Context, email string, name string) (*entities.User, error) {
	user, err := d.userRepo.FindOne(ctx, repos.Filter{"email": email})
	if err != nil {
		return nil, errors.NewE(err)
	}

	if user == nil {
		user, err = d.userRepo.Create(ctx, &entities.User{
			Email:    email,
			Name:     name,
			Verified: true,
		})
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
	}

	return user, nil
}

func (d *domainI) Login(ctx context.Context, email string, password string) (*entities.User, error) {
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

	return user, nil
}

func (d *domainI) MachineLogin(ctx context.Context, userId string, machineId string, cluster string) (*common.AuthSession, error) {
	user, err := d.userRepo.FindOne(ctx, repos.Filter{"id": userId})
	if err != nil {
		return nil, errors.NewE(err)
	}
	session := newAuthSession(user.Id, user.Email, user.Name, user.Verified, "work_machine")
	session.Extras = map[string]any{}
	session.Extras[common.MACHINE_ID_KEY] = machineId
	session.Extras[common.CLUSTER_KEY] = cluster
	return session, nil
}

func (d *domainI) SignUp(ctx context.Context, name string, email string, password string) (*entities.User, error) {
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
			Verified:     !d.envVars.UserEmailVerificationEnabled,
			Approved:     false,
			Metadata:     nil,
			Joined:       time.Now(),
			PasswordSalt: salt,
		},
	)
	if err != nil {
		return nil, errors.NewE(err)
	}

	if d.envVars.UserEmailVerificationEnabled {
		err = d.generateAndSendVerificationToken(ctx, user)
		if err != nil {
			return nil, errors.NewE(err)
		}
	}

	return user, nil
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

func (d *domainI) RequestResetPassword(ctx context.Context, email string) (bool, error) {
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

func (d *domainI) ResendVerificationEmail(ctx context.Context, email string) (bool, error) {
	user, err := d.userRepo.FindOne(ctx, repos.Filter{"email": email})
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

func (d *domainI) addOAuthLogin(ctx context.Context, provider string, token *oauth2.Token, u *entities.User, avatarUrl *string) (*entities.User, error) {
	user, err := d.userRepo.FindOne(ctx, repos.Filter{"email": u.Email})
	if err != nil {
		return nil, errors.NewEf(err, "could not find user")
	}
	if user == nil {
		user = u
		user.Joined = time.Now()
		user, err = d.userRepo.Create(ctx, user)
		if _, err := d.commsClient.SendWelcomeEmail(
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

func (gl *domainI) Hash(t *oauth2.Token) (string, error) {
	b, err := json.Marshal(t)
	if err != nil {
		return "", errors.NewE(err)
	}
	return b64.StdEncoding.EncodeToString(b), nil
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
	commsClient comms.CommsClient,
	ev *env.AuthEnv,
) Domain {
	return &domainI{
		remoteLoginRepo: remoteLoginRepo,
		commsClient:     commsClient,
		userRepo:        userRepo,
		accessTokenRepo: accessTokenRepo,
		verifyTokenRepo: verifyTokenRepo,
		resetTokenRepo:  resetTokenRepo,
		inviteCodeRepo:  inviteCodeRepo,
		envVars:         ev,
	}
}
