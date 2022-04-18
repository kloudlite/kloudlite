package graph

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.

import (
	"context"
	"errors"
	"fmt"

	"kloudlite.io/apps/auth/internal/app/graph/generated"
	"kloudlite.io/apps/auth/internal/app/graph/model"
	"kloudlite.io/common"
	"kloudlite.io/pkg/cache"
	klErrors "kloudlite.io/pkg/errors"
	"kloudlite.io/pkg/repos"
)

func (r *mutationResolver) Login(ctx context.Context, email string, password string) (*model.Session, error) {
	sessionEntity, err := r.d.Login(ctx, email, password)
	if err != nil {
		return nil, err
	}
	cache.SetSession(ctx, sessionEntity)
	return sessionModelFromAuthSession(sessionEntity), err
}

func (r *mutationResolver) Signup(ctx context.Context, name string, email string, password string) (*model.Session, error) {
	sessionEntity, err := r.d.SignUp(ctx, name, email, password)
	if err != nil {
		return nil, err
	}
	cache.SetSession(ctx, sessionEntity)
	session := sessionModelFromAuthSession(sessionEntity)
	return session, err
}

func (r *mutationResolver) Logout(ctx context.Context) (bool, error) {
	session := cache.GetSession[*common.AuthSession](ctx)
	if session == nil {
		return true, nil
	}
	cache.DeleteSession(ctx)
	return true, nil
}

func (r *mutationResolver) SetMetadata(ctx context.Context, values map[string]interface{}) (*model.User, error) {
	session := cache.GetSession[*common.AuthSession](ctx)
	if session == nil {
		return nil, errors.New("user not logged in")
	}
	userEntity, err := r.d.SetUserMetadata(ctx, repos.ID(session.UserId), values)
	return userModelFromEntity(userEntity), err
}

func (r *mutationResolver) ClearMetadata(ctx context.Context) (*model.User, error) {
	session := cache.GetSession[*common.AuthSession](ctx)
	if session == nil {
		return nil, errors.New("user not logged in")
	}
	userEntity, err := r.d.ClearUserMetadata(ctx, repos.ID(session.UserId))
	return userModelFromEntity(userEntity), err
}

func (r *mutationResolver) VerifyEmail(ctx context.Context, token string) (*model.Session, error) {
	sessionEntity, err := r.d.VerifyEmail(ctx, token)
	cache.SetSession(ctx, sessionEntity)
	return sessionModelFromAuthSession(sessionEntity), err
}

func (r *mutationResolver) ResetPassword(ctx context.Context, token string, password string) (bool, error) {
	return r.d.ResetPassword(ctx, token, password)
}

func (r *mutationResolver) RequestResetPassword(ctx context.Context, email string) (bool, error) {
	return r.d.RequestResetPassword(ctx, email)
}

func (r *mutationResolver) LoginWithInviteToken(ctx context.Context, inviteToken string) (*model.Session, error) {
	// TODO
	sessionE, err := r.d.LoginWithInviteToken(ctx, inviteToken)
	return sessionModelFromAuthSession(sessionE), err
}

func (r *mutationResolver) InviteSignup(ctx context.Context, email string, name string) (repos.ID, error) {
	// TODO
	return r.d.InviteUser(ctx, email, name)
}

func (r *mutationResolver) ChangeEmail(ctx context.Context, email string) (bool, error) {
	session := cache.GetSession[*common.AuthSession](ctx)
	if session == nil {
		return false, errors.New("user is not logged in")
	}
	return r.d.ChangeEmail(ctx, repos.ID(session.UserId), email)
}

func (r *mutationResolver) ResendVerificationEmail(ctx context.Context) (bool, error) {
	session := cache.GetSession[*common.AuthSession](ctx)
	if session == nil {
		return false, errors.New("user is not logged in")
	}
	return r.d.ResendVerificationEmail(ctx, repos.ID(session.UserId))
}

func (r *mutationResolver) ChangePassword(ctx context.Context, currentPassword string, newPassword string) (bool, error) {
	session := cache.GetSession[*common.AuthSession](ctx)
	if session == nil {
		return false, errors.New("user is not logged in")
	}
	return r.d.ChangePassword(ctx, repos.ID(session.UserId), currentPassword, newPassword)
}

func (r *mutationResolver) OAuthLogin(ctx context.Context, provider string, code string, state *string) (*model.Session, error) {
	st := ""
	if state != nil {
		st = *state
	}
	sessionEntity, err := r.d.OauthLogin(ctx, provider, st, code)
	if err != nil {
		return nil, klErrors.NewEf(err, "could not create session")
	}
	cache.SetSession(ctx, sessionEntity)
	return sessionModelFromAuthSession(sessionEntity), err
}

func (r *mutationResolver) OAuthAddLogin(ctx context.Context, provider string, state string, code string) (bool, error) {
	panic(fmt.Errorf("not implemented"))
}

func (r *queryResolver) Me(ctx context.Context) (*model.User, error) {
	session := cache.GetSession[*common.AuthSession](ctx)
	if session == nil {
		return nil, errors.New("user not logged in")
	}
	u, err := r.d.GetUserById(ctx, repos.ID(session.UserId))
	if err != nil {
		return nil, err
	}
	return userModelFromEntity(u), err
}

func (r *queryResolver) FindByEmail(ctx context.Context, email string) (*model.User, error) {
	userEntity, err := r.d.GetUserByEmail(ctx, email)
	return userModelFromEntity(userEntity), err
}

func (r *queryResolver) RequestLogin(ctx context.Context, provider string, state *string) (string, error) {
	_state := ""
	if state != nil {
		_state = *state
	}
	url, err := r.d.OauthRequestLogin(ctx, provider, _state)
	if err != nil {
		return "", klErrors.NewE(err)
	}
	return url, nil
}

// Mutation returns generated.MutationResolver implementation.
func (r *Resolver) Mutation() generated.MutationResolver { return &mutationResolver{r} }

// Query returns generated.QueryResolver implementation.
func (r *Resolver) Query() generated.QueryResolver { return &queryResolver{r} }

type mutationResolver struct{ *Resolver }
type queryResolver struct{ *Resolver }

// !!! WARNING !!!
// The code below was going to be deleted when updating resolvers. It has been copied here so you have
// one last chance to move it out of harms way if you want. There are two reasons this happens:
//  - When renaming or deleting a resolver the old code will be put in here. You can safely delete
//    it when you're done.
//  - You have helper methods in this file. Move them out to keep these resolver files clean.
func (r *mutationResolver) OauthAddLogin(ctx context.Context, provider string, state string, code string) (bool, error) {
	session := cache.GetSession[*common.AuthSession](ctx)
	if session == nil {
		return false, errors.New("user is not logged in")
	}
	return r.d.OauthAddLogin(ctx, repos.ID(session.UserId), provider, state, code)
}
