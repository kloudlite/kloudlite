package common

type ResourceType string

const (
	ResourceConfig          ResourceType = "config"
	ResourceSecret          ResourceType = "secret"
	ResourceApp             ResourceType = "app"
	ResourceProject         ResourceType = "project"
	ResourceAccount         ResourceType = "account"
	ResourceRouter          ResourceType = "router"
	ResourceManagedService  ResourceType = "mService"
	ResourceManagedResource ResourceType = "mResource"
	ResourceGitPipeline     ResourceType = "gitPipeline"
)

const (
	CacheSessionPrefix = "redis-auth"
	CookieName         = "hotspot-session"
)

const (
	ProviderGithub = "github"
	ProviderGitlab = "gitlab"
	ProviderGoogle = "google"
)

type Action string

const (
	Undefined           Action = ""
	CreateAccount       Action = "create-account"
	UpdateAccount       Action = "update-account"
	DeleteAccount       Action = "delete-account"
	InviteAccountMember Action = "invite-account-member"
	PayBill             Action = "pay-bill"

	ReadProject         Action = "read-project"
	CreateProject       Action = "create-project"
	UpdateProject       Action = "update-project"
	DeleteProject       Action = "delete-project"
	InviteProjectMember Action = "invite-project-member"
)

type Role string

const (
	AccountOwner  Role = "account-owner"
	AccountAdmin  Role = "account-admin"
	AccountMember Role = "account-member"
	AccountGuest  Role = "account-guest"
	AccountBiller Role = "account-biller"

	ProjectAdmin  Role = "project-admin"
	ProjectMember Role = "project-member"
	ProjectGuest  Role = "project-guest"
)

var ActionMap = map[Action][]Role{
	CreateAccount:       {AccountOwner},
	UpdateAccount:       {AccountOwner, AccountAdmin},
	DeleteAccount:       {AccountOwner, AccountAdmin},
	InviteAccountMember: {AccountOwner, AccountAdmin},
	PayBill:             {AccountOwner, AccountAdmin, AccountBiller},

	CreateProject: {AccountOwner, AccountAdmin},
	ReadProject: {AccountOwner, AccountAdmin, AccountMember, ProjectAdmin,
		ProjectMember, ProjectGuest},

	UpdateProject: {AccountOwner, AccountAdmin, ProjectAdmin,
		ProjectMember},

	DeleteProject:       {AccountOwner, AccountAdmin, ProjectAdmin},
	InviteProjectMember: {AccountOwner, AccountAdmin, ProjectAdmin},
}
