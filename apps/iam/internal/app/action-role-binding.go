package app

import (
	t "kloudlite.io/apps/iam/types"
)

type RoleBindingMap map[t.Action][]t.Role

var roleBindings RoleBindingMap = RoleBindingMap{
	t.CreateAccount: []t.Role{t.RoleAccountOwner},
	t.GetAccount:    []t.Role{t.RoleAccountOwner, t.RoleAccountAdmin, t.RoleAccountMember, t.RoleProjectAdmin, t.RoleProjectMember},

	t.UpdateAccount: []t.Role{t.RoleAccountOwner, t.RoleAccountAdmin},
	t.DeleteAccount: []t.Role{t.RoleAccountOwner},

	t.ListAccountInvitations:  []t.Role{t.RoleAccountOwner, t.RoleAccountAdmin},
	t.GetAccountInvitation:    []t.Role{t.RoleAccountOwner, t.RoleAccountAdmin},
	t.DeleteAccountInvitation: []t.Role{t.RoleAccountOwner, t.RoleAccountAdmin},

	t.DeactivateAccount: []t.Role{t.RoleAccountOwner},
	t.ActivateAccount:   []t.Role{t.RoleAccountOwner},

	t.InviteAccountAdmin:  []t.Role{t.RoleAccountOwner},
	t.InviteAccountMember: []t.Role{t.RoleAccountOwner, t.RoleAccountAdmin},

	t.UpdateAccountMembership: []t.Role{t.RoleAccountOwner, t.RoleAccountAdmin}, // should not update role of himself
	t.RemoveAccountMembership: []t.Role{t.RoleAccountOwner, t.RoleAccountAdmin},

	t.ListMembershipsForAccount: []t.Role{t.RoleAccountOwner, t.RoleAccountAdmin},

	t.CreateCluster: []t.Role{t.RoleAccountOwner, t.RoleAccountAdmin},
	t.DeleteCluster: []t.Role{t.RoleAccountOwner, t.RoleAccountAdmin},
	t.UpdateCluster: []t.Role{t.RoleAccountOwner, t.RoleAccountAdmin},
	t.ListClusters:  []t.Role{t.RoleAccountOwner, t.RoleAccountAdmin, t.RoleAccountMember, t.RoleProjectAdmin, t.RoleProjectMember},
	t.GetCluster:    []t.Role{t.RoleAccountOwner, t.RoleAccountAdmin, t.RoleAccountAdmin, t.RoleProjectAdmin, t.RoleProjectMember},

	t.CreateNodepool: []t.Role{t.RoleAccountOwner, t.RoleAccountAdmin},
	t.UpdateNodepool: []t.Role{t.RoleAccountOwner, t.RoleAccountAdmin},
	t.DeleteNodepool: []t.Role{t.RoleAccountOwner, t.RoleAccountAdmin},
	t.ListNodepools:  []t.Role{t.RoleAccountOwner, t.RoleAccountAdmin, t.RoleAccountMember, t.RoleProjectAdmin, t.RoleProjectMember},
	t.GetNodepool:    []t.Role{t.RoleAccountOwner, t.RoleAccountAdmin, t.RoleAccountMember, t.RoleProjectAdmin, t.RoleProjectMember},

	t.CreateCloudProviderSecret: []t.Role{t.RoleAccountOwner, t.RoleAccountAdmin},
	t.UpdateCloudProviderSecret: []t.Role{t.RoleAccountOwner, t.RoleAccountAdmin},
	t.DeleteCloudProviderSecret: []t.Role{t.RoleAccountOwner, t.RoleAccountAdmin},
	t.ListCloudProviderSecrets:  []t.Role{t.RoleAccountOwner, t.RoleAccountAdmin},
	t.GetCloudProviderSecret:    []t.Role{t.RoleAccountOwner, t.RoleAccountAdmin},

	t.CreateProject: []t.Role{t.RoleAccountOwner, t.RoleAccountAdmin},
	t.ListProjects:  []t.Role{t.RoleAccountOwner, t.RoleAccountAdmin, t.RoleProjectAdmin, t.RoleProjectMember},
	t.GetProject:    []t.Role{t.RoleAccountOwner, t.RoleAccountAdmin, t.RoleProjectAdmin, t.RoleProjectMember},
	t.UpdateProject: []t.Role{t.RoleAccountOwner, t.RoleAccountAdmin, t.RoleProjectAdmin, t.RoleProjectMember},
	t.DeleteProject: []t.Role{t.RoleAccountOwner, t.RoleAccountAdmin, t.RoleProjectAdmin},

	t.InviteProjectAdmin:  []t.Role{t.RoleAccountOwner, t.RoleAccountAdmin},
	t.InviteProjectMember: []t.Role{t.RoleAccountOwner, t.RoleAccountAdmin, t.RoleProjectAdmin},

	t.ListProjectInvitations:  []t.Role{t.RoleAccountOwner, t.RoleAccountAdmin, t.RoleProjectAdmin},
	t.GetProjectInvitation:    []t.Role{t.RoleAccountOwner, t.RoleAccountAdmin, t.RoleProjectAdmin},
	t.DeleteProjectInvitation: []t.Role{t.RoleAccountOwner, t.RoleAccountAdmin, t.RoleProjectAdmin},

	t.MutateResourcesInProject: []t.Role{t.RoleAccountOwner, t.RoleAccountAdmin, t.RoleProjectAdmin, t.RoleProjectMember},

	t.ListMembershipsForProject: []t.Role{t.RoleAccountOwner, t.RoleAccountAdmin, t.RoleProjectAdmin},
	t.UpdateProjectMembership:   []t.Role{t.RoleAccountOwner, t.RoleAccountAdmin, t.RoleProjectAdmin},
	t.RemoveProjectMembership:   []t.Role{t.RoleAccountOwner, t.RoleAccountAdmin, t.RoleProjectAdmin},

	t.ListEnvironments:  []t.Role{t.RoleAccountOwner, t.RoleAccountAdmin, t.RoleProjectAdmin, t.RoleProjectMember},
	t.GetEnvironment:    []t.Role{t.RoleAccountOwner, t.RoleAccountAdmin, t.RoleProjectAdmin, t.RoleProjectMember},
	t.CreateEnvironment: []t.Role{t.RoleAccountOwner, t.RoleAccountAdmin, t.RoleProjectAdmin, t.RoleProjectMember},
	t.UpdateEnvironment: []t.Role{t.RoleAccountOwner, t.RoleAccountAdmin, t.RoleProjectAdmin, t.RoleResourceOwner},
	t.DeleteEnvironment: []t.Role{t.RoleAccountOwner, t.RoleAccountAdmin, t.RoleProjectAdmin, t.RoleResourceOwner},

	t.ReadResourcesInEnvironment:   []t.Role{t.RoleAccountOwner, t.RoleAccountAdmin, t.RoleProjectAdmin, t.RoleProjectMember},
	t.MutateResourcesInEnvironment: []t.Role{t.RoleAccountOwner, t.RoleAccountAdmin, t.RoleProjectAdmin, t.RoleResourceOwner},

	t.ListWorkspaces:  []t.Role{t.RoleAccountOwner, t.RoleAccountAdmin, t.RoleProjectAdmin, t.RoleProjectMember, t.RoleResourceOwner},
	t.GetWorkspace:    []t.Role{t.RoleAccountOwner, t.RoleAccountAdmin, t.RoleProjectAdmin, t.RoleProjectMember, t.RoleResourceOwner},
	t.CreateWorkspace: []t.Role{t.RoleAccountOwner, t.RoleAccountAdmin, t.RoleProjectAdmin, t.RoleProjectMember},
	t.UpdateWorkspace: []t.Role{t.RoleAccountOwner, t.RoleAccountAdmin, t.RoleProjectAdmin, t.RoleResourceOwner},
	t.DeleteWorkspace: []t.Role{t.RoleAccountOwner, t.RoleAccountAdmin, t.RoleProjectAdmin, t.RoleResourceOwner},

	t.ReadResourcesInWorkspace:   []t.Role{t.RoleAccountOwner, t.RoleAccountAdmin, t.RoleProjectAdmin, t.RoleProjectMember},
	t.MutateResourcesInWorkspace: []t.Role{t.RoleResourceOwner},

	t.ListVPNDevices:  []t.Role{t.RoleAccountOwner, t.RoleAccountAdmin, t.RoleProjectAdmin, t.RoleProjectMember},
	t.GetVPNDevice:    []t.Role{t.RoleAccountOwner, t.RoleAccountAdmin, t.RoleProjectAdmin, t.RoleProjectMember},
	t.CreateVPNDevice: []t.Role{t.RoleAccountOwner, t.RoleAccountAdmin, t.RoleProjectAdmin, t.RoleProjectMember},
	t.UpdateVPNDevice: []t.Role{t.RoleResourceOwner},
	t.DeleteVPNDevice: []t.Role{t.RoleAccountOwner, t.RoleAccountAdmin, t.RoleProjectAdmin, t.RoleResourceOwner},
}
