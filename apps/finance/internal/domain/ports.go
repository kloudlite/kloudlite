package domain

import (
	"context"
	"kloudlite.io/common"
	"kloudlite.io/pkg/repos"
)

type Domain interface {
	CreateAccount(
		ctx context.Context,
		userId repos.ID,
		name string,
		billing Billing,
		initialProvider string,
		initialRegion string,
	) (*Account, error)
	UpdateAccount(ctx context.Context, id repos.ID, name *string, email *string) (*Account, error)
	UpdateAccountBilling(ctx context.Context, id repos.ID, d *Billing) (*Account, error)
	AddAccountMember(
		ctx context.Context,
		accountId repos.ID,
		email string,
		role common.Role,
	) (bool, error)
	ConfirmAccountMembership(
		ctx context.Context,
		invitationToken string,
	) (bool, error)
	RemoveAccountMember(
		ctx context.Context,
		accountId repos.ID,
		userId repos.ID,
	) (bool, error)
	UpdateAccountMember(ctx context.Context, id repos.ID, id2 repos.ID, role string) (bool, error)
	DeactivateAccount(ctx context.Context, id repos.ID) (bool, error)
	ActivateAccount(ctx context.Context, id repos.ID) (bool, error)
	DeleteAccount(ctx context.Context, id repos.ID) (bool, error)
	GetAccount(ctx context.Context, id repos.ID) (*Account, error)
	GetAccountMemberships(ctx context.Context, userId repos.ID) ([]*Membership, error)
	GetAccountMembership(ctx context.Context, userId repos.ID, accountId repos.ID) (*Membership, error)
	GetUserMemberships(ctx context.Context, resourceId repos.ID) ([]*Membership, error)

	StartBillable(
		ctx context.Context,
		accountId repos.ID,
		resourceType string,
		quantity float32,
	) (*Billable, error)
	StopBillable(ctx context.Context, billableId repos.ID) error
}
