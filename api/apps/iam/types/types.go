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

	CreateSecretsInAccount Action = "create-secrets-in-account"
	ReadSecretsFromAccount Action = "read-secrets-from-account"

	InviteAccountMember Action = "invite-account-member"
	InviteAccountAdmin  Action = "invite-account-admin"

	ListAccountInvitations Action = "list-account-invitations"
	GetAccountInvitation   Action = "get-account-invitation"

	ListProjectInvitations Action = "list-project-invitations"
	GetProjectInvitation   Action = "get-project-invitation"

	DeleteAccountInvitation Action = "delete-account-invitation"
	DeleteProjectInvitation Action = "delete-project-invitation"

	ListMembershipsForAccount Action = "list-memberships-for-account"

	RemoveAccountMembership Action = "remove-account-membership"
	UpdateAccountMembership Action = "update-account-membership"

	ActivateAccount   Action = "activate-account"
	DeactivateAccount Action = "deactivate-account"

	// clusters
	CreateCluster Action = "create-cluster"
	DeleteCluster Action = "delete-cluster"
	ListClusters  Action = "list-clusters"
	GetCluster    Action = "get-cluster"
	UpdateCluster Action = "update-cluster"

	// nodepools
	CreateNodepool Action = "create-nodepool"
	DeleteNodepool Action = "delete-nodepool"
	ListNodepools  Action = "list-nodepools"
	GetNodepool    Action = "get-nodepool"
	UpdateNodepool Action = "update-nodepool"

	CreateCloudProviderSecret Action = "create-cloud-provider-secret"
	UpdateCloudProviderSecret Action = "update-cloud-provider-secret"
	DeleteCloudProviderSecret Action = "delete-cloud-provider-secret"

	ListCloudProviderSecrets Action = "list-cloud-provider-secrets"
	GetCloudProviderSecret   Action = "get-cloud-provider-secret"

	CreateProject Action = "create-project"
	ListProjects  Action = "list-projects"
	GetProject    Action = "get-project"
	UpdateProject Action = "update-project"
	DeleteProject Action = "delete-project"

	// environments
	CreateEnvironment Action = "create-environment"
	UpdateEnvironment Action = "update-environment"
	DeleteEnvironment Action = "delete-environment"
	GetEnvironment    Action = "get-environment"
	ListEnvironments  Action = "list-environments"

	// invite
	InviteProjectAdmin  Action = "invite-project-admin"
	InviteProjectMember Action = "invite-project-member"

	MutateResourcesInProject Action = "mutate-resources-in-project"

	ListMembershipsForProject Action = "list-memberships-for-project"
	UpdateProjectMembership   Action = "update-project-membership"
	RemoveProjectMembership   Action = "remove-project-membership"
)

func NewResourceRef(accountName string, resourceType ResourceType, resourceName string) string {
	return fmt.Sprintf("%s/%s/%s", accountName, resourceType, resourceName)
}
