package types

import "fmt"

type ResourceType string

const (
	ResourceAccount ResourceType = "account"
	ResourceProject ResourceType = "project"
)

type Role string

const (
	RoleAccountOwner  Role = "account-owner"
	RoleAccountAdmin  Role = "account-admin"
	RoleAccountMember Role = "account-member"

	RoleProjectAdmin  Role = "project-admin"
	RoleProjectMember Role = "project-member"
)

type Action string

const (
	CreateAccount Action = "create-account"
	ListAccounts  Action = "list-accounts"
	GetAccount    Action = "get-account"
	UpdateAccount Action = "update-account"
	DeleteAccount Action = "delete-account"

	InviteAccountMember Action = "invite-account-member"
	InviteAccountAdmin  Action = "invite-account-admin"

	UpdateAccountMember Action = "update-account-member"

	ActivateAccount   Action = "activate-account"
	DeactivateAccount Action = "deactivate-account"

	CreateProject Action = "create-project"
	ListProjects  Action = "list-projects"
	GetProject    Action = "get-project"
	UpdateProject Action = "update-project"
	DeleteProject Action = "delete-project"

	InviteProjectAdmin  Action = "invite-project-admin"
	InviteProjectMember Action = "invite-project-member"

	MutateResourcesInProject Action = "mutate-resources-in-project"
)

func NewResourceRef(accountName string, resourceType ResourceType, resourceName string) string {
	return fmt.Sprintf("%s/%s/%s", accountName, resourceType, resourceName)
}
