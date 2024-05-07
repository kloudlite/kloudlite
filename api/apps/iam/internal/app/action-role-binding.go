package app

import (
	t "github.com/kloudlite/api/apps/iam/types"
)

type RoleBindingMap map[t.Action][]t.Role

var roleBindings RoleBindingMap = RoleBindingMap{
	// for accounts
	t.CreateAccount: []t.Role{t.RoleAccountOwner},
	t.GetAccount:    []t.Role{t.RoleAccountOwner, t.RoleAccountAdmin, t.RoleAccountMember, t.RoleProjectAdmin, t.RoleProjectMember},
	t.UpdateAccount: []t.Role{t.RoleAccountOwner, t.RoleAccountAdmin},
	t.DeleteAccount: []t.Role{t.RoleAccountOwner},

	// for account invitations
	t.ListAccountInvitations:  []t.Role{t.RoleAccountOwner, t.RoleAccountAdmin, t.RoleAccountMember},
	t.GetAccountInvitation:    []t.Role{t.RoleAccountOwner, t.RoleAccountAdmin},
	t.DeleteAccountInvitation: []t.Role{t.RoleAccountOwner, t.RoleAccountAdmin},

	t.ReadLogs:    []t.Role{t.RoleAccountMember},
	t.ReadMetrics: []t.Role{t.RoleAccountMember},

	// for account advance actions
	t.DeactivateAccount: []t.Role{t.RoleAccountOwner},
	t.ActivateAccount:   []t.Role{t.RoleAccountOwner},

	// for account membership
	t.InviteAccountAdmin:        []t.Role{t.RoleAccountOwner},
	t.InviteAccountMember:       []t.Role{t.RoleAccountOwner, t.RoleAccountAdmin},
	t.UpdateAccountMembership:   []t.Role{t.RoleAccountOwner, t.RoleAccountAdmin}, // should not update role of himself
	t.RemoveAccountMembership:   []t.Role{t.RoleAccountOwner, t.RoleAccountAdmin},
	t.ListMembershipsForAccount: []t.Role{t.RoleAccountOwner, t.RoleAccountAdmin, t.RoleAccountMember},

	// for clusters
	t.CreateCluster: []t.Role{t.RoleAccountOwner, t.RoleAccountAdmin, t.RoleAccountMember},
	t.DeleteCluster: []t.Role{t.RoleAccountOwner, t.RoleAccountAdmin, t.RoleAccountMember},
	t.UpdateCluster: []t.Role{t.RoleAccountOwner, t.RoleAccountAdmin, t.RoleAccountMember},
	t.ListClusters:  []t.Role{t.RoleAccountOwner, t.RoleAccountAdmin, t.RoleAccountMember, t.RoleProjectAdmin, t.RoleProjectMember},
	t.GetCluster:    []t.Role{t.RoleAccountOwner, t.RoleAccountAdmin, t.RoleAccountMember, t.RoleProjectAdmin, t.RoleProjectMember},

	// for helm release
	t.CreateHelmRelease: []t.Role{t.RoleAccountOwner, t.RoleAccountAdmin, t.RoleAccountMember},
	t.DeleteHelmRelease: []t.Role{t.RoleAccountOwner, t.RoleAccountAdmin, t.RoleAccountMember},
	t.UpdateHelmRelease: []t.Role{t.RoleAccountOwner, t.RoleAccountAdmin, t.RoleAccountMember},
	t.ListHelmReleases:  []t.Role{t.RoleAccountOwner, t.RoleAccountAdmin, t.RoleAccountMember},
	t.GetHelmRelease:    []t.Role{t.RoleAccountOwner, t.RoleAccountAdmin, t.RoleAccountMember},

	// for clusterManagedService
	t.CreateClusterManagedService: []t.Role{t.RoleAccountOwner, t.RoleAccountAdmin, t.RoleAccountMember},
	t.DeleteClusterManagedService: []t.Role{t.RoleAccountOwner, t.RoleAccountAdmin, t.RoleAccountMember},
	t.UpdateClusterManagedService: []t.Role{t.RoleAccountOwner, t.RoleAccountAdmin, t.RoleAccountMember},
	t.ListClusterManagedServices:  []t.Role{t.RoleAccountOwner, t.RoleAccountAdmin, t.RoleAccountMember},
	t.GetClusterManagedService:    []t.Role{t.RoleAccountOwner, t.RoleAccountAdmin, t.RoleAccountMember},

	// for domain entries
	t.CreateDomainEntry: []t.Role{t.RoleAccountOwner, t.RoleAccountAdmin, t.RoleAccountMember},
	t.UpdateDomainEntry: []t.Role{t.RoleAccountOwner, t.RoleAccountAdmin, t.RoleAccountMember},
	t.DeleteDomainEntry: []t.Role{t.RoleAccountOwner, t.RoleAccountAdmin, t.RoleAccountMember},
	t.ListDomainEntries: []t.Role{t.RoleAccountOwner, t.RoleAccountAdmin, t.RoleAccountMember, t.RoleProjectAdmin, t.RoleProjectMember},
	t.GetDomainEntry:    []t.Role{t.RoleAccountOwner, t.RoleAccountAdmin, t.RoleAccountMember, t.RoleProjectAdmin, t.RoleProjectMember},

	// for nodepools
	t.CreateNodepool: []t.Role{t.RoleAccountOwner, t.RoleAccountAdmin, t.RoleAccountMember},
	t.UpdateNodepool: []t.Role{t.RoleAccountOwner, t.RoleAccountAdmin, t.RoleAccountMember},
	t.DeleteNodepool: []t.Role{t.RoleAccountOwner, t.RoleAccountAdmin, t.RoleAccountMember},
	t.ListNodepools:  []t.Role{t.RoleAccountOwner, t.RoleAccountAdmin, t.RoleAccountMember, t.RoleProjectAdmin, t.RoleProjectMember},
	t.GetNodepool:    []t.Role{t.RoleAccountOwner, t.RoleAccountAdmin, t.RoleAccountMember, t.RoleProjectAdmin, t.RoleProjectMember},

	// for cloud provider secrets
	t.CreateCloudProviderSecret: []t.Role{t.RoleAccountOwner, t.RoleAccountAdmin, t.RoleAccountMember},
	t.UpdateCloudProviderSecret: []t.Role{t.RoleAccountOwner, t.RoleAccountAdmin, t.RoleAccountMember},
	t.DeleteCloudProviderSecret: []t.Role{t.RoleAccountOwner, t.RoleAccountAdmin, t.RoleAccountMember},
	t.ListCloudProviderSecrets:  []t.Role{t.RoleAccountOwner, t.RoleAccountAdmin, t.RoleAccountMember},
	t.GetCloudProviderSecret:    []t.Role{t.RoleAccountOwner, t.RoleAccountAdmin, t.RoleAccountMember},

	// image pull secrets
	t.CreateImagePullSecret: []t.Role{t.RoleAccountOwner, t.RoleAccountAdmin, t.RoleAccountMember},
	t.UpdateImagePullSecret: []t.Role{t.RoleAccountOwner, t.RoleAccountAdmin, t.RoleAccountMember},
	t.DeleteImagePullSecret: []t.Role{t.RoleAccountOwner, t.RoleAccountAdmin, t.RoleAccountMember},
	t.ListImagePullSecrets:  []t.Role{t.RoleAccountOwner, t.RoleAccountAdmin, t.RoleAccountMember},
	t.GetImagePullSecret:    []t.Role{t.RoleAccountOwner, t.RoleAccountAdmin, t.RoleAccountMember},

	// for projects
	t.CreateProject: []t.Role{t.RoleAccountOwner, t.RoleAccountAdmin, t.RoleAccountMember},
	t.ListProjects:  []t.Role{t.RoleAccountOwner, t.RoleAccountAdmin, t.RoleAccountMember, t.RoleProjectAdmin, t.RoleProjectMember},
	t.GetProject:    []t.Role{t.RoleAccountOwner, t.RoleAccountAdmin, t.RoleAccountMember, t.RoleProjectAdmin, t.RoleProjectMember},
	t.UpdateProject: []t.Role{t.RoleAccountOwner, t.RoleAccountAdmin, t.RoleAccountMember, t.RoleProjectAdmin, t.RoleProjectMember},
	t.DeleteProject: []t.Role{t.RoleAccountOwner, t.RoleAccountAdmin, t.RoleAccountMember, t.RoleProjectAdmin},

	// for project invitations
	t.InviteProjectAdmin:      []t.Role{t.RoleAccountOwner, t.RoleAccountAdmin, t.RoleAccountMember},
	t.InviteProjectMember:     []t.Role{t.RoleAccountOwner, t.RoleAccountAdmin, t.RoleAccountMember, t.RoleProjectAdmin},
	t.ListProjectInvitations:  []t.Role{t.RoleAccountOwner, t.RoleAccountAdmin, t.RoleAccountMember, t.RoleProjectAdmin},
	t.GetProjectInvitation:    []t.Role{t.RoleAccountOwner, t.RoleAccountAdmin, t.RoleAccountMember, t.RoleProjectAdmin},
	t.DeleteProjectInvitation: []t.Role{t.RoleAccountOwner, t.RoleAccountAdmin, t.RoleAccountMember, t.RoleProjectAdmin},

	t.MutateResourcesInProject: []t.Role{t.RoleAccountOwner, t.RoleAccountAdmin, t.RoleAccountMember, t.RoleProjectAdmin, t.RoleProjectMember},

	t.ListMembershipsForProject: []t.Role{t.RoleAccountOwner, t.RoleAccountAdmin, t.RoleAccountMember, t.RoleProjectAdmin},
	t.UpdateProjectMembership:   []t.Role{t.RoleAccountOwner, t.RoleAccountAdmin, t.RoleAccountMember, t.RoleProjectAdmin},
	t.RemoveProjectMembership:   []t.Role{t.RoleAccountOwner, t.RoleAccountAdmin, t.RoleAccountMember, t.RoleProjectAdmin},

	// for environments
	t.ListEnvironments:  []t.Role{t.RoleAccountOwner, t.RoleAccountAdmin, t.RoleAccountMember, t.RoleProjectAdmin, t.RoleProjectMember},
	t.GetEnvironment:    []t.Role{t.RoleAccountOwner, t.RoleAccountAdmin, t.RoleAccountMember, t.RoleProjectAdmin, t.RoleProjectMember},
	t.CreateEnvironment: []t.Role{t.RoleAccountOwner, t.RoleAccountAdmin, t.RoleAccountMember, t.RoleProjectAdmin, t.RoleProjectMember},
	t.CloneEnvironment:  []t.Role{t.RoleAccountOwner, t.RoleAccountAdmin, t.RoleAccountMember, t.RoleProjectAdmin, t.RoleProjectMember},
	t.UpdateEnvironment: []t.Role{t.RoleAccountOwner, t.RoleAccountAdmin, t.RoleAccountMember, t.RoleProjectAdmin, t.RoleResourceOwner},
	t.DeleteEnvironment: []t.Role{t.RoleAccountOwner, t.RoleAccountAdmin, t.RoleAccountMember, t.RoleProjectAdmin, t.RoleResourceOwner},

	t.ReadResourcesInEnvironment:   []t.Role{t.RoleAccountOwner, t.RoleAccountAdmin, t.RoleAccountMember, t.RoleProjectAdmin, t.RoleProjectMember},
	t.MutateResourcesInEnvironment: []t.Role{t.RoleAccountOwner, t.RoleAccountAdmin, t.RoleAccountMember, t.RoleProjectAdmin, t.RoleResourceOwner},

	// for vpn devices
	t.ListVPNDevices:            []t.Role{t.RoleAccountOwner, t.RoleAccountAdmin, t.RoleAccountMember, t.RoleProjectAdmin, t.RoleProjectMember},
	t.GetVPNDevice:              []t.Role{t.RoleAccountOwner, t.RoleAccountAdmin, t.RoleAccountMember, t.RoleProjectAdmin, t.RoleProjectMember},
	t.CreateVPNDevice:           []t.Role{t.RoleAccountOwner, t.RoleAccountAdmin, t.RoleAccountMember, t.RoleProjectAdmin, t.RoleProjectMember},
	t.UpdateVPNDevice:           []t.Role{t.RoleResourceOwner},
	t.DeleteVPNDevice:           []t.Role{t.RoleAccountOwner, t.RoleAccountAdmin, t.RoleProjectAdmin, t.RoleResourceOwner},
	t.GetVPNDeviceConnectConfig: []t.Role{t.RoleResourceOwner},

	t.ListBuildIntegrations:  []t.Role{t.RoleAccountOwner, t.RoleAccountAdmin, t.RoleAccountMember, t.RoleProjectAdmin, t.RoleProjectMember},
	t.GetBuildIntegration:    []t.Role{t.RoleAccountOwner, t.RoleAccountAdmin, t.RoleAccountMember, t.RoleProjectAdmin, t.RoleProjectMember},
	t.CreateBuildIntegration: []t.Role{t.RoleAccountOwner, t.RoleAccountAdmin, t.RoleAccountMember, t.RoleProjectAdmin, t.RoleProjectMember},
	t.UpdateBuildIntegration: []t.Role{t.RoleAccountOwner, t.RoleAccountAdmin, t.RoleAccountMember, t.RoleProjectAdmin, t.RoleProjectMember},
	t.DeleteBuildIntegration: []t.Role{t.RoleAccountOwner, t.RoleAccountAdmin, t.RoleAccountMember, t.RoleProjectAdmin, t.RoleProjectMember},

	t.ListBuildRuns:  []t.Role{t.RoleAccountOwner, t.RoleAccountAdmin, t.RoleAccountMember, t.RoleProjectAdmin, t.RoleProjectMember},
	t.GetBuildRun:    []t.Role{t.RoleAccountOwner, t.RoleAccountAdmin, t.RoleAccountMember, t.RoleProjectAdmin, t.RoleProjectMember},
	t.CreateBuildRun: []t.Role{t.RoleAccountOwner, t.RoleAccountAdmin, t.RoleAccountMember, t.RoleProjectAdmin, t.RoleProjectMember},
	t.UpdateBuildRun: []t.Role{t.RoleAccountOwner, t.RoleAccountAdmin, t.RoleAccountMember, t.RoleProjectAdmin, t.RoleProjectMember},
	t.DeleteBuildRun: []t.Role{t.RoleAccountOwner, t.RoleAccountAdmin, t.RoleAccountMember, t.RoleProjectAdmin, t.RoleProjectMember},
}
