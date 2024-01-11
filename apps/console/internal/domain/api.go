package domain

import (
	"context"
	"time"

	"github.com/kloudlite/operator/operators/resource-watcher/types"

	"github.com/kloudlite/api/apps/console/internal/entities"
	"github.com/kloudlite/api/pkg/repos"
	wgv1 "github.com/kloudlite/operator/apis/wireguard/v1"
)

type ConsoleContext struct {
	context.Context
	AccountName string

	UserId    repos.ID
	UserEmail string
	UserName  string
}

func (c ConsoleContext) GetAccountName() string {
	return c.AccountName
}

type ResourceContext struct {
	ConsoleContext
	ProjectName     string
	EnvironmentName string
}

func (r ResourceContext) DBFilters() repos.Filter {
	return repos.Filter{
		"accountName":     r.AccountName,
		"projectName":     r.ProjectName,
		"environmentName": r.EnvironmentName,
	}
}

type UserAndAccountsContext struct {
	context.Context
	AccountName string
	UserId      repos.ID
}

func NewConsoleContext(parent context.Context, userId repos.ID, accountName string) ConsoleContext {
	return ConsoleContext{
		Context: parent,
		UserId:  userId,

		AccountName: accountName,
	}
}

type CheckNameAvailabilityOutput struct {
	Result         bool     `json:"result"`
	SuggestedNames []string `json:"suggestedNames,omitempty"`
}

type ConfigKeyRef struct {
	ConfigName string `json:"configName"`
	Key        string `json:"key"`
}

type ConfigKeyValueRef struct {
	ConfigName string `json:"configName"`
	Key        string `json:"key"`
	Value      string `json:"value"`
}

type SecretKeyRef struct {
	SecretName string `json:"secretName"`
	Key        string `json:"key"`
}

type SecretKeyValueRef struct {
	SecretName string `json:"secretName"`
	Key        string `json:"key"`
	Value      string `json:"value"`
}

type ManagedResourceKeyRef struct {
	MresName string `json:"mresName"`
	Key      string `json:"key"`
}

type ManagedResourceKeyValueRef struct {
	MresName string `json:"mresName"`
	Key      string `json:"key"`
	Value    string `json:"value"`
}

type ResType string

type UpdateAndDeleteOpts struct {
	MessageTimestamp time.Time
}

