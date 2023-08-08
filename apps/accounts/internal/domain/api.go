package domain

import (
	"kloudlite.io/apps/accounts/internal/entities"
	iamT "kloudlite.io/apps/iam/types"
	"kloudlite.io/pkg/repos"
)

type Domain interface {
	// CRUD
	CreateAccount(ctx AccountsContext, name string, displayName string) (*entities.Account, error)
	ListAccounts(ctx AccountsContext) ([]*entities.Account, error)
	GetAccount(ctx AccountsContext, name string) (*entities.Account, error)
	UpdateAccount(ctx AccountsContext, name string, email *string) (*entities.Account, error)
	DeleteAccount(ctx AccountsContext, name string) (bool, error)
	ReSyncToK8s(ctx AccountsContext, name string) error

	DeactivateAccount(ctx AccountsContext, name string) (bool, error)
	ActivateAccount(ctx AccountsContext, name string) (bool, error)

	// invitations
	InviteUser(ctx AccountsContext, accountName string, email string, role iamT.Role) (bool, error)
	ListInvitations(ctx AccountsContext, accountName string) ([]*entities.Membership, error)
	DeleteInvitation(ctx AccountsContext, email string) (bool, error)

	// Memberships
	RemoveAccountMember(ctx AccountsContext, accountName string, userId repos.ID) (bool, error)
	UpdateAccountMember(ctx AccountsContext, accountName string, userId repos.ID, role string) (bool, error)

	GetUserMemberships(ctx AccountsContext, resourceRef string) ([]*entities.Membership, error)
	GetAccountMemberships(ctx AccountsContext) ([]*entities.Membership, error)
	GetAccountMembership(ctx AccountsContext, accountName string) (*entities.Membership, error)
}
