package domain

import (
	"context"
	"kloudlite.io/apps/finance/internal/app/graph/model"
	"kloudlite.io/pkg/repos"
)

type Domain interface {
	CreateAccount(
		ctx context.Context,
		userId repos.ID,
		name string,
		billing *model.BillingInput,
	) (*Account, error)
	UpdateAccount(ctx context.Context, id repos.ID, name *string, email *string) (*Account, error)
	UpdateAccountBilling(ctx context.Context, id repos.ID, d *Billing) (*Account, error)
	AddAccountMember(ctx context.Context, id string, email string, name string, role string) (bool, error)
	RemoveAccountMember(ctx context.Context, id repos.ID, id2 repos.ID) (bool, error)
	UpdateAccountMember(ctx context.Context, id repos.ID, id2 repos.ID, role string) (bool, error)
	DeactivateAccount(ctx context.Context, id repos.ID) (bool, error)
	ActivateAccount(ctx context.Context, id repos.ID) (bool, error)
	DeleteAccount(ctx context.Context, id repos.ID) (bool, error)
	ListAccounts(ctx context.Context, id repos.ID) ([]*Account, error)
	GetAccount(ctx context.Context, id repos.ID) (*Account, error)
	GetAccountMemberships(ctx context.Context, userId repos.ID) ([]*Membership, error)
	GetUserMemberships(ctx context.Context, resourceId repos.ID) ([]*Membership, error)
}
