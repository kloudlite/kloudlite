package domain

import (
	"context"

	"kloudlite.io/apps/console/internal/domain/entities"
	"kloudlite.io/pkg/repos"
)

type ConsoleContext struct {
	context.Context
	clusterName string
	accountName string
	userId      repos.ID
}

func NewConsoleContext(parent context.Context, userId repos.ID, accountName, clusterName string) ConsoleContext {
	return ConsoleContext{
		Context:     parent,
		userId:      userId,
		clusterName: clusterName,
		accountName: accountName,
	}
}

type Domain interface {
	// project:query
	ListProjects(ctx ConsoleContext) ([]*entities.Project, error)
	GetProject(ctx ConsoleContext, name string) (*entities.Project, error)

	// project:mutation
	CreateProject(ctx ConsoleContext, project entities.Project) (*entities.Project, error)
	UpdateProject(ctx ConsoleContext, project entities.Project) (*entities.Project, error)
	DeleteProject(ctx ConsoleContext, name string) error

	// project:messaging-updates
	OnApplyProjectError(ctx ConsoleContext, err error, name string) error
	OnDeleteProjectMessage(ctx ConsoleContext, cluster entities.Project) error
	OnUpdateProjectMessage(ctx ConsoleContext, cluster entities.Project) error

	// apps:query
	ListApps(ctx ConsoleContext, namespace string) ([]*entities.App, error)
	GetApp(ctx ConsoleContext, namespace, name string) (*entities.App, error)

	// apps:mutation
	CreateApp(ctx ConsoleContext, app entities.App) (*entities.App, error)
	UpdateApp(ctx ConsoleContext, app entities.App) (*entities.App, error)
	DeleteApp(ctx ConsoleContext, namespace, name string) error

	// apps:messaging-updates
	OnApplyAppError(ctx ConsoleContext, err error, namespace string, name string) error
	OnDeleteAppMessage(ctx ConsoleContext, app entities.App) error
	OnUpdateAppMessage(ctx ConsoleContext, app entities.App) error

	//configs:query
	ListConfigs(ctx ConsoleContext, namespace string) ([]*entities.Config, error)
	GetConfig(ctx ConsoleContext, namespace, name string) (*entities.Config, error)

	//configs:mutation
	CreateConfig(ctx ConsoleContext, config entities.Config) (*entities.Config, error)
	UpdateConfig(ctx ConsoleContext, config entities.Config) (*entities.Config, error)
	DeleteConfig(ctx ConsoleContext, namespace, name string) error

	// configs:messaging-updates
	OnApplyConfigError(ctx ConsoleContext, err error, namespace, name string) error
	OnDeleteConfigMessage(ctx ConsoleContext, config entities.Config) error
	OnUpdateConfigMessage(ctx ConsoleContext, config entities.Config) error

	//secrets:query
	ListSecrets(ctx ConsoleContext, namespace string) ([]*entities.Secret, error)
	GetSecret(ctx ConsoleContext, namespace, name string) (*entities.Secret, error)

	//secrets:mutation
	CreateSecret(ctx ConsoleContext, secret entities.Secret) (*entities.Secret, error)
	UpdateSecret(ctx ConsoleContext, secret entities.Secret) (*entities.Secret, error)
	DeleteSecret(ctx ConsoleContext, namespace, name string) error

	// apps:messaging-updates
	OnApplySecretError(ctx ConsoleContext, err error, namespace, name string) error
	OnDeleteSecretMessage(ctx ConsoleContext, secret entities.Secret) error
	OnUpdateSecretMessage(ctx ConsoleContext, secret entities.Secret) error

	//routers:query
	ListRouters(ctx ConsoleContext, namespace string) ([]*entities.Router, error)
	GetRouter(ctx ConsoleContext, namespace, name string) (*entities.Router, error)

	//routers:mutation
	CreateRouter(ctx ConsoleContext, router entities.Router) (*entities.Router, error)
	UpdateRouter(ctx ConsoleContext, router entities.Router) (*entities.Router, error)
	DeleteRouter(ctx ConsoleContext, namespace, name string) error

	// routers:messaging-updates
	OnApplyRouterError(ctx ConsoleContext, err error, namespace string, name string) error
	OnDeleteRouterMessage(ctx ConsoleContext, router entities.Router) error
	OnUpdateRouterMessage(ctx ConsoleContext, router entities.Router) error

	//msvc:query
	ListManagedServices(ctx ConsoleContext, namespace string) ([]*entities.MSvc, error)
	GetManagedService(ctx ConsoleContext, namespace, name string) (*entities.MSvc, error)

	//msvc:mutation
	CreateManagedService(ctx ConsoleContext, msvc entities.MSvc) (*entities.MSvc, error)
	UpdateManagedService(ctx ConsoleContext, msvc entities.MSvc) (*entities.MSvc, error)
	DeleteManagedService(ctx ConsoleContext, namespace, name string) error

	// msvc:messaging-updates
	OnApplyManagedServiceError(ctx ConsoleContext, err error, namespacee string, name string) error
	OnDeleteManagedServiceMessage(ctx ConsoleContext, msvc entities.MSvc) error
	OnUpdateManagedServiceMessage(ctx ConsoleContext, msvc entities.MSvc) error

	//mres:query
	ListManagedResources(ctx ConsoleContext, namespace string) ([]*entities.MRes, error)
	GetManagedResource(ctx ConsoleContext, namespace, name string) (*entities.MRes, error)

	//mres:mutation
	CreateManagedResource(ctx ConsoleContext, mres entities.MRes) (*entities.MRes, error)
	UpdateManagedResource(ctx ConsoleContext, mres entities.MRes) (*entities.MRes, error)
	DeleteManagedResource(ctx ConsoleContext, namespace, name string) error

	// mres:messaging-updates
	OnApplyManagedResourceError(ctx ConsoleContext, err error, namespace string, name string) error
	OnDeleteManagedResourceMessage(ctx ConsoleContext, mres entities.MRes) error
	OnUpdateManagedResourceMessage(ctx ConsoleContext, mres entities.MRes) error
}
