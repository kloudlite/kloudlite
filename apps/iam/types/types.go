package types

import "fmt"

type Action string

const (
	CreateAccount Action = "create-account"
	ListAccounts  Action = "list-accounts"
	GetAccount    Action = "get-account"
	UpdateAccount Action = "update-account"
	DeleteAccount Action = "delete-account"

	InviteAccountMember Action = "invite-account-member"
	InviteAccountAdmin  Action = "invite-account-admin"

	CreateProject Action = "create-project"
	ListProjects  Action = "list-projects"
	GetProject    Action = "get-project"
	UpdateProject Action = "update-project"
	DeleteProject Action = "delete-project"

	InviteProjectAdmin  Action = "invite-project-admin"
	InviteProjectMember Action = "invite-project-member"

	MutateResourcesInProject Action = "mutate-resources-in-project"
)

func NewResourceRef(clusterName, kind, namespace, name string) string {
	return fmt.Sprintf("%s/%s/%s/%s", clusterName, kind, namespace, name)
}
