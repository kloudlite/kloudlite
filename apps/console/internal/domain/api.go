package domain

import (
	"context"
	"time"

	fc "github.com/kloudlite/api/apps/console/internal/entities/field-constants"

	"github.com/kloudlite/api/common/fields"

	crdsv1 "github.com/kloudlite/operator/apis/crds/v1"

	"github.com/kloudlite/operator/operators/resource-watcher/types"

	"github.com/kloudlite/api/apps/console/internal/entities"
	"github.com/kloudlite/api/pkg/repos"

	networkingv1 "github.com/kloudlite/operator/apis/networking/v1"
	corev1 "k8s.io/api/core/v1"
)

type ConsoleContext struct {
	context.Context
	AccountName string

	UserId    repos.ID
	UserEmail string
	UserName  string
}

func (c ConsoleContext) GetUserId() repos.ID {
	return c.UserId
}

func (c ConsoleContext) GetUserEmail() string {
	return c.UserEmail
}

func (c ConsoleContext) GetUserName() string {
	return c.UserName
}

func (c ConsoleContext) GetAccountName() string {
	return c.AccountName
}

type ResourceContext struct {
	ConsoleContext
	EnvironmentName string
}

type ManagedResourceContext struct {
	ConsoleContext
	ManagedServiceName *string
	EnvironmentName    *string
}

func (m ManagedResourceContext) MresDBFilters() (*repos.Filter, error) {
	return &repos.Filter{
		fields.AccountName:                   m.AccountName,
		fc.ManagedResourceManagedServiceName: m.ManagedServiceName,
	}, nil
}

