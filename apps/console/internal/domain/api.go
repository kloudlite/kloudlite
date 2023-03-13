package domain

import (
	"context"

	"kloudlite.io/apps/console/internal/domain/entities"
)

type Domain interface {
	// apps:query
	GetApps(ctx context.Context, namespace string) ([]*entities.App, error)
	GetApp(ctx context.Context, namespace, name string) (*entities.App, error)

	// apps:mutation
	CreateApp(ctx context.Context, app entities.App) (*entities.App, error)
	UpdateApp(ctx context.Context, app entities.App) (*entities.App, error)
	DeleteApp(ctx context.Context, namespace, name string) error

	//configs:query
	GetConfigs(ctx context.Context, namespace string) ([]*entities.Config, error)
	GetConfig(ctx context.Context, namespace, name string) (*entities.Config, error)

	//configs:mutation
	CreateConfig(ctx context.Context, config entities.Config) (*entities.Config, error)
	UpdateConfig(ctx context.Context, config entities.Config) (*entities.Config, error)
	DeleteConfig(ctx context.Context, namespace, name string) error

	//secrets:query
	GetSecrets(ctx context.Context, namespace string) ([]*entities.Secret, error)
	GetSecret(ctx context.Context, namespace, name string) (*entities.Secret, error)

	//secrets:mutation
	CreateSecret(ctx context.Context, secret entities.Secret) (*entities.Secret, error)
	UpdateSecret(ctx context.Context, secret entities.Secret) (*entities.Secret, error)
	DeleteSecret(ctx context.Context, namespace, name string) error

	//routers:query
	GetRouters(ctx context.Context, namespace string) ([]*entities.Router, error)
	GetRouter(ctx context.Context, namespace, name string) (*entities.Router, error)

	//routers:mutation
	CreateRouter(ctx context.Context, router entities.Router) (*entities.Router, error)
	UpdateRouter(ctx context.Context, router entities.Router) (*entities.Router, error)
	DeleteRouter(ctx context.Context, namespace, name string) error

	//msvc:query
	GetManagedServices(ctx context.Context, namespace string) ([]*entities.MSvc, error)
	GetManagedService(ctx context.Context, namespace, name string) (*entities.MSvc, error)

	//msvc:mutation
	CreateManagedService(ctx context.Context, msvc entities.MSvc) (*entities.MSvc, error)
	UpdateManagedService(ctx context.Context, msvc entities.MSvc) (*entities.MSvc, error)
	DeleteManagedService(ctx context.Context, namespace, name string) error


	//mres:query
	GetManagedResources(ctx context.Context, namespace string) ([]*entities.MRes, error)
	GetManagedResource(ctx context.Context, namespace, name string) (*entities.MRes, error)

	//mres:mutation
	CreateManagedResource(ctx context.Context, mres entities.MRes) (*entities.MRes, error)
	UpdateManagedResource(ctx context.Context, mres entities.MRes) (*entities.MRes, error)
	DeleteManagedResource(ctx context.Context, namespace, name string) error
}
