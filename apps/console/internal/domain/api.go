package domain

import (
	"context"

	"kloudlite.io/apps/console/internal/domain/entities"
)

type ConsoleContext struct {
	context.Context
	clusterName string
	accountName string
}

func NewConsoleContext(parent context.Context, accountName string, clusterName string) ConsoleContext {
	return ConsoleContext{
		Context:     parent,
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

	// apps:query
	ListApps(ctx ConsoleContext, namespace string) ([]*entities.App, error)
	GetApp(ctx ConsoleContext, namespace, name string) (*entities.App, error)

	// apps:mutation
	CreateApp(ctx ConsoleContext, app entities.App) (*entities.App, error)
	UpdateApp(ctx ConsoleContext, app entities.App) (*entities.App, error)
	DeleteApp(ctx ConsoleContext, namespace, name string) error

	//configs:query
	ListConfigs(ctx ConsoleContext, namespace string) ([]*entities.Config, error)
	GetConfig(ctx ConsoleContext, namespace, name string) (*entities.Config, error)

	//configs:mutation
	CreateConfig(ctx ConsoleContext, config entities.Config) (*entities.Config, error)
	UpdateConfig(ctx ConsoleContext, config entities.Config) (*entities.Config, error)
	DeleteConfig(ctx ConsoleContext, namespace, name string) error

	//secrets:query
	ListSecrets(ctx ConsoleContext, namespace string) ([]*entities.Secret, error)
	GetSecret(ctx ConsoleContext, namespace, name string) (*entities.Secret, error)

	//secrets:mutation
	CreateSecret(ctx ConsoleContext, secret entities.Secret) (*entities.Secret, error)
	UpdateSecret(ctx ConsoleContext, secret entities.Secret) (*entities.Secret, error)
	DeleteSecret(ctx ConsoleContext, namespace, name string) error

	//routers:query
	ListRouters(ctx ConsoleContext, namespace string) ([]*entities.Router, error)
	GetRouter(ctx ConsoleContext, namespace, name string) (*entities.Router, error)

	//routers:mutation
	CreateRouter(ctx ConsoleContext, router entities.Router) (*entities.Router, error)
	UpdateRouter(ctx ConsoleContext, router entities.Router) (*entities.Router, error)
	DeleteRouter(ctx ConsoleContext, namespace, name string) error

	//msvc:query
	ListManagedServices(ctx ConsoleContext, namespace string) ([]*entities.MSvc, error)
	GetManagedService(ctx ConsoleContext, namespace, name string) (*entities.MSvc, error)

	//msvc:mutation
	CreateManagedService(ctx ConsoleContext, msvc entities.MSvc) (*entities.MSvc, error)
	UpdateManagedService(ctx ConsoleContext, msvc entities.MSvc) (*entities.MSvc, error)
	DeleteManagedService(ctx ConsoleContext, namespace, name string) error

	//mres:query
	ListManagedResources(ctx ConsoleContext, namespace string) ([]*entities.MRes, error)
	GetManagedResource(ctx ConsoleContext, namespace, name string) (*entities.MRes, error)

	//mres:mutation
	CreateManagedResource(ctx ConsoleContext, mres entities.MRes) (*entities.MRes, error)
	UpdateManagedResource(ctx ConsoleContext, mres entities.MRes) (*entities.MRes, error)
	DeleteManagedResource(ctx ConsoleContext, namespace, name string) error
}