func (r ResourceContext) DBFilters() repos.Filter {
	return repos.Filter{
		fields.AccountName:     r.AccountName,
		fields.EnvironmentName: r.EnvironmentName,
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

func NewManagedResourceContext(ctx ConsoleContext, msvcName string) ManagedResourceContext {
	return ManagedResourceContext{
		ConsoleContext:     ctx,
		ManagedServiceName: &msvcName,
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
	ClusterName      string
}

type AccountsSvc interface {
	GetAccountRegion(ctx context.Context, userId string, accountName string) (string, error)
}

type ResourceDispatcher interface {
	ApplyResource()
	DeleteResource()
	RestartResource()
}

type Domain interface {
	AccountsSvc

	CheckNameAvailability(ctx context.Context, accountName string, environmentName *string, msvcName *string, resType entities.ResourceType, name string) (*CheckNameAvailabilityOutput, error)

	// INFO: project have been disabled
	// ListProjects(ctx context.Context, userId repos.ID, accountName string, search map[string]repos.MatchFilter, pagination repos.CursorPagination) (*repos.PaginatedRecord[*entities.Project], error)
	// GetProject(ctx ConsoleContext, name string) (*entities.Project, error)
	//
	// CreateProject(ctx ConsoleContext, project entities.Project) (*entities.Project, error)
	// UpdateProject(ctx ConsoleContext, project entities.Project) (*entities.Project, error)
	// DeleteProject(ctx ConsoleContext, name string) error
	//
	// OnProjectApplyError(ctx ConsoleContext, errMsg string, name string, opts UpdateAndDeleteOpts) error
	// OnProjectDeleteMessage(ctx ConsoleContext, project entities.Project) error
	// OnProjectUpdateMessage(ctx ConsoleContext, cluster entities.Project, status types.ResourceStatus, opts UpdateAndDeleteOpts) error

	// ResyncProject(ctx ConsoleContext, name string) error

	ListEnvironments(ctx ConsoleContext, search map[string]repos.MatchFilter, pq repos.CursorPagination) (*repos.PaginatedRecord[*entities.Environment], error)
	GetEnvironment(ctx ConsoleContext, name string) (*entities.Environment, error)

	SetupDefaultEnvTemplate(ctx ConsoleContext) error
	CreateEnvironment(ctx ConsoleContext, env entities.Environment) (*entities.Environment, error)
	CloneEnvironment(ctx ConsoleContext, args CloneEnvironmentArgs) (*entities.Environment, error)
	UpdateEnvironment(ctx ConsoleContext, env entities.Environment) (*entities.Environment, error)
	DeleteEnvironment(ctx ConsoleContext, name string) error
	ArchiveEnvironmentsForCluster(ctx ConsoleContext, clusterName string) (bool, error)

	OnEnvironmentApplyError(ctx ConsoleContext, errMsg, namespace, name string, opts UpdateAndDeleteOpts) error
	OnEnvironmentDeleteMessage(ctx ConsoleContext, env entities.Environment) error
	OnEnvironmentUpdateMessage(ctx ConsoleContext, env entities.Environment, status types.ResourceStatus, opts UpdateAndDeleteOpts) error

	ResyncEnvironment(ctx ConsoleContext, name string) error

	GetRegistryImageURL(ctx ConsoleContext) (*entities.RegistryImageURL, error)
	GetRegistryImage(ctx ConsoleContext, image string) (*entities.RegistryImage, error)
	DeleteRegistryImage(ctx ConsoleContext, image string) error
	UpsertRegistryImage(ctx context.Context, accountName string, image string, meta map[string]any) (*entities.RegistryImage, error)
	ListRegistryImages(ctx ConsoleContext, pq repos.CursorPagination) (*repos.PaginatedRecord[*entities.RegistryImage], error)
	SearchRegistryImages(ctx ConsoleContext, query string) ([]*entities.RegistryImage, error)

	ListHelmCharts(ctx ResourceContext, search map[string]repos.MatchFilter, pq repos.CursorPagination) (*repos.PaginatedRecord[*entities.HelmChart], error)
	GetHelmChart(ctx ResourceContext, name string) (*entities.HelmChart, error)

	CreateHelmChart(ctx ResourceContext, app entities.HelmChart) (*entities.HelmChart, error)
	UpdateHelmChart(ctx ResourceContext, app entities.HelmChart) (*entities.HelmChart, error)
	DeleteHelmChart(ctx ResourceContext, name string) error

	ListApps(ctx ResourceContext, search map[string]repos.MatchFilter, pq repos.CursorPagination) (*repos.PaginatedRecord[*entities.App], error)
	GetApp(ctx ResourceContext, name string) (*entities.App, error)

	CreateApp(ctx ResourceContext, app entities.App) (*entities.App, error)
	UpdateApp(ctx ResourceContext, app entities.App) (*entities.App, error)
	DeleteApp(ctx ResourceContext, name string) error

	InterceptApp(ctx ResourceContext, appName string, deviceName string, intercept bool, portMappings []crdsv1.AppInterceptPortMappings) (bool, error)
	InterceptAppOnLocalCluster(ctx ResourceContext, appName string, clusterName string, ipAddr string, intercept bool, portMappings []crdsv1.AppInterceptPortMappings) (bool, error)
	RestartApp(ctx ResourceContext, appName string) error
	RemoveDeviceIntercepts(ctx ResourceContext, deviceName string) error

	OnHelmChartApplyError(ctx ResourceContext, errMsg string, name string, opts UpdateAndDeleteOpts) error
	OnHelmChartDeleteMessage(ctx ResourceContext, app entities.HelmChart) error
	OnHelmChartUpdateMessage(ctx ResourceContext, app entities.HelmChart, status types.ResourceStatus, opts UpdateAndDeleteOpts) error

	OnAppApplyError(ctx ResourceContext, errMsg string, name string, opts UpdateAndDeleteOpts) error
	OnAppDeleteMessage(ctx ResourceContext, app entities.App) error
	OnAppUpdateMessage(ctx ResourceContext, app entities.App, status types.ResourceStatus, opts UpdateAndDeleteOpts) error

	ResyncApp(ctx ResourceContext, name string) error
	RolloutAppsByImage(ctx ConsoleContext, imageName string) error

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

	ListManagedResources(ctx ConsoleContext, search map[string]repos.MatchFilter, pq repos.CursorPagination) (*repos.PaginatedRecord[*entities.ManagedResource], error)
	GetManagedResource(ctx ManagedResourceContext, name string) (*entities.ManagedResource, error)
	GetManagedResourceByID(ctx ConsoleContext, id repos.ID) (*entities.ManagedResource, error)

	// ListImportedManagedResources(ctx ResourceContext, search map[string]repos.MatchFilter, pq repos.CursorPagination) (*repos.PaginatedRecord[*entities.ManagedResource], error)
	// GetImportedManagedResource(ctx ResourceContext, name string) (*entities.ManagedResource, error)

	GetManagedResourceOutputKeys(ctx ManagedResourceContext, name string) ([]string, error)
	GetImportedManagedResourceOutputKeys(ctx ResourceContext, name string) ([]string, error)
	GetManagedResourceOutputKVs(ctx ManagedResourceContext, keyrefs []ManagedResourceKeyRef) ([]*ManagedResourceKeyValueRef, error)
	GetImportedManagedResourceOutputKVs(ctx ResourceContext, keyrefs []ManagedResourceKeyRef) ([]*ManagedResourceKeyValueRef, error)

	CreateManagedResource(ctx ManagedResourceContext, mres entities.ManagedResource) (*entities.ManagedResource, error)
	UpdateManagedResource(ctx ManagedResourceContext, mres entities.ManagedResource) (*entities.ManagedResource, error)
	DeleteManagedResource(ctx ManagedResourceContext, name string) error

	// ImportManagedResource(ctx ManagedResourceContext, mresName string, importName string) (*entities.ManagedResource, error)
	ImportManagedResource(ctx ManagedResourceContext, mresName string, importName string) (*entities.ImportedManagedResource, error)
	// ListImportedManagedResources(ctx ConsoleContext, search map[string]repos.MatchFilter, pq repos.CursorPagination) (*repos.PaginatedRecord[*entities.ManagedResource], error)
	ListImportedManagedResources(ctx ConsoleContext, envName string, search map[string]repos.MatchFilter, pq repos.CursorPagination) (*repos.PaginatedRecord[*entities.ImportedManagedResource], error)
	DeleteImportedManagedResource(ctx ResourceContext, importName string) error

	// ImportManagedResource(ctx ConsoleContext, imr entities.ImportedManagedResource) (*entities.ImportedManagedResource, error)
	// DeleteImportedManagedResource(ctx ConsoleContext, name string) error

	// OnImportedManagedResourceDeleteMessage(ctx ConsoleContext, secret *corev1.Secret) error
	// OnImportedManagedResourceUpdateMessage(ctx ConsoleContext, secret *corev1.Secret, status types.ResourceStatus, opts UpdateAndDeleteOpts) error

	OnManagedResourceApplyError(ctx ConsoleContext, errMsg string, msvcName string, name string, opts UpdateAndDeleteOpts) error
	OnManagedResourceDeleteMessage(ctx ConsoleContext, msvcName string, mres crdsv1.ManagedResource) error
	OnManagedResourceUpdateMessage(ctx ConsoleContext, msvcName string, mres crdsv1.ManagedResource, outputSecret *corev1.Secret, status types.ResourceStatus, opts UpdateAndDeleteOpts) error

	ResyncManagedResource(ctx ConsoleContext, msvcName string, name string) error

	/// External Apps
	ListExternalApps(ctx ResourceContext, search map[string]repos.MatchFilter, pq repos.CursorPagination) (*repos.PaginatedRecord[*entities.ExternalApp], error)
	GetExternalApp(ctx ResourceContext, name string) (*entities.ExternalApp, error)

	CreateExternalApp(ctx ResourceContext, externalApp entities.ExternalApp) (*entities.ExternalApp, error)
	UpdateExternalApp(ctx ResourceContext, externalAppIn entities.ExternalApp) (*entities.ExternalApp, error)
	DeleteExternalApp(ctx ResourceContext, name string) error

	InterceptExternalApp(ctx ResourceContext, externalAppName string, deviceName string, intercept bool, portMappings []crdsv1.AppInterceptPortMappings) (bool, error)

	OnExternalAppApplyError(ctx ResourceContext, errMsg string, name string, opts UpdateAndDeleteOpts) error
	OnExternalAppDeleteMessage(ctx ResourceContext, externalApp entities.ExternalApp) error
	OnExternalAppUpdateMessage(ctx ResourceContext, externalApp entities.ExternalApp, status types.ResourceStatus, opts UpdateAndDeleteOpts) error

	ResyncExternalApp(ctx ResourceContext, name string) error

	// image pull secrets
	ListImagePullSecrets(ctx ConsoleContext, search map[string]repos.MatchFilter, pagination repos.CursorPagination) (*repos.PaginatedRecord[*entities.ImagePullSecret], error)
	GetImagePullSecret(ctx ConsoleContext, name string) (*entities.ImagePullSecret, error)
	CreateImagePullSecret(ctx ConsoleContext, secret entities.ImagePullSecret) (*entities.ImagePullSecret, error)
	UpdateImagePullSecret(ctx ConsoleContext, secret entities.ImagePullSecret) (*entities.ImagePullSecret, error)
	DeleteImagePullSecret(ctx ConsoleContext, name string) error

	OnImagePullSecretApplyError(ctx ConsoleContext, errMsg string, name string, opts UpdateAndDeleteOpts) error
	OnImagePullSecretDeleteMessage(ctx ConsoleContext, ips entities.ImagePullSecret) error
	OnImagePullSecretUpdateMessage(ctx ConsoleContext, ips entities.ImagePullSecret, status types.ResourceStatus, opts UpdateAndDeleteOpts) error

	ResyncImagePullSecret(ctx ConsoleContext, name string) error

	GetEnvironmentResourceMapping(ctx ConsoleContext, resType entities.ResourceType, clusterName string, namespace string, name string) (*entities.ResourceMapping, error)

	ServiceBinding
	ClusterManagedService
}

type ServiceBinding interface {
	OnServiceBindingUpdateMessage(ctx ConsoleContext, svcb *networkingv1.ServiceBinding, status types.ResourceStatus, opts UpdateAndDeleteOpts) error
	OnServiceBindingDeleteMessage(ctx ConsoleContext, svcb *networkingv1.ServiceBinding) error
}

type ClusterManagedService interface {
	ListClusterManagedServices(ctx ConsoleContext, search map[string]repos.MatchFilter, pagination repos.CursorPagination) (*repos.PaginatedRecord[*entities.ClusterManagedService], error)

	GetClusterManagedService(ctx ConsoleContext, serviceName string) (*entities.ClusterManagedService, error)

	CreateClusterManagedService(ctx ConsoleContext, cmsvc entities.ClusterManagedService) (*entities.ClusterManagedService, error)
	// CloneClusterManagedService(ctx ConsoleContext, args CloneManagedServiceArgs) (*entities.ClusterManagedService, error)
	UpdateClusterManagedService(ctx ConsoleContext, cmsvc entities.ClusterManagedService) (*entities.ClusterManagedService, error)
	DeleteClusterManagedService(ctx ConsoleContext, name string) error
	ArchiveClusterManagedServicesForCluster(ctx ConsoleContext, clusterName string) (bool, error)

	OnClusterManagedServiceApplyError(ctx ConsoleContext, clusterName, name, errMsg string, opts UpdateAndDeleteOpts) error
	OnClusterManagedServiceDeleteMessage(ctx ConsoleContext, clusterName string, service entities.ClusterManagedService) error
	OnClusterManagedServiceUpdateMessage(ctx ConsoleContext, clusterName string, service entities.ClusterManagedService, status types.ResourceStatus, opts UpdateAndDeleteOpts) error
}

type PublishMsg string

const (
	PublishAdd    PublishMsg = "added"
	PublishDelete PublishMsg = "deleted"
	PublishUpdate PublishMsg = "updated"
)

type ResourceEventPublisher interface {
	PublishConsoleEvent(ctx ConsoleContext, resourceType entities.ResourceType, name string, update PublishMsg)
	PublishEnvironmentResourceEvent(ctx ConsoleContext, envName string, resourceType entities.ResourceType, name string, update PublishMsg)
	PublishResourceEvent(ctx ResourceContext, resourceType entities.ResourceType, name string, update PublishMsg)
	PublishClusterManagedServiceEvent(ctx ConsoleContext, msvcName string, resourceType entities.ResourceType, name string, update PublishMsg)
}
