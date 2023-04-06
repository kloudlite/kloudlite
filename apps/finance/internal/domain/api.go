package domain

import (
	"context"

	iamT "kloudlite.io/apps/iam/types"
	"kloudlite.io/pkg/repos"
)

type FinanceContext struct {
	context.Context
	UserId repos.ID
}

type Domain interface {
	// CRUD
	CreateAccount(ctx FinanceContext, name string, displayName string) (*Account, error)
	ListAccounts(ctx FinanceContext) ([]*Account, error)
	GetAccount(ctx FinanceContext, name string) (*Account, error)
	UpdateAccount(ctx FinanceContext, name string, email *string) (*Account, error)
	DeleteAccount(ctx FinanceContext, name string) (bool, error)

	DeactivateAccount(ctx FinanceContext, name string) (bool, error)
	ActivateAccount(ctx FinanceContext, name string) (bool, error)

	// invitations
	InviteUser(ctx FinanceContext, accountName string, email string, role iamT.Role) (bool, error)
	ListInvitations(ctx FinanceContext, accountName string) ([]*Membership, error)
	DeleteInvitation(ctx FinanceContext, email string) (bool, error)

	// Memberships
	RemoveAccountMember(ctx FinanceContext, accountName string, userId repos.ID) (bool, error)
	UpdateAccountMember(ctx FinanceContext, accountName string, userId repos.ID, role string) (bool, error)

	GetUserMemberships(ctx FinanceContext, resourceRef string) ([]*Membership, error)
	GetAccountMemberships(ctx FinanceContext) ([]*Membership, error)
	GetAccountMembership(ctx FinanceContext, accountName string) (*Membership, error)
}