type Domain interface {
	CheckNameAvailability(ctx context.Context, resType entities.ResourceType, accountName string, namespace *string, name string) (*CheckNameAvailabilityOutput, error)

	ListProjects(ctx context.Context, userId repos.ID, accountName string, clusterName *string, search map[string]repos.MatchFilter, pagination repos.CursorPagination) (*repos.PaginatedRecord[*entities.Project], error)
	GetProject(ctx ConsoleContext, name string) (*entities.Project, error)

	CreateProject(ctx ConsoleContext, project entities.Project) (*entities.Project, error)
	UpdateProject(ctx ConsoleContext, project entities.Project) (*entities.Project, error)
	DeleteProject(ctx ConsoleContext, name string) error

	OnProjectApplyError(ctx ConsoleContext, errMsg string, name string, opts UpdateAndDeleteOpts) error
	OnProjectDeleteMessage(ctx ConsoleContext, project entities.Project) error
	OnProjectUpdateMessage(ctx ConsoleContext, cluster entities.Project, status types.ResourceStatus, opts UpdateAndDeleteOpts) error

	ResyncProject(ctx ConsoleContext, name string) error

	ListEnvironments(ctx ConsoleContext, projectName string, search map[string]repos.MatchFilter, pq repos.CursorPagination) (*repos.PaginatedRecord[*entities.Environment], error)
	GetEnvironment(ctx ConsoleContext, projectName string, name string) (*entities.Environment, error)

	CreateEnvironment(ctx ConsoleContext, projectName string, env entities.Environment) (*entities.Environment, error)
	CloneEnvironment(ctx ConsoleContext, projectName string, sourceEnvName string, envName string) (*entities.Environment, error)
	UpdateEnvironment(ctx ConsoleContext, projectName string, env entities.Environment) (*entities.Environment, error)
	DeleteEnvironment(ctx ConsoleContext, projectName string, name string) error

	OnEnvironmentApplyError(ctx ConsoleContext, errMsg, namespace, name string, opts UpdateAndDeleteOpts) error
	OnEnvironmentDeleteMessage(ctx ConsoleContext, env entities.Environment) error
	OnEnvironmentUpdateMessage(ctx ConsoleContext, env entities.Environment, status types.ResourceStatus, opts UpdateAndDeleteOpts) error

	ResyncEnvironment(ctx ConsoleContext, projectName string, name string) error

	ListApps(ctx ResourceContext, search map[string]repos.MatchFilter, pq repos.CursorPagination) (*repos.PaginatedRecord[*entities.App], error)
	GetApp(ctx ResourceContext, name string) (*entities.App, error)

	CreateApp(ctx ResourceContext, app entities.App) (*entities.App, error)
	UpdateApp(ctx ResourceContext, app entities.App) (*entities.App, error)
	DeleteApp(ctx ResourceContext, name string) error

	InterceptApp(ctx ResourceContext, appName string, deviceName string, intercept bool) (bool, error)

	OnAppApplyError(ctx ResourceContext, errMsg string, name string, opts UpdateAndDeleteOpts) error
	OnAppDeleteMessage(ctx ResourceContext, app entities.App) error
	OnAppUpdateMessage(ctx ResourceContext, app entities.App, status types.ResourceStatus, opts UpdateAndDeleteOpts) error

	ResyncApp(ctx ResourceContext, name string) error

	ListConfigs(ctx ResourceContext, search map[string]repos.MatchFilter, pq repos.CursorPagination) (*repos.PaginatedRecord[*entities.Config], error)
	GetConfig(ctx ResourceContext, name string) (*entities.Config, error)
	GetConfigEntries(ctx ResourceContext, keyrefs []ConfigKeyRef) ([]*ConfigKeyValueRef, error)

	CreateConfig(ctx ResourceContext, config entities.Config) (*entities.Config, error)
	UpdateConfig(ctx ResourceContext, config entities.Config) (*entities.Config, error)
	DeleteConfig(ctx ResourceContext, name string) error

	OnConfigApplyError(ctx ResourceContext, errMsg, name string, opts UpdateAndDeleteOpts) error
	OnConfigDeleteMessage(ctx ResourceContext, config entities.Config) error
	OnConfigUpdateMessage(ctx ResourceContext, config entities.Config, status types.ResourceStatus, opts UpdateAndDeleteOpts) error

	ResyncConfig(ctx ResourceContext, name string) error

	ListSecrets(ctx ResourceContext, search map[string]repos.MatchFilter, pq repos.CursorPagination) (*repos.PaginatedRecord[*entities.Secret], error)
	GetSecret(ctx ResourceContext, name string) (*entities.Secret, error)
	GetSecretEntries(ctx ResourceContext, keyrefs []SecretKeyRef) ([]*SecretKeyValueRef, error)

	CreateSecret(ctx ResourceContext, secret entities.Secret) (*entities.Secret, error)
	UpdateSecret(ctx ResourceContext, secret entities.Secret) (*entities.Secret, error)
	DeleteSecret(ctx ResourceContext, name string) error

	OnSecretApplyError(ctx ResourceContext, errMsg, name string, opts UpdateAndDeleteOpts) error
	OnSecretDeleteMessage(ctx ResourceContext, secret entities.Secret) error
	OnSecretUpdateMessage(ctx ResourceContext, secret entities.Secret, status types.ResourceStatus, opts UpdateAndDeleteOpts) error

	ResyncSecret(ctx ResourceContext, name string) error

	ListRouters(ctx ResourceContext, search map[string]repos.MatchFilter, pq repos.CursorPagination) (*repos.PaginatedRecord[*entities.Router], error)
	GetRouter(ctx ResourceContext, name string) (*entities.Router, error)

	CreateRouter(ctx ResourceContext, router entities.Router) (*entities.Router, error)
	UpdateRouter(ctx ResourceContext, router entities.Router) (*entities.Router, error)
	DeleteRouter(ctx ResourceContext, name string) error

	OnRouterApplyError(ctx ResourceContext, errMsg string, name string, opts UpdateAndDeleteOpts) error
	OnRouterDeleteMessage(ctx ResourceContext, router entities.Router) error
	OnRouterUpdateMessage(ctx ResourceContext, router entities.Router, status types.ResourceStatus, opts UpdateAndDeleteOpts) error

	ResyncRouter(ctx ResourceContext, name string) error

	ListManagedResources(ctx ResourceContext, search map[string]repos.MatchFilter, pq repos.CursorPagination) (*repos.PaginatedRecord[*entities.ManagedResource], error)
	GetManagedResource(ctx ResourceContext, name string) (*entities.ManagedResource, error)

  GetManagedResourceOutputKeys(ctx ResourceContext, name string) ([]string, error)
  GetManagedResourceOutputKVs(ctx ResourceContext, keyrefs []ManagedResourceKeyRef) ([]*ManagedResourceKeyValueRef, error)

	CreateManagedResource(ctx ResourceContext, mres entities.ManagedResource) (*entities.ManagedResource, error)
	UpdateManagedResource(ctx ResourceContext, mres entities.ManagedResource) (*entities.ManagedResource, error)
	DeleteManagedResource(ctx ResourceContext, name string) error

	OnManagedResourceApplyError(ctx ResourceContext, errMsg string, name string, opts UpdateAndDeleteOpts) error
	OnManagedResourceDeleteMessage(ctx ResourceContext, mres entities.ManagedResource) error
	OnManagedResourceUpdateMessage(ctx ResourceContext, mres entities.ManagedResource, status types.ResourceStatus, opts UpdateAndDeleteOpts) error

	ResyncManagedResource(ctx ResourceContext, name string) error

	// image pull secrets
	ListImagePullSecrets(ctx ResourceContext, search map[string]repos.MatchFilter, pagination repos.CursorPagination) (*repos.PaginatedRecord[*entities.ImagePullSecret], error)
	GetImagePullSecret(ctx ResourceContext, name string) (*entities.ImagePullSecret, error)
	CreateImagePullSecret(ctx ResourceContext, secret entities.ImagePullSecret) (*entities.ImagePullSecret, error)
	DeleteImagePullSecret(ctx ResourceContext, name string) error

	OnImagePullSecretApplyError(ctx ResourceContext, errMsg string, name string, opts UpdateAndDeleteOpts) error
	OnImagePullSecretDeleteMessage(ctx ResourceContext, ips entities.ImagePullSecret) error
	OnImagePullSecretUpdateMessage(ctx ResourceContext, ips entities.ImagePullSecret, status types.ResourceStatus, opts UpdateAndDeleteOpts) error

	ResyncImagePullSecret(ctx ResourceContext, name string) error

	GetResourceMapping(ctx ConsoleContext, resType entities.ResourceType, namespace string, name string) (*entities.ResourceMapping, error)

	ListProjectManagedServices(ctx ConsoleContext, projectName string, mf map[string]repos.MatchFilter, pagination repos.CursorPagination) (*repos.PaginatedRecord[*entities.ProjectManagedService], error)
	GetProjectManagedService(ctx ConsoleContext, projectName string, serviceName string) (*entities.ProjectManagedService, error)
	CreateProjectManagedService(ctx ConsoleContext, projectName string, service entities.ProjectManagedService) (*entities.ProjectManagedService, error)
	UpdateProjectManagedService(ctx ConsoleContext, projectName string, service entities.ProjectManagedService) (*entities.ProjectManagedService, error)
	DeleteProjectManagedService(ctx ConsoleContext, projectName string, name string) error
	OnProjectManagedServiceApplyError(ctx ConsoleContext, projectName, name, errMsg string, opts UpdateAndDeleteOpts) error
	OnProjectManagedServiceDeleteMessage(ctx ConsoleContext, projectName string, service entities.ProjectManagedService) error
	OnProjectManagedServiceUpdateMessage(ctx ConsoleContext, projectName string, service entities.ProjectManagedService, status types.ResourceStatus, opts UpdateAndDeleteOpts) error
	ResyncProjectManagedService(ctx ConsoleContext, projectName, name string) error

	ListVPNDevices(ctx ConsoleContext, search map[string]repos.MatchFilter, pagination repos.CursorPagination) (*repos.PaginatedRecord[*entities.ConsoleVPNDevice], error)
	GetVPNDevice(ctx ConsoleContext, name string) (*entities.ConsoleVPNDevice, error)
	CreateVPNDevice(ctx ConsoleContext, device entities.ConsoleVPNDevice) (*entities.ConsoleVPNDevice, error)
	UpdateVPNDevice(ctx ConsoleContext, device entities.ConsoleVPNDevice) (*entities.ConsoleVPNDevice, error)
	DeleteVPNDevice(ctx ConsoleContext, name string) error

	UpdateVpnDevicePorts(ctx ConsoleContext, devName string, ports []*wgv1.Port) error
	UpdateVpnDeviceNs(ctx ConsoleContext, devName string, namespace string) error
}

type PublishMsg string

const (
	PublishAdd    PublishMsg = "added"
	PublishDelete PublishMsg = "deleted"
	PublishUpdate PublishMsg = "updated"
)

type ResourceEventPublisher interface {
	PublishAppEvent(app *entities.App, msg PublishMsg)
	PublishMresEvent(mres *entities.ManagedResource, msg PublishMsg)
	PublishProjectEvent(project *entities.Project, msg PublishMsg)
	PublishProjectManagedServiceEvent(project *entities.ProjectManagedService, msg PublishMsg)
	PublishRouterEvent(router *entities.Router, msg PublishMsg)
	PublishWorkspaceEvent(workspace *entities.Environment, msg PublishMsg)
	PublishVpnDeviceEvent(device *entities.ConsoleVPNDevice, msg PublishMsg)
}
