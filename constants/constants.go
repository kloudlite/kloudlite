package constants

type ResourceType string

const (
	ResourceConfig          ResourceType = "config"
	ResourceSecret          ResourceType = "secret"
	ResourceApp             ResourceType = "app"
	ResourceLambda          ResourceType = "lambda"
	ResourceProject         ResourceType = "project"
	ResourceAccount         ResourceType = "account"
	ResourceRouter          ResourceType = "router"
	ResourceManagedService  ResourceType = "mService"
	ResourceManagedResource ResourceType = "mResource"
	ResourceGitPipeline     ResourceType = "gitPipeline"

	// other resources

	ResourceEdgeRegion    ResourceType = "edge-region"
	ResourceCloudProvider ResourceType = "cloud-provider"
	ResourceDevice        ResourceType = "device"
	ResourceEnvironment   ResourceType = "environment"
)

const (
	CacheSessionPrefix = "sessions"
	CookieName         = "hotspot-session"
)

const (
	// source: kubectl apply with an incorrect resource name
	K8sNameValidatorRegex = `^[a-z0-9]([-a-z0-9]*[a-z0-9])?([.][a-z0-9]([-a-z0-9]*[a-z0-9])?)*$`
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

	ReadProject   Action = "read-project"
	CreateProject Action = "create-project"
	UpdateProject Action = "update-project"
	DeleteProject Action = "delete-project"

	InviteProjectMember Action = "invite-project-member"
	RemoveProjectMember Action = "remove-project-member"

	CreateApp         Action = "create-app"
	DeleteApp         Action = "delete-app"
	UpdateApp         Action = "update-app"
	FreezeApp         Action = "freeze-app"
	UnfreezeApp       Action = "unfreeze-app"
	RestartApp        Action = "restart-app"
	InterceptApp      Action = "intercept-app"
	CloseInterceptApp Action = "close-intercept-app"

	CreateConfig Action = "create-config"
	DeleteConfig Action = "delete-config"
	UpdateConfig Action = "update-config"

	CreateSecret Action = "create-secret"
	DeleteSecret Action = "delete-secret"
	UpdateSecret Action = "update-secret"

	CreateMsvc Action = "create-msvc"
	DeleteMsvc Action = "delete-msvc"
	UpdateMsvc Action = "update-msvc"

	CreateMres Action = "create-mres"
	DeleteMres Action = "delete-mres"
	UpdateMres Action = "update-mres"

	CreateRouter Action = "create-router"
	DeleteRouter Action = "delete-router"
	UpdateRouter Action = "update-router"

	CreateDevice Action = "create-device"
	DeleteDevice Action = "delete-device"
	UpdateDevice Action = "update-device"

	CreateEdgeRegion Action = "create-edge-region"
	UpdateEdgeRegion Action = "update-edge-region"
	DeleteEdgeRegion Action = "delete-edge-region"

	CreateCloudProvider Action = "create-cloud-provider"
	UpdateCloudProvider Action = "update-cloud-provider"
	DeleteCloudProvider Action = "delete-cloud-provider"

	CreateEnvironment Action = "create-environment"
	DeleteEnvironment Action = "delete-environment"
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
	ReadProject: {
		AccountOwner, AccountAdmin, AccountMember, ProjectAdmin,
		ProjectMember, ProjectGuest,
	},

	UpdateProject: {
		AccountOwner, AccountAdmin, ProjectAdmin,
		ProjectMember,
	},

	DeleteProject:       {AccountOwner, AccountAdmin, ProjectAdmin},
	InviteProjectMember: {AccountOwner, AccountAdmin, ProjectAdmin},
}

const (
	NamespaceCore string = "kl-core"
)

const (
	ClusterNameKey string = "kloudlite.io/cluster.name"
	EdgeNameKey    string = "kloudlite.io/edge.name"
	AccountNameKey string = "kloudlite.io/account.name"
)
