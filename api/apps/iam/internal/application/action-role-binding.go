package application

import (
	t "kloudlite.io/apps/iam/types"
)

type RoleBindingMap map[t.Action][]t.Role

var roleBindings RoleBindingMap = RoleBindingMap{
	t.CreateAccount: []t.Role{t.RoleAccountOwner},
	t.GetAccount:    []t.Role{t.RoleAccountAdmin, t.RoleAccountMember, t.RoleProjectAdmin, t.RoleProjectMember},
	//t.ListAccounts:  []t.Role{t.RoleAccountOwner, t.RoleAccountAdmin, t.RoleProjectAdmin},
	t.UpdateAccount: []t.Role{t.RoleAccountOwner, t.RoleAccountAdmin},
	t.DeleteAccount: []t.Role{t.RoleAccountOwner},

	t.InviteAccountAdmin:  []t.Role{t.RoleAccountOwner},
	t.InviteAccountMember: []t.Role{t.RoleAccountOwner, t.RoleAccountAdmin},

	t.UpdateAccountMember: []t.Role{t.RoleAccountOwner, t.RoleAccountAdmin},

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

	t.CreateEnvironment: []t.Role{t.RoleAccountOwner, t.RoleAccountAdmin, t.RoleProjectAdmin, t.RoleProjectMember},
	t.ListEnvironments:  []t.Role{t.RoleAccountOwner, t.RoleAccountAdmin, t.RoleProjectAdmin, t.RoleProjectMember},
	t.GetEnvironment:    []t.Role{t.RoleAccountOwner, t.RoleAccountAdmin, t.RoleProjectAdmin, t.RoleProjectMember},
	t.UpdateEnvironment: []t.Role{t.RoleAccountOwner, t.RoleAccountAdmin, t.RoleProjectAdmin, t.RoleProjectMember},
	t.DeleteEnvironment: []t.Role{t.RoleAccountOwner, t.RoleAccountAdmin, t.RoleProjectAdmin, t.RoleProjectMember},

	t.InviteProjectAdmin:  []t.Role{t.RoleAccountOwner, t.RoleAccountAdmin},
	t.InviteProjectMember: []t.Role{t.RoleAccountOwner, t.RoleAccountAdmin, t.RoleProjectAdmin},

	t.MutateResourcesInProject: []t.Role{t.RoleAccountOwner, t.RoleAccountAdmin, t.RoleProjectAdmin, t.RoleProjectMember},
}
