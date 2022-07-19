package graph

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.

import (
	"context"
	"errors"
	"kloudlite.io/apps/finance/internal/app/graph/generated"
	"kloudlite.io/apps/finance/internal/app/graph/model"
	"kloudlite.io/apps/finance/internal/domain"
	"kloudlite.io/common"
	httpServer "kloudlite.io/pkg/http-server"
	"kloudlite.io/pkg/repos"
)

func (r *accountResolver) Memberships(ctx context.Context, obj *model.Account) ([]*model.AccountMembership, error) {
	entities, err := r.domain.GetUserMemberships(ctx, obj.ID)
	accountMemberships := make([]*model.AccountMembership, len(entities))
	for i, entity := range entities {
		accountMemberships[i] = &model.AccountMembership{
			Account: &model.Account{
				ID: entity.AccountId,
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
	amount, err := r.domain.GetOutstandingAmount(ctx, obj.ID)
	return amount, err
}

func (r *accountMembershipResolver) User(ctx context.Context, obj *model.AccountMembership) (*model.User, error) {
	return &model.User{
		ID: obj.User.ID,
	}, nil
}

func (r *accountMembershipResolver) Account(ctx context.Context, obj *model.AccountMembership) (*model.Account, error) {
	ae, err := r.domain.GetAccount(ctx, obj.Account.ID)
	if err != nil {
		return nil, err
	}
	if ae == nil {
		return nil, errors.New("account not found")
	}
	return AccountModelFromEntity(ae), nil
}

func (r *mutationResolver) FinanceCreateAccount(ctx context.Context, name string, billing model.BillingInput) (*model.Account, error) {
	session := httpServer.GetSession[*common.AuthSession](ctx)
	if session == nil {
		return nil, errors.New("not logged in")
	}
	account, err := r.domain.CreateAccount(ctx, session.UserId, name, domain.Billing{
		PaymentMethodId: billing.StripePaymentMethodID,
		CardholderName:  billing.CardholderName,
		Address:         billing.Address,
	})
	if err != nil {
		return nil, err
	}
	return AccountModelFromEntity(account), nil
}

func (r *mutationResolver) FinanceUpdateAccount(ctx context.Context, accountID repos.ID, name *string, contactEmail *string) (*model.Account, error) {
	session := httpServer.GetSession[*common.AuthSession](ctx)
	if session == nil {
		return nil, errors.New("not logged in")
	}
	account, err := r.domain.UpdateAccount(ctx, accountID, name, contactEmail)
	if err != nil {
		return nil, err
	}
	return AccountModelFromEntity(account), nil
}

func (r *mutationResolver) FinanceUpdateAccountBilling(ctx context.Context, accountID repos.ID, billing model.BillingInput) (*model.Account, error) {
	session := httpServer.GetSession[*common.AuthSession](ctx)
	if session == nil {
		return nil, errors.New("not logged in")
	}
	account, err := r.domain.UpdateAccountBilling(ctx, accountID, &domain.Billing{
		PaymentMethodId: billing.StripePaymentMethodID,
		CardholderName:  billing.CardholderName,
		Address:         billing.Address,
	})
	if err != nil {
		return nil, err
	}
	return AccountModelFromEntity(account), nil
}

func (r *mutationResolver) FinanceInviteAccountMember(ctx context.Context, accountID string, name *string, email string, role string) (bool, error) {
	session := httpServer.GetSession[*common.AuthSession](ctx)
	if session == nil {
		return false, errors.New("not logged in")
	}
	return r.domain.AddAccountMember(ctx, repos.ID(accountID), email, common.Role(role))
}

func (r *mutationResolver) FinanceRemoveAccountMember(ctx context.Context, accountID repos.ID, userID repos.ID) (bool, error) {
	session := httpServer.GetSession[*common.AuthSession](ctx)
	if session == nil {
		return false, errors.New("not logged in")
	}
	return r.domain.RemoveAccountMember(ctx, accountID, userID)
}

func (r *mutationResolver) FinanceUpdateAccountMember(ctx context.Context, accountID repos.ID, userID repos.ID, role string) (bool, error) {
	session := httpServer.GetSession[*common.AuthSession](ctx)
	if session == nil {
		return false, errors.New("not logged in")
	}
	return r.domain.UpdateAccountMember(ctx, accountID, userID, role)
}

func (r *mutationResolver) FinanceDeactivateAccount(ctx context.Context, accountID repos.ID) (bool, error) {
	session := httpServer.GetSession[*common.AuthSession](ctx)
	if session == nil {
		return false, errors.New("not logged in")
	}
	return r.domain.DeactivateAccount(ctx, accountID)
}

func (r *mutationResolver) FinanceActivateAccount(ctx context.Context, accountID repos.ID) (bool, error) {
	session := httpServer.GetSession[*common.AuthSession](ctx)
	if session == nil {
		return false, errors.New("not logged in")
	}
	return r.domain.ActivateAccount(ctx, accountID)
}

func (r *mutationResolver) FinanceDeleteAccount(ctx context.Context, accountID repos.ID) (bool, error) {
	session := httpServer.GetSession[*common.AuthSession](ctx)
	if session == nil {
		return false, errors.New("not logged in")
	}
	return r.domain.DeleteAccount(ctx, accountID)
}

func (r *queryResolver) FinanceAccount(ctx context.Context, accountID repos.ID) (*model.Account, error) {
	session := httpServer.GetSession[*common.AuthSession](ctx)
	if session == nil {
		return nil, errors.New("not logged in")
	}
	accountEntity, err := r.domain.GetAccount(ctx, accountID)
	return AccountModelFromEntity(accountEntity), err
}

func (r *queryResolver) FinanceStripeSetupIntent(ctx context.Context) (*string, error) {
	intent, err := r.domain.GetSetupIntent(ctx)
	if err != nil {
		return nil, err
	}
	return &intent, nil
}

func (r *queryResolver) FinanceTestStripe(ctx context.Context, accountID repos.ID) (bool, error) {
	err := r.domain.Test(ctx, accountID)
	if err != nil {
		return false, err
	}
	return true, nil
}

func (r *userResolver) AccountMemberships(ctx context.Context, obj *model.User) ([]*model.AccountMembership, error) {
	entities, err := r.domain.GetAccountMemberships(ctx, obj.ID)
	accountMemberships := make([]*model.AccountMembership, len(entities))
	for i, entity := range entities {
		accountMemberships[i] = &model.AccountMembership{
			Account: &model.Account{
				ID: entity.AccountId,
			},
			User: &model.User{
				ID: entity.UserId,
			},
			Role: string(entity.Role),
		}
	}
	return accountMemberships, err
}

func (r *userResolver) AccountMembership(ctx context.Context, obj *model.User, accountID *repos.ID) (*model.AccountMembership, error) {
	membership, err := r.domain.GetAccountMembership(ctx, obj.ID, *accountID)
	if err != nil {
		return nil, err
	}
	return &model.AccountMembership{
		Account: &model.Account{
			ID: membership.AccountId,
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
