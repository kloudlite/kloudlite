package application

import (
	t "kloudlite.io/apps/iam/types"
)

type RoleBindingMap map[t.Action][]t.Role

var roleBindings RoleBindingMap = RoleBindingMap{
	t.CreateAccount: []t.Role{t.RoleAccountOwner},
	t.GetAccount:    []t.Role{t.RoleAccountAdmin, t.RoleAccountMember, t.RoleProjectAdmin, t.RoleProjectMember},
	// ListAccounts:  []t.Role{t.RoleAccountOwner, t.RoleAccountAdmin, t.RoleProjectAdmin},
	t.UpdateAccount: []t.Role{t.RoleAccountOwner, t.RoleAccountAdmin},
	t.DeleteAccount: []t.Role{t.RoleAccountOwner},

	t.InviteAccountAdmin:  []t.Role{t.RoleAccountOwner},
	t.InviteAccountMember: []t.Role{t.RoleAccountOwner, t.RoleAccountAdmin},

	t.UpdateAccountMember: []t.Role{t.RoleAccountOwner, t.RoleAccountAdmin},

	t.CreateProject: []t.Role{t.RoleAccountOwner, t.RoleAccountAdmin},
	t.ListProjects:  []t.Role{t.RoleAccountOwner, t.RoleAccountAdmin, t.RoleProjectAdmin, t.RoleProjectMember},
	t.GetProject:    []t.Role{t.RoleAccountOwner, t.RoleAccountAdmin, t.RoleProjectAdmin, t.RoleProjectMember},
	t.UpdateProject: []t.Role{t.RoleAccountOwner, t.RoleAccountAdmin, t.RoleProjectAdmin, t.RoleProjectMember},
	t.DeleteProject: []t.Role{t.RoleAccountOwner, t.RoleAccountAdmin, t.RoleProjectAdmin, t.RoleProjectMember},

	t.InviteProjectAdmin:  []t.Role{t.RoleAccountOwner, t.RoleAccountAdmin},
	t.InviteProjectMember: []t.Role{t.RoleAccountOwner, t.RoleAccountAdmin, t.RoleProjectAdmin},

	t.MutateResourcesInProject: []t.Role{t.RoleAccountOwner, t.RoleAccountAdmin, t.RoleProjectAdmin, t.RoleProjectMember},
}
