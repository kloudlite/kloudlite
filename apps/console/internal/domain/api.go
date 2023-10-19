package domain

import (
	"context"
	"kloudlite.io/apps/console/internal/entities"
	"kloudlite.io/pkg/repos"
)

type ConsoleContext struct {
	context.Context
	ClusterName string
	AccountName string

	UserId    repos.ID
	UserEmail string
	UserName  string
}

type UserAndAccountsContext struct {
	context.Context
	AccountName string
	UserId      repos.ID
}

func (c ConsoleContext) GetAccountName() string {
	return c.AccountName
}

func NewConsoleContext(parent context.Context, userId repos.ID, accountName, clusterName string) ConsoleContext {
	return ConsoleContext{
		Context: parent,
		UserId:  userId,

		ClusterName: clusterName,
		AccountName: accountName,
	}
}

type CheckNameAvailabilityOutput struct {
	Result         bool     `json:"result"`
	SuggestedNames []string `json:"suggestedNames,omitempty"`
}

type ResType string

const (
	ResTypeProject         ResType = "project"
	ResTypeEnvironment     ResType = "environment"
	ResTypeWorkspace       ResType = "workspace"
	ResTypeApp             ResType = "app"
	ResTypeConfig          ResType = "config"
	ResTypeSecret          ResType = "secret"
	ResTypeRouter          ResType = "router"
	ResTypeManagedService  ResType = "managed_service"
	ResTypeManagedResource ResType = "managed_resource"
	ResTypeVPNDevice       ResType = "vpn_device"
)

