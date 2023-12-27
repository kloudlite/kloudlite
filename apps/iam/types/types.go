package types

import "fmt"

type ResourceType string

const (
	ResourceAccount ResourceType = "account"
	ResourceProject ResourceType = "project"

	ResourceEnvironment ResourceType = "environment"
	ResourceWorkspace   ResourceType = "workspace"
	ResourceVPNDevice   ResourceType = "vpn_device"
)

type Role string

const (
	RoleResourceOwner Role = "resource_owner"

	RoleAccountOwner  Role = "account_owner"
	RoleAccountAdmin  Role = "account_admin"
	RoleAccountMember Role = "account_member"

	RoleProjectAdmin  Role = "project_admin"
	RoleProjectMember Role = "project_member"
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

	// cluster managed services
	CreateClusterManagedService Action = "create-cluster-managed-service"
	DeleteClusterManagedService Action = "delete-cluster-managed-service"
	ListClusterManagedServices  Action = "list-cluster-managed-services"
	GetClusterManagedService    Action = "get-cluster-managed-service"
	UpdateClusterManagedService Action = "update-cluster-managed-service"

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

	// invite
	InviteProjectAdmin  Action = "invite-project-admin"
	InviteProjectMember Action = "invite-project-member"

	MutateResourcesInProject Action = "mutate-resources-in-project"

	ListMembershipsForProject Action = "list-memberships-for-project"
	UpdateProjectMembership   Action = "update-project-membership"
	RemoveProjectMembership   Action = "remove-project-membership"

	CreateEnvironment Action = "create-environment"
	UpdateEnvironment Action = "update-environment"
	DeleteEnvironment Action = "delete-environment"
	GetEnvironment    Action = "get-environment"
	ListEnvironments  Action = "list-environments"

	MutateResourcesInEnvironment Action = "mutate-resources-in-environment"
	ReadResourcesInEnvironment   Action = "read-resources-in-environment"

	CreateWorkspace Action = "create-workspace"
	UpdateWorkspace Action = "update-workspace"
	DeleteWorkspace Action = "delete-workspace"
	GetWorkspace    Action = "get-workspace"
	ListWorkspaces  Action = "list-workspaces"

	MutateResourcesInWorkspace Action = "mutate-resources-in-workspace"
	ReadResourcesInWorkspace   Action = "read-resources-in-workspace"

	ListVPNDevices  Action = "list-vpn-devices"
	GetVPNDevice    Action = "get-vpn-device"
	CreateVPNDevice Action = "create-vpn-device"
	UpdateVPNDevice Action = "update-vpn-device"
	DeleteVPNDevice Action = "delete-vpn-device"

	CreateDomainEntry Action = "create-domain-entry"
	UpdateDomainEntry Action = "update-domain-entry"
	DeleteDomainEntry Action = "delete-domain-entry"

	ListDomainEntries Action = "list-domain-entries"
	GetDomainEntry    Action = "get-domain-entry"
)

func NewResourceRef(accountName string, resourceType ResourceType, resourceName string) string {
	return fmt.Sprintf("%s/%s/%s", accountName, resourceType, resourceName)
}
