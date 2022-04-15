package graph

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.

import (
	"context"
	"kloudlite.io/pkg/cache"

	"kloudlite.io/apps/auth/internal/app/graph/generated"
	"kloudlite.io/apps/auth/internal/app/graph/model"
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

func (r *mutationResolver) InviteSignup(ctx context.Context, email string, name string) (repos.ID, error) {
	return r.d.InviteUser(ctx, email, name)
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
	userId := ctx.Value("user_id").(repos.ID)
	return r.d.Logout(ctx, userId)
}

func (r *mutationResolver) SetMetadata(ctx context.Context, values map[string]interface{}) (*model.User, error) {
	userId := ctx.Value("user_id").(repos.ID)
	userEntity, err := r.d.SetUserMetadata(ctx, userId, values)
	return userModelFromEntity(userEntity), err
}

func (r *mutationResolver) ClearMetadata(ctx context.Context) (*model.User, error) {
	userId := ctx.Value("user_id").(repos.ID)
	userEntity, err := r.d.ClearUserMetadata(ctx, userId)
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
	sessionE, err := r.d.LoginWithInviteToken(ctx, inviteToken)
	return sessionModelFromAuthSession(sessionE), err
}

func (r *mutationResolver) ChangeEmail(ctx context.Context, email string) (bool, error) {
	userId := ctx.Value("user_id").(repos.ID)
	return r.d.ChangeEmail(ctx, userId, email)
}

func (r *mutationResolver) ResendVerificationEmail(ctx context.Context, email string) (bool, error) {
	return r.d.ResendVerificationEmail(ctx, email)
}

func (r *mutationResolver) VerifyChangeEmail(ctx context.Context, token string) (bool, error) {
	return r.d.VerifyChangeEmail(ctx, token)
}

func (r *mutationResolver) ChangePassword(ctx context.Context, currentPassword string, newPassword string) (bool, error) {
	userId := ctx.Value("user_id").(repos.ID)
	return r.d.ChangePassword(ctx, userId, currentPassword, newPassword)
}

func (r *mutationResolver) OauthLogin(ctx context.Context, provider string, state string, code string) (*model.Session, error) {
	sessionEntity, err := r.d.OauthLogin(ctx, provider, state, code)
	cache.SetSession(ctx, sessionEntity)
	return sessionModelFromAuthSession(sessionEntity), err
}

func (r *mutationResolver) OauthAddLogin(ctx context.Context, provider string, state string, code string) (bool, error) {
	userId := ctx.Value("user_id").(repos.ID)
	return r.d.OauthAddLogin(ctx, userId, provider, state, code)
}

func (r *queryResolver) Me(ctx context.Context) (*model.User, error) {
	userId := ctx.Value("user_id").(repos.ID)
	userEntity, err := r.d.GetUserById(ctx, userId)
	return userModelFromEntity(userEntity), err
}

func (r *queryResolver) FindByEmail(ctx context.Context, email string) (*model.User, error) {
	userEntity, err := r.d.GetUserByEmail(ctx, email)
	return userModelFromEntity(userEntity), err
}

func (r *queryResolver) RequestLogin(ctx context.Context, provider string, state *string) (string, error) {
	return r.d.GetLoginDetails(ctx, provider, state)
}

// Mutation returns generated.MutationResolver implementation.
func (r *Resolver) Mutation() generated.MutationResolver { return &mutationResolver{r} }

// Query returns generated.QueryResolver implementation.
func (r *Resolver) Query() generated.QueryResolver { return &queryResolver{r} }

type mutationResolver struct{ *Resolver }
type queryResolver struct{ *Resolver }
