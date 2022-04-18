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

func (r *accountResolver) Memberships(ctx context.Context, obj *model.Account) ([]*model.AccountMembership, error) {
	me,err:=r.domain.GetAccountMemberShips(ctx, obj.ID)

	if err != nil {
		return nil, err
	}
	fmt.Println(me)
	return , nil
}

func (r *accountMembershipResolver) User(ctx context.Context, obj *model.AccountMembership) (*model.User, error) {
	panic(fmt.Errorf("not implemented"))
}

func (r *accountMembershipResolver) Account(ctx context.Context, obj *model.AccountMembership) (*model.Account, error) {
	ae, err := r.domain.GetAccount(ctx, obj.Account.ID)

	if err != nil {
		return nil, err
	}

	return AccountModelFromEntity(ae), nil
}

func (r *mutationResolver) CreateAccount(ctx context.Context, name string, billing *model.BillingInput) (*model.Account, error) {
	fmt.Println("create account")
	session := cache.GetSession[*common.AuthSession](ctx)
	if session == nil {
		return nil, errors.New("not logged in")
	}
	account, err := r.domain.CreateAccount(ctx, repos.ID(session.UserId), name, billing)
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

func (r *mutationResolver) AddAccountMember(ctx context.Context, accountID string, email string, name string, role string) (bool, error) {
	session := cache.GetSession[*common.AuthSession](ctx)
	if session == nil {
		return false, errors.New("not logged in")
	}

	return r.domain.AddAccountMember(ctx, accountID, email, name, role)
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

func (r *queryResolver) Account(ctx context.Context, accountID repos.ID) (*model.Account, error) {
	session := cache.GetSession[*common.AuthSession](ctx)
	if session == nil {
		return nil, errors.New("not logged in")
	}
	accountEntity, err := r.domain.GetAccount(ctx, accountID)
	return AccountModelFromEntity(accountEntity), err
}

func (r *userResolver) AccountMemberships(ctx context.Context, obj *model.User) ([]*model.AccountMembership, error) {
	entities, err := r.domain.GetAccountMemberShips(ctx, obj.ID)
	fmt.Println(entities, err, "entities")
	accountMemeberships := make([]*model.AccountMembership, len(entities))

	for i, entity := range entities {
		accountMemeberships[i] = &model.AccountMembership{
			Account: &model.Account{
				ID: entity.AccountId,
			},
			User: &model.User{
				ID: entity.UserId,
			},
			Role: string(entity.Role),
		}
	}

	return accountMemeberships, err
	// panic(fmt.Errorf("not implemented 1"))
}

// Account returns generated.AccountResolver implementation.
func (r *Resolver) Account() generated.AccountResolver { return &accountResolver{r} }

// AccountMembership returns generated.AccountMembershipResolver implementation.
func (r *Resolver) AccountMembership() generated.AccountMembershipResolver {
	return &accountMembershipResolver{r}
}

// Mutation returns generated.MutationResolver implementation.
func (r *Resolver) Mutation() generated.MutationResolver { return &mutationResolver{r} }

// Query returns generated.QueryResolver implementation.
func (r *Resolver) Query() generated.QueryResolver { return &queryResolver{r} }

// User returns generated.UserResolver implementation.
func (r *Resolver) User() generated.UserResolver { return &userResolver{r} }

type accountResolver struct{ *Resolver }
type accountMembershipResolver struct{ *Resolver }
type mutationResolver struct{ *Resolver }
type queryResolver struct{ *Resolver }
type userResolver struct{ *Resolver }
