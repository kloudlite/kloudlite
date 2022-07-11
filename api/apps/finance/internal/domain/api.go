package domain

import (
	"context"
	"kloudlite.io/common"
	"kloudlite.io/pkg/repos"
	"time"
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
	GetAccount(ctx context.Context, id repos.ID) (*Account, error)
	UpdateAccount(ctx context.Context, id repos.ID, name *string, email *string) (*Account, error)
	DeleteAccount(ctx context.Context, id repos.ID) (bool, error)

	DeactivateAccount(ctx context.Context, id repos.ID) (bool, error)
	ActivateAccount(ctx context.Context, id repos.ID) (bool, error)

	UpdateAccountBilling(ctx context.Context, id repos.ID, d *Billing) (*Account, error)

	AddAccountMember(ctx context.Context, accountId repos.ID, email string, role common.Role) (bool, error)
	RemoveAccountMember(ctx context.Context, accountId repos.ID, userId repos.ID) (bool, error)
	UpdateAccountMember(ctx context.Context, id repos.ID, id2 repos.ID, role string) (bool, error)

	GetAccountMemberships(ctx context.Context, userId repos.ID) ([]*Membership, error)
	GetAccountMembership(ctx context.Context, userId repos.ID, accountId repos.ID) (*Membership, error)
	GetUserMemberships(ctx context.Context, resourceId repos.ID) ([]*Membership, error)
	GetComputePlanByName(ctx context.Context, name string) (*ComputePlan, error)
	GetLambdaPlanByName(ctx context.Context, name string) (*LamdaPlan, error)
	TriggerBillingEvent(
		ctx context.Context,
		accountId repos.ID,
		resourceId repos.ID,
		projectId repos.ID,
		eventType string,
		billables []Billable,
		timeStamp time.Time,
	) error
}
