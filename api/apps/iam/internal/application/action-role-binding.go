package application

import (
	"kloudlite.io/apps/iam/internal/domain/entities"
	t "kloudlite.io/apps/iam/types"
)

type RoleBindingMap map[t.Action][]entities.Role

var roleBindings RoleBindingMap = RoleBindingMap{
	t.CreateAccount: []entities.Role{entities.RoleAccountOwner},
	// ListAccounts:  []entities.Role{entities.RoleAccountOwner, entities.RoleAccountAdmin, entities.RoleProjectAdmin},
	t.UpdateAccount: []entities.Role{entities.RoleAccountOwner, entities.RoleAccountAdmin},
	t.DeleteAccount: []entities.Role{entities.RoleAccountOwner},

	t.InviteAccountAdmin:  []entities.Role{entities.RoleAccountOwner},
	t.InviteAccountMember: []entities.Role{entities.RoleAccountOwner, entities.RoleAccountAdmin},

	t.CreateProject: []entities.Role{entities.RoleAccountOwner, entities.RoleAccountAdmin},
	t.ListProjects:  []entities.Role{entities.RoleAccountOwner, entities.RoleAccountAdmin, entities.RoleProjectAdmin, entities.RoleProjectMember},
	t.GetProject:    []entities.Role{entities.RoleAccountOwner, entities.RoleAccountAdmin, entities.RoleProjectAdmin, entities.RoleProjectMember},
	t.UpdateProject: []entities.Role{entities.RoleAccountOwner, entities.RoleAccountAdmin, entities.RoleProjectAdmin, entities.RoleProjectMember},
	t.DeleteProject: []entities.Role{entities.RoleAccountOwner, entities.RoleAccountAdmin, entities.RoleProjectAdmin, entities.RoleProjectMember},

	t.InviteProjectAdmin:  []entities.Role{entities.RoleAccountOwner, entities.RoleAccountAdmin},
	t.InviteProjectMember: []entities.Role{entities.RoleAccountOwner, entities.RoleAccountAdmin, entities.RoleProjectAdmin},

	t.MutateResourcesInProject: []entities.Role{entities.RoleAccountOwner, entities.RoleAccountAdmin, entities.RoleProjectAdmin, entities.RoleProjectMember},
}
