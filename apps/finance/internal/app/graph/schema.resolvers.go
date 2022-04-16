package graph

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.

import (
	"context"
	"errors"
	"fmt"

	"kloudlite.io/apps/finance/internal/app/graph/generated"
	"kloudlite.io/apps/finance/internal/app/graph/model"
	"kloudlite.io/apps/finance/internal/domain"
	"kloudlite.io/common"
	"kloudlite.io/pkg/cache"
	"kloudlite.io/pkg/repos"
)

func (r *accountResolver) Memberships(ctx context.Context, obj *model.Account) ([]*model.Membership, error) {
	panic(fmt.Errorf("not implemented"))
}

func (r *membershipResolver) User(ctx context.Context, obj *model.Membership) (*model.User, error) {
	panic(fmt.Errorf("not implemented"))
}

func (r *membershipResolver) Account(ctx context.Context, obj *model.Membership) (*model.Account, error) {
	panic(fmt.Errorf("not implemented"))
}

func (r *mutationResolver) CreateAccount(ctx context.Context, name string, billing *model.BillingInput) (*model.Account, error) {
	session := cache.GetSession[*common.AuthSession](ctx)
	if session == nil {
		return nil, errors.New("not logged in")
	}
	account, err := r.domain.CreateAccount(ctx, name, billing)
	if err != nil {
		return nil, err
	}
	return AccountModelFromEntity(account), nil
}

func (r *mutationResolver) UpdateAccount(ctx context.Context, accountID repos.ID, name *string, contactEmail *string) (*model.Account, error) {
	session := cache.GetSession[*common.AuthSession](ctx)
	if session == nil {
		return nil, errors.New("not logged in")
	}
	account, err := r.domain.UpdateAccount(ctx, accountID, name, contactEmail)
	if err != nil {
		return nil, err
	}
	return AccountModelFromEntity(account), nil
}

func (r *mutationResolver) UpdateAccountBilling(ctx context.Context, accountID repos.ID, billing model.BillingInput) (*model.Account, error) {
	session := cache.GetSession[*common.AuthSession](ctx)
	if session == nil {
		return nil, errors.New("not logged in")
	}
	account, err := r.domain.UpdateAccountBilling(ctx, accountID, &domain.Billing{
		StripeSetupIntentId: billing.StripeSetupIntentID,
		CardholderName:      billing.CardholderName,
		Address:             billing.Address,
	})
	if err != nil {
		return nil, err
	}
	return AccountModelFromEntity(account), nil
}

func (r *mutationResolver) InviteAccountMember(ctx context.Context, accountID string, email string, name string, role string) (bool, error) {
	session := cache.GetSession[*common.AuthSession](ctx)
	if session == nil {
		return false, errors.New("not logged in")
	}
	return r.domain.InviteAccountMember(ctx, accountID, email, name, role)
}

func (r *mutationResolver) RemoveAccountMember(ctx context.Context, accountID repos.ID, userID repos.ID) (bool, error) {
	session := cache.GetSession[*common.AuthSession](ctx)
	if session == nil {
		return false, errors.New("not logged in")
	}
	return r.domain.RemoveAccountMember(ctx, accountID, userID)
}

func (r *mutationResolver) UpdateAccountMember(ctx context.Context, accountID repos.ID, userID repos.ID, role string) (bool, error) {
	session := cache.GetSession[*common.AuthSession](ctx)
	if session == nil {
		return false, errors.New("not logged in")
	}
	return r.domain.UpdateAccountMember(ctx, accountID, userID, role)
}

func (r *mutationResolver) DeactivateAccount(ctx context.Context, accountID repos.ID) (bool, error) {
	session := cache.GetSession[*common.AuthSession](ctx)
	if session == nil {
		return false, errors.New("not logged in")
	}
	return r.domain.DeactivateAccount(ctx, accountID)
}

func (r *mutationResolver) ActivateAccount(ctx context.Context, accountID repos.ID) (bool, error) {
	session := cache.GetSession[*common.AuthSession](ctx)
	if session == nil {
		return false, errors.New("not logged in")
	}
	return r.domain.ActivateAccount(ctx, accountID)
}

func (r *mutationResolver) DeleteAccount(ctx context.Context, accountID repos.ID) (bool, error) {
	session := cache.GetSession[*common.AuthSession](ctx)
	if session == nil {
		return false, errors.New("not logged in")
	}
	return r.domain.DeleteAccount(ctx, accountID)
}

func (r *queryResolver) Accounts(ctx context.Context) ([]*model.Account, error) {
	session := cache.GetSession[*common.AuthSession](ctx)
	if session == nil {
		return nil, errors.New("not logged in")
	}
	accountEntities, err := r.domain.ListAccounts(ctx, repos.ID(session.UserId))
	if err != nil {
		return nil, err
	}
	accountModels := make([]*model.Account, 0)
	for _, ae := range accountEntities {
		accountModels = append(accountModels, AccountModelFromEntity(ae))
	}
	return accountModels, nil
}

func (r *queryResolver) Account(ctx context.Context, accountID repos.ID) (*model.Account, error) {
	session := cache.GetSession[*common.AuthSession](ctx)
	if session == nil {
		return nil, errors.New("not logged in")
	}
	accountEntity, err := r.domain.GetAccount(ctx, accountID)
	return AccountModelFromEntity(accountEntity), err
}

func (r *queryResolver) AccountsMembership(ctx context.Context) ([]*model.AccountMembership, error) {
	panic(fmt.Errorf("not implemented"))
}

func (r *queryResolver) AccountMembership(ctx context.Context, accountID repos.ID) (*model.AccountMembership, error) {
	panic(fmt.Errorf("not implemented"))
}

func (r *queryResolver) StripeSetupIntent(ctx context.Context) (string, error) {
	panic(fmt.Errorf("not implemented"))
}

func (r *userResolver) Memberships(ctx context.Context, obj *model.User) ([]*model.Membership, error) {
	panic(fmt.Errorf("not implemented"))
}

// Account returns generated.AccountResolver implementation.
func (r *Resolver) Account() generated.AccountResolver { return &accountResolver{r} }

// Membership returns generated.MembershipResolver implementation.
func (r *Resolver) Membership() generated.MembershipResolver { return &membershipResolver{r} }

// Mutation returns generated.MutationResolver implementation.
func (r *Resolver) Mutation() generated.MutationResolver { return &mutationResolver{r} }

// Query returns generated.QueryResolver implementation.
func (r *Resolver) Query() generated.QueryResolver { return &queryResolver{r} }

// User returns generated.UserResolver implementation.
func (r *Resolver) User() generated.UserResolver { return &userResolver{r} }

type accountResolver struct{ *Resolver }
type membershipResolver struct{ *Resolver }
type mutationResolver struct{ *Resolver }
type queryResolver struct{ *Resolver }
type userResolver struct{ *Resolver }
