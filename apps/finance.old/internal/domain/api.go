package domain

import (
	"context"
	"kloudlite.io/constants"
	"time"

	"kloudlite.io/pkg/repos"
)

type Domain interface {
	// CRUD

	CreateAccount(ctx context.Context, userId repos.ID, name string, billing Billing) (*Account, error)
	GetAccount(ctx context.Context, id repos.ID) (*Account, error)
	UpdateAccount(ctx context.Context, id repos.ID, name *string, email *string) (*Account, error)
	DeleteAccount(ctx context.Context, id repos.ID) (bool, error)

	DeactivateAccount(ctx context.Context, id repos.ID) (bool, error)
	ActivateAccount(ctx context.Context, id repos.ID) (bool, error)

	// Membership

	AddAccountMember(ctx context.Context, accountId repos.ID, email string, role constants.Role) (bool, error)
	RemoveAccountMember(ctx context.Context, accountId repos.ID, userId repos.ID) (bool, error)
	UpdateAccountMember(ctx context.Context, id repos.ID, id2 repos.ID, role string) (bool, error)

	// Billing

	UpdateAccountBilling(ctx context.Context, id repos.ID, d *Billing) (*Account, error)
	GetOutstandingAmount(ctx context.Context, accountId repos.ID) (float64, error)

	GetAccountMemberships(ctx context.Context, userId repos.ID) ([]*Membership, error)
	GetAccountMembership(ctx context.Context, userId repos.ID, accountId repos.ID) (*Membership, error)
	GetUserMemberships(ctx context.Context, resourceId repos.ID) ([]*Membership, error)
	GetComputePlanByName(ctx context.Context, name string) (*ComputePlan, error)
	GetLambdaPlanByName(ctx context.Context, name string) (*LamdaPlan, error)
	GenerateBillingInvoice(ctx context.Context, accountId repos.ID) (*BillingInvoice, error)
	TriggerBillingEvent(
		ctx context.Context,
		accountId repos.ID,
		resourceId repos.ID,
		projectId repos.ID,
		eventType string,
		billables []Billable,
		timeStamp time.Time,
	) error
	GetStoragePlanByName(ctx context.Context, name string) (*StoragePlan, error)
	GetSetupIntent(ctx context.Context) (string, error)
	Test(ctx context.Context, accountId repos.ID) error
	AttachToCluster(ctx context.Context, accountId repos.ID, clusterId repos.ID) (bool, error)
}
