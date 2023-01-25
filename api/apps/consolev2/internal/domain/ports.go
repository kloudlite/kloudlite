package domain

import (
	"context"
	fWebsocket "github.com/gofiber/websocket/v2"
	"kloudlite.io/apps/consolev2/internal/domain/entities"
	"kloudlite.io/pkg/cache"
	"kloudlite.io/pkg/repos"
)

type Domain interface {
	SetupAccount(ctx context.Context, accountId repos.ID) (bool, error)

	//cloudproviders

	CreateCloudProvider(ctx context.Context, cloudProvider *entities.CloudProvider, creds entities.SecretData) (*entities.CloudProvider, error)
	UpdateCloudProvider(ctx context.Context, cloudProvider entities.CloudProvider, creds entities.SecretData) (*entities.CloudProvider, error)
	DeleteCloudProvider(ctx context.Context, name string) error
	ListCloudProviders(ctx context.Context, accountId repos.ID) ([]*entities.CloudProvider, error)
	GetCloudProvider(ctx context.Context, name string) (*entities.CloudProvider, error)

	// environments

	GetEnvironments(ctx context.Context, projectName string) ([]*entities.Environment, error)
	GetEnvironment(ctx context.Context, envName string) (*entities.Environment, error)
	CreateEnvironment(ctx context.Context, env entities.Environment) (*entities.Environment, error)

	//projects

	CreateProject(ctx context.Context, project entities.Project) (*entities.Project, error)
	GetAccountProjects(ctx context.Context, accountId repos.ID) ([]*entities.Project, error)
	GetProjectWithID(ctx context.Context, projectId repos.ID) (*entities.Project, error)
	GetProjectWithName(ctx context.Context, projectName string) (*entities.Project, error)

	//apps

	InstallApp(ctx context.Context, app entities.App) (*entities.App, error)
	UpdateApp(ctx context.Context, app entities.App) (*entities.App, error)
	GetApps(ctx context.Context, projectName string) ([]*entities.App, error)
	GetInterceptedApps(ctx context.Context, deviceName string) ([]*entities.App, error)
	FreezeApp(ctx context.Context, appName string) error
	UnFreezeApp(ctx context.Context, appName string) error
	RestartApp(ctx context.Context, appName string) error
	GetApp(ctx context.Context, appName string) (*entities.App, error)
	DeleteApp(ctx context.Context, appName string) (bool, error)

	// config

	CreateConfig(ctx context.Context, config entities.Config) (*entities.Config, error)
	UpdateConfig(ctx context.Context, config entities.Config) (bool, error)
	GetConfigs(ctx context.Context, namespace string) ([]*entities.Config, error)
	GetConfig(ctx context.Context, namespace, name string) (*entities.Config, error)
	DeleteConfig(ctx context.Context, namespace, name string) (bool, error)

	// secrets

	CreateSecret(ctx context.Context, secret entities.Secret) (*entities.Secret, error)
	UpdateSecret(ctx context.Context, secret entities.Secret) (bool, error)
	DeleteSecret(ctx context.Context, namespace, name string) (bool, error)
	GetSecrets(ctx context.Context, namespace string) ([]*entities.Secret, error)
	GetSecret(ctx context.Context, namespace, name string) (*entities.Secret, error)

	//OnUpdateApp(ctx context.Context, r *op_crds.StatusUpdate) error
	//OnDeleteApp(ctx context.Context, r *op_crds.StatusUpdate) error

	// router

	GetRouters(ctx context.Context, namespace string) ([]*entities.Router, error)
	GetRouter(ctx context.Context, namespace string, name string) (*entities.Router, error)
	DeleteRouter(ctx context.Context, namespace string, name string) (bool, error)
	CreateRouter(ctx context.Context, router entities.Router) (*entities.Router, error)
	UpdateRouter(ctx context.Context, router entities.Router) (bool, error)

	// managed service

	GetManagedServiceTemplates(ctx context.Context) ([]*entities.ManagedServiceCategory, error)
	GetManagedServiceTemplate(ctx context.Context, name string) (*entities.ManagedServiceTemplate, error)
	InstallManagedSvc(ctx context.Context, msvc entities.ManagedService) (*entities.ManagedService, error)
	UnInstallManagedSvc(ctx context.Context, namespace, name string) (bool, error)
	GetManagedSvc(ctx context.Context, namespace string, name string) (*entities.ManagedService, error)
	UpdateManagedSvc(ctx context.Context, msvc entities.ManagedService) (bool, error)
	GetManagedSvcOutput(ctx context.Context, namespace, name string) (map[string]any, error)
	GetManagedSvcs(ctx context.Context, namespace string) ([]*entities.ManagedService, error)
	//OnUpdateManagedSvc(ctx context.Context, r *op_crds.StatusUpdate) error
	//OnDeleteManagedService(todo context.Context, o *op_crds.StatusUpdate) error

	// managed resources

	GetManagedRes(ctx context.Context, namespace string, name string) (*entities.ManagedResource, error)
	GetManagedResOutput(ctx context.Context, namespace string, name string) (map[string]any, error)
	GetManagedResources(ctx context.Context, namespace string) ([]*entities.ManagedResource, error)
	GetManagedResourcesOfService(ctx context.Context, msvcNamespace string, msvcName string) ([]*entities.ManagedResource, error)

	InstallManagedRes(ctx context.Context, mres entities.ManagedResource) (*entities.ManagedResource, error)
	UpdateManagedRes(ctx context.Context, mres entities.ManagedResource) (bool, error)
	UnInstallManagedRes(ctx context.Context, namespace string, name string) (bool, error)
	//OnDeleteManagedResource(todo context.Context, o *op_crds.StatusUpdate) error
	//OnUpdateManagedRes(ctx context.Context, r *op_crds.StatusUpdate) error

	GetSocketCtx(
		conn *fWebsocket.Conn,
		cacheClient AuthCacheClient,
		cookieName,
		cookieDomain string,
		sessionKeyPrefix string,
	) context.Context
}

type AuthCacheClient cache.Client

const (
	ActionApply  string = "apply"
	ActionDelete string = "delete"
)

type WorkloadMessenger interface {
	SendAction(action string, kafkaTopic string, resId string, res any) error
}
