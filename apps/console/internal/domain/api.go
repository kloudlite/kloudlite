package domain

import (
	"context"

	"kloudlite.io/apps/console/internal/domain/entities"
	"kloudlite.io/pkg/repos"
)

type ConsoleContext struct {
	context.Context
	ClusterName string
	AccountName string
	UserId      repos.ID
}

func (c ConsoleContext) GetAccountName() string {
	return c.AccountName
}

func NewConsoleContext(parent context.Context, userId repos.ID, accountName, clusterName string) ConsoleContext {
	return ConsoleContext{
		Context:     parent,
		UserId:      userId,
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
	ResTypeApp             ResType = "app"
	ResTypeConfig          ResType = "config"
	ResTypeSecret          ResType = "secret"
	ResTypeRouter          ResType = "router"
	ResTypeManagedService  ResType = "managedservice"
	ResTypeManagedResource ResType = "managedresource"
)

type Domain interface {
	CheckNameAvailability(ctx context.Context, resType ResType, accountName string, name string) (*CheckNameAvailabilityOutput, error)

	ListProjects(ctx context.Context, userId repos.ID, accountName string, clusterName *string) ([]*entities.Project, error)
	GetProject(ctx ConsoleContext, name string) (*entities.Project, error)

	CreateProject(ctx ConsoleContext, project entities.Project) (*entities.Project, error)
	UpdateProject(ctx ConsoleContext, project entities.Project) (*entities.Project, error)
	DeleteProject(ctx ConsoleContext, name string) error

	OnApplyProjectError(ctx ConsoleContext, errMsg string, name string) error
	OnDeleteProjectMessage(ctx ConsoleContext, cluster entities.Project) error
	OnUpdateProjectMessage(ctx ConsoleContext, cluster entities.Project) error

	ResyncProject(ctx ConsoleContext, name string) error

	ListWorkspaces(ctx ConsoleContext, namespace string) ([]*entities.Workspace, error)
	GetWorkspace(ctx ConsoleContext, namespace, name string) (*entities.Workspace, error)

	CreateWorkspace(ctx ConsoleContext, env entities.Workspace) (*entities.Workspace, error)
	UpdateWorkspace(ctx ConsoleContext, env entities.Workspace) (*entities.Workspace, error)
	DeleteWorkspace(ctx ConsoleContext, namespace, name string) error

	OnApplyWorkspaceError(ctx ConsoleContext, errMsg, namespace, name string) error
	OnDeleteEnvironmentMessage(ctx ConsoleContext, cluster entities.Workspace) error
	OnUpdateEnvironmentMessage(ctx ConsoleContext, cluster entities.Workspace) error

	ResyncWorkspace(ctx ConsoleContext, namespace, name string) error

	ListApps(ctx ConsoleContext, namespace string) ([]*entities.App, error)
	GetApp(ctx ConsoleContext, namespace, name string) (*entities.App, error)

	CreateApp(ctx ConsoleContext, app entities.App) (*entities.App, error)
	UpdateApp(ctx ConsoleContext, app entities.App) (*entities.App, error)
	DeleteApp(ctx ConsoleContext, namespace, name string) error

	OnApplyAppError(ctx ConsoleContext, errMsg string, namespace string, name string) error
	OnDeleteAppMessage(ctx ConsoleContext, app entities.App) error
	OnUpdateAppMessage(ctx ConsoleContext, app entities.App) error

	ResyncApp(ctx ConsoleContext, namespace, name string) error

	ListConfigs(ctx ConsoleContext, namespace string) ([]*entities.Config, error)
	GetConfig(ctx ConsoleContext, namespace, name string) (*entities.Config, error)

	CreateConfig(ctx ConsoleContext, config entities.Config) (*entities.Config, error)
	UpdateConfig(ctx ConsoleContext, config entities.Config) (*entities.Config, error)
	DeleteConfig(ctx ConsoleContext, namespace, name string) error

	OnApplyConfigError(ctx ConsoleContext, errMsg, namespace, name string) error
	OnDeleteConfigMessage(ctx ConsoleContext, config entities.Config) error
	OnUpdateConfigMessage(ctx ConsoleContext, config entities.Config) error

	ResyncConfig(ctx ConsoleContext, namespace, name string) error

	ListSecrets(ctx ConsoleContext, namespace string) ([]*entities.Secret, error)
	GetSecret(ctx ConsoleContext, namespace, name string) (*entities.Secret, error)

	CreateSecret(ctx ConsoleContext, secret entities.Secret) (*entities.Secret, error)
	UpdateSecret(ctx ConsoleContext, secret entities.Secret) (*entities.Secret, error)
	DeleteSecret(ctx ConsoleContext, namespace, name string) error

	OnApplySecretError(ctx ConsoleContext, errMsg, namespace, name string) error
	OnDeleteSecretMessage(ctx ConsoleContext, secret entities.Secret) error
	OnUpdateSecretMessage(ctx ConsoleContext, secret entities.Secret) error

	ResyncSecret(ctx ConsoleContext, namespace, name string) error

	ListRouters(ctx ConsoleContext, namespace string) ([]*entities.Router, error)
	GetRouter(ctx ConsoleContext, namespace, name string) (*entities.Router, error)

	CreateRouter(ctx ConsoleContext, router entities.Router) (*entities.Router, error)
	UpdateRouter(ctx ConsoleContext, router entities.Router) (*entities.Router, error)
	DeleteRouter(ctx ConsoleContext, namespace, name string) error

	OnApplyRouterError(ctx ConsoleContext, errMsg string, namespace string, name string) error
	OnDeleteRouterMessage(ctx ConsoleContext, router entities.Router) error
	OnUpdateRouterMessage(ctx ConsoleContext, router entities.Router) error

	ResyncRouter(ctx ConsoleContext, namespace, name string) error

	ListManagedServices(ctx ConsoleContext, namespace string) ([]*entities.MSvc, error)
	GetManagedService(ctx ConsoleContext, namespace, name string) (*entities.MSvc, error)

	CreateManagedService(ctx ConsoleContext, msvc entities.MSvc) (*entities.MSvc, error)
	UpdateManagedService(ctx ConsoleContext, msvc entities.MSvc) (*entities.MSvc, error)
	DeleteManagedService(ctx ConsoleContext, namespace, name string) error

	OnApplyManagedServiceError(ctx ConsoleContext, errMsg string, namespace string, name string) error
	OnDeleteManagedServiceMessage(ctx ConsoleContext, msvc entities.MSvc) error
	OnUpdateManagedServiceMessage(ctx ConsoleContext, msvc entities.MSvc) error

	ResyncManagedService(ctx ConsoleContext, namespace, name string) error

	ListManagedResources(ctx ConsoleContext, namespace string) ([]*entities.MRes, error)
	GetManagedResource(ctx ConsoleContext, namespace, name string) (*entities.MRes, error)

	CreateManagedResource(ctx ConsoleContext, mres entities.MRes) (*entities.MRes, error)
	UpdateManagedResource(ctx ConsoleContext, mres entities.MRes) (*entities.MRes, error)
	DeleteManagedResource(ctx ConsoleContext, namespace, name string) error

	OnApplyManagedResourceError(ctx ConsoleContext, errMsg string, namespace string, name string) error
	OnDeleteManagedResourceMessage(ctx ConsoleContext, mres entities.MRes) error
	OnUpdateManagedResourceMessage(ctx ConsoleContext, mres entities.MRes) error

	ResyncManagedResource(ctx ConsoleContext, namespace, name string) error
}
