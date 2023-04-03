package graph

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.

import (
	"context"
	"errors"
	"fmt"

	"kloudlite.io/apps/finance/internal/app/graph/generated"
	"kloudlite.io/apps/finance/internal/app/graph/model"
	iamT "kloudlite.io/apps/iam/types"
	"kloudlite.io/pkg/repos"
)

func (r *accountResolver) Memberships(ctx context.Context, obj *model.Account) ([]*model.AccountMembership, error) {
	entities, err := r.domain.GetUserMemberships(toFinanceContext(ctx), iamT.NewResourceRef(obj.Name, iamT.ResourceAccount, obj.Name))
	accountMemberships := make([]*model.AccountMembership, len(entities))
	for i, entity := range entities {
		accountMemberships[i] = &model.AccountMembership{
			Account: &model.Account{
				Name: entity.AccountName,
			},
			User: &model.User{
				ID: entity.UserId,
			},
			Role:     string(entity.Role),
			Accepted: entity.Accepted,
		}
	}
	return accountMemberships, err
}

func (r *accountResolver) OutstandingAmount(ctx context.Context, obj *model.Account) (float64, error) {
	panic(fmt.Errorf("not implemented"))
}

func (r *accountMembershipResolver) User(ctx context.Context, obj *model.AccountMembership) (*model.User, error) {
	return &model.User{
		ID: obj.User.ID,
	}, nil
}

func (r *accountMembershipResolver) Account(ctx context.Context, obj *model.AccountMembership) (*model.Account, error) {
	acc, err := r.domain.GetAccount(toFinanceContext(ctx), obj.Account.Name)
	if err != nil {
		return nil, err
	}
	if acc == nil {
		return nil, errors.New("account not found")
	}
	return AccountModelFromEntity(acc), nil
}

func (r *mutationResolver) FinanceCreateAccount(ctx context.Context, name string, billing model.BillingInput) (*model.Account, error) {
	account, err := r.domain.CreateAccount(toFinanceContext(ctx), name)
	if err != nil {
		return nil, err
	}

	return AccountModelFromEntity(account), nil
}

func (r *mutationResolver) FinanceUpdateAccount(ctx context.Context, accountName string, name *string, contactEmail *string) (*model.Account, error) {
	account, err := r.domain.UpdateAccount(toFinanceContext(ctx), accountName, contactEmail)
	if err != nil {
		return nil, err
	}
	return AccountModelFromEntity(account), nil
}

func (r *mutationResolver) FinanceInviteAccountMember(ctx context.Context, accountName string, name *string, email string, role string) (bool, error) {
	return r.domain.AddAccountMember(toFinanceContext(ctx), accountName, email, iamT.Role(role))
}

func (r *mutationResolver) FinanceRemoveAccountMember(ctx context.Context, accountName string, userID repos.ID) (bool, error) {
	return r.domain.RemoveAccountMember(toFinanceContext(ctx), accountName, userID)
}

func (r *mutationResolver) FinanceUpdateAccountMember(ctx context.Context, accountName string, userID repos.ID, role string) (bool, error) {
	return r.domain.UpdateAccountMember(toFinanceContext(ctx), accountName, userID, role)
}

func (r *mutationResolver) FinanceDeactivateAccount(ctx context.Context, accountName string) (bool, error) {
	return r.domain.DeactivateAccount(toFinanceContext(ctx), accountName)
}

func (r *mutationResolver) FinanceActivateAccount(ctx context.Context, accountName string) (bool, error) {
	return r.domain.ActivateAccount(toFinanceContext(ctx), accountName)
}

func (r *mutationResolver) FinanceDeleteAccount(ctx context.Context, accountName string) (bool, error) {
	return r.domain.DeleteAccount(toFinanceContext(ctx), accountName)
}

func (r *queryResolver) FinanceAccount(ctx context.Context, accountName string) (*model.Account, error) {
	accountEntity, err := r.domain.GetAccount(toFinanceContext(ctx), accountName)
	return AccountModelFromEntity(accountEntity), err
}

func (r *userResolver) AccountMemberships(ctx context.Context, obj *model.User) ([]*model.AccountMembership, error) {
	entities, err := r.domain.GetAccountMemberships(toFinanceContext(ctx), obj.ID)
	accountMemberships := make([]*model.AccountMembership, len(entities))
	for i, entity := range entities {
		accountMemberships[i] = &model.AccountMembership{
			Account: &model.Account{
				Name: entity.AccountName,
			},
			User: &model.User{
				ID: entity.UserId,
			},
			Role: string(entity.Role),
		}
	}
	return accountMemberships, err
}

func (r *userResolver) AccountMembership(ctx context.Context, obj *model.User, accountName string) (*model.AccountMembership, error) {
	membership, err := r.domain.GetAccountMembership(toFinanceContext(ctx), obj.ID, accountName)
	if err != nil {
		return nil, err
	}
	return &model.AccountMembership{
		Account: &model.Account{
			Name: membership.AccountName,
		},
		User: &model.User{
			ID: membership.UserId,
		},
		Role: string(membership.Role),
	}, nil
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