type Domain interface {
	CheckNameAvailability(ctx context.Context, resType ResType, accountName string, namespace *string, name string) (*CheckNameAvailabilityOutput, error)

	ListProjects(ctx context.Context, userId repos.ID, accountName string, clusterName *string, search map[string]repos.MatchFilter, pagination repos.CursorPagination) (*repos.PaginatedRecord[*entities.Project], error)
	GetProject(ctx ConsoleContext, name string) (*entities.Project, error)

	CreateProject(ctx ConsoleContext, project entities.Project) (*entities.Project, error)
	UpdateProject(ctx ConsoleContext, project entities.Project) (*entities.Project, error)
	DeleteProject(ctx ConsoleContext, name string) error

	OnApplyProjectError(ctx ConsoleContext, errMsg string, name string) error
	OnDeleteProjectMessage(ctx ConsoleContext, cluster entities.Project) error
	OnUpdateProjectMessage(ctx ConsoleContext, cluster entities.Project) error

	ResyncProject(ctx ConsoleContext, name string) error

	ListEnvironments(ctx ConsoleContext, namespace string, search map[string]repos.MatchFilter, pq repos.CursorPagination) (*repos.PaginatedRecord[*entities.Workspace], error)
	GetEnvironment(ctx ConsoleContext, namespace, name string) (*entities.Workspace, error)

	CreateEnvironment(ctx ConsoleContext, env entities.Workspace) (*entities.Workspace, error)
	UpdateEnvironment(ctx ConsoleContext, env entities.Workspace) (*entities.Workspace, error)
	DeleteEnvironment(ctx ConsoleContext, namespace, name string) error

	ResyncEnvironment(ctx ConsoleContext, namespace, name string) error

	ListWorkspaces(ctx ConsoleContext, namespace string, search map[string]repos.MatchFilter, pq repos.CursorPagination) (*repos.PaginatedRecord[*entities.Workspace], error)
	GetWorkspace(ctx ConsoleContext, namespace, name string) (*entities.Workspace, error)

	CreateWorkspace(ctx ConsoleContext, env entities.Workspace) (*entities.Workspace, error)
	UpdateWorkspace(ctx ConsoleContext, env entities.Workspace) (*entities.Workspace, error)
	DeleteWorkspace(ctx ConsoleContext, namespace, name string) error

	OnApplyWorkspaceError(ctx ConsoleContext, errMsg, namespace, name string) error
	OnDeleteWorkspaceMessage(ctx ConsoleContext, cluster entities.Workspace) error
	OnUpdateWorkspaceMessage(ctx ConsoleContext, cluster entities.Workspace) error

	ResyncWorkspace(ctx ConsoleContext, namespace, name string) error

	ListApps(ctx ConsoleContext, namespace string, search map[string]repos.MatchFilter, pq repos.CursorPagination) (*repos.PaginatedRecord[*entities.App], error)
	GetApp(ctx ConsoleContext, namespace, name string) (*entities.App, error)

	CreateApp(ctx ConsoleContext, app entities.App) (*entities.App, error)
	UpdateApp(ctx ConsoleContext, app entities.App) (*entities.App, error)
	DeleteApp(ctx ConsoleContext, namespace, name string) error

	OnApplyAppError(ctx ConsoleContext, errMsg string, namespace string, name string) error
	OnDeleteAppMessage(ctx ConsoleContext, app entities.App) error
	OnUpdateAppMessage(ctx ConsoleContext, app entities.App) error

	ResyncApp(ctx ConsoleContext, namespace, name string) error

	ListConfigs(ctx ConsoleContext, namespace string, search map[string]repos.MatchFilter, pq repos.CursorPagination) (*repos.PaginatedRecord[*entities.Config], error)
	GetConfig(ctx ConsoleContext, namespace, name string) (*entities.Config, error)

	CreateConfig(ctx ConsoleContext, config entities.Config) (*entities.Config, error)
	UpdateConfig(ctx ConsoleContext, config entities.Config) (*entities.Config, error)
	DeleteConfig(ctx ConsoleContext, namespace, name string) error

	OnApplyConfigError(ctx ConsoleContext, errMsg, namespace, name string) error
	OnDeleteConfigMessage(ctx ConsoleContext, config entities.Config) error
	OnUpdateConfigMessage(ctx ConsoleContext, config entities.Config) error

	ResyncConfig(ctx ConsoleContext, namespace, name string) error

	ListSecrets(ctx ConsoleContext, namespace string, search map[string]repos.MatchFilter, pq repos.CursorPagination) (*repos.PaginatedRecord[*entities.Secret], error)
	GetSecret(ctx ConsoleContext, namespace, name string) (*entities.Secret, error)

	CreateSecret(ctx ConsoleContext, secret entities.Secret) (*entities.Secret, error)
	UpdateSecret(ctx ConsoleContext, secret entities.Secret) (*entities.Secret, error)
	DeleteSecret(ctx ConsoleContext, namespace, name string) error

	OnApplySecretError(ctx ConsoleContext, errMsg, namespace, name string) error
	OnDeleteSecretMessage(ctx ConsoleContext, secret entities.Secret) error
	OnUpdateSecretMessage(ctx ConsoleContext, secret entities.Secret) error

	ResyncSecret(ctx ConsoleContext, namespace, name string) error

	ListRouters(ctx ConsoleContext, namespace string, search map[string]repos.MatchFilter, pq repos.CursorPagination) (*repos.PaginatedRecord[*entities.Router], error)
	GetRouter(ctx ConsoleContext, namespace, name string) (*entities.Router, error)

	CreateRouter(ctx ConsoleContext, router entities.Router) (*entities.Router, error)
	UpdateRouter(ctx ConsoleContext, router entities.Router) (*entities.Router, error)
	DeleteRouter(ctx ConsoleContext, namespace, name string) error

	OnApplyRouterError(ctx ConsoleContext, errMsg string, namespace string, name string) error
	OnDeleteRouterMessage(ctx ConsoleContext, router entities.Router) error
	OnUpdateRouterMessage(ctx ConsoleContext, router entities.Router) error

	ResyncRouter(ctx ConsoleContext, namespace, name string) error

	ListManagedServices(ctx ConsoleContext, namespace string, search map[string]repos.MatchFilter, pq repos.CursorPagination) (*repos.PaginatedRecord[*entities.ManagedService], error)
	GetManagedService(ctx ConsoleContext, namespace, name string) (*entities.ManagedService, error)

	CreateManagedService(ctx ConsoleContext, msvc entities.ManagedService) (*entities.ManagedService, error)
	UpdateManagedService(ctx ConsoleContext, msvc entities.ManagedService) (*entities.ManagedService, error)
	DeleteManagedService(ctx ConsoleContext, namespace, name string) error

	// Managed Service Templates

	ListManagedSvcTemplates() ([]*entities.MsvcTemplate, error)
	GetManagedSvcTemplate(category string, name string) (*entities.MsvcTemplateEntry, error)

	OnApplyManagedServiceError(ctx ConsoleContext, errMsg string, namespace string, name string) error
	OnDeleteManagedServiceMessage(ctx ConsoleContext, msvc entities.ManagedService) error
	OnUpdateManagedServiceMessage(ctx ConsoleContext, msvc entities.ManagedService) error

	ResyncManagedService(ctx ConsoleContext, namespace, name string) error

	ListManagedResources(ctx ConsoleContext, namespace string, search map[string]repos.MatchFilter, pq repos.CursorPagination) (*repos.PaginatedRecord[*entities.ManagedResource], error)
	GetManagedResource(ctx ConsoleContext, namespace, name string) (*entities.ManagedResource, error)

	CreateManagedResource(ctx ConsoleContext, mres entities.ManagedResource) (*entities.ManagedResource, error)
	UpdateManagedResource(ctx ConsoleContext, mres entities.ManagedResource) (*entities.ManagedResource, error)
	DeleteManagedResource(ctx ConsoleContext, namespace, name string) error

	OnApplyManagedResourceError(ctx ConsoleContext, errMsg string, namespace string, name string) error
	OnDeleteManagedResourceMessage(ctx ConsoleContext, mres entities.ManagedResource) error
	OnUpdateManagedResourceMessage(ctx ConsoleContext, mres entities.ManagedResource) error

	ResyncManagedResource(ctx ConsoleContext, namespace, name string) error

	// image pull secrets

	ListImagePullSecrets(ctx ConsoleContext, namespace string, search map[string]repos.MatchFilter, pagination repos.CursorPagination) (*repos.PaginatedRecord[*entities.ImagePullSecret], error)
	GetImagePullSecret(ctx ConsoleContext, namespace string, name string) (*entities.ImagePullSecret, error)
	CreateImagePullSecret(ctx ConsoleContext, secret entities.ImagePullSecret) (*entities.ImagePullSecret, error)
	DeleteImagePullSecret(ctx ConsoleContext, namespace string, name string) error

	OnApplyImagePullSecretError(ctx ConsoleContext, errMsg string, namespace string, name string) error
	OnDeleteImagePullSecretMessage(ctx ConsoleContext, mres entities.ImagePullSecret) error
	OnUpdateImagePullSecretMessage(ctx ConsoleContext, mres entities.ImagePullSecret) error

	ResyncImagePullSecret(ctx ConsoleContext, namespace, name string) error
}
