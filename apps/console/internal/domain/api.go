package domain

import (
	"context"
	"github.com/kloudlite/operator/operators/resource-watcher/types"
	"time"

	"github.com/kloudlite/api/apps/console/internal/entities"
	"github.com/kloudlite/api/pkg/repos"
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
	ResTypeManagedResource ResType = "managed_resource"
)

type UpdateAndDeleteOpts struct {
	MessageTimestamp time.Time
}

type Domain interface {
	CheckNameAvailability(ctx context.Context, resType ResType, accountName string, namespace *string, name string) (*CheckNameAvailabilityOutput, error)

	ListProjects(ctx context.Context, userId repos.ID, accountName string, clusterName *string, search map[string]repos.MatchFilter, pagination repos.CursorPagination) (*repos.PaginatedRecord[*entities.Project], error)
	GetProject(ctx ConsoleContext, name string) (*entities.Project, error)

	CreateProject(ctx ConsoleContext, project entities.Project) (*entities.Project, error)
	UpdateProject(ctx ConsoleContext, project entities.Project) (*entities.Project, error)
	DeleteProject(ctx ConsoleContext, name string) error

	OnProjectApplyError(ctx ConsoleContext, errMsg string, name string, opts UpdateAndDeleteOpts) error
	OnProjectDeleteMessage(ctx ConsoleContext, project entities.Project) error
	OnProjectUpdateMessage(ctx ConsoleContext, cluster entities.Project, status types.ResourceStatus, opts UpdateAndDeleteOpts) error

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

	OnWorkspaceApplyError(ctx ConsoleContext, errMsg, namespace, name string, opts UpdateAndDeleteOpts) error
	OnWorkspaceDeleteMessage(ctx ConsoleContext, workspace entities.Workspace) error
	OnWorkspaceUpdateMessage(ctx ConsoleContext, workspace entities.Workspace, status types.ResourceStatus, opts UpdateAndDeleteOpts) error

	ResyncWorkspace(ctx ConsoleContext, namespace, name string) error

	ListApps(ctx ConsoleContext, namespace string, search map[string]repos.MatchFilter, pq repos.CursorPagination) (*repos.PaginatedRecord[*entities.App], error)
	GetApp(ctx ConsoleContext, namespace, name string) (*entities.App, error)

	CreateApp(ctx ConsoleContext, app entities.App) (*entities.App, error)
	UpdateApp(ctx ConsoleContext, app entities.App) (*entities.App, error)
	DeleteApp(ctx ConsoleContext, namespace, name string) error

	OnAppApplyError(ctx ConsoleContext, errMsg string, namespace string, name string, opts UpdateAndDeleteOpts) error
	OnAppDeleteMessage(ctx ConsoleContext, app entities.App) error
	OnAppUpdateMessage(ctx ConsoleContext, app entities.App, status types.ResourceStatus, opts UpdateAndDeleteOpts) error

	ResyncApp(ctx ConsoleContext, namespace, name string) error

	ListConfigs(ctx ConsoleContext, namespace string, search map[string]repos.MatchFilter, pq repos.CursorPagination) (*repos.PaginatedRecord[*entities.Config], error)
	GetConfig(ctx ConsoleContext, namespace, name string) (*entities.Config, error)

	CreateConfig(ctx ConsoleContext, config entities.Config) (*entities.Config, error)
	UpdateConfig(ctx ConsoleContext, config entities.Config) (*entities.Config, error)
	DeleteConfig(ctx ConsoleContext, namespace, name string) error

	OnConfigApplyError(ctx ConsoleContext, errMsg, namespace, name string, opts UpdateAndDeleteOpts) error
	OnConfigDeleteMessage(ctx ConsoleContext, config entities.Config) error
	OnConfigUpdateMessage(ctx ConsoleContext, config entities.Config, status types.ResourceStatus, opts UpdateAndDeleteOpts) error

	ResyncConfig(ctx ConsoleContext, namespace, name string) error

	ListSecrets(ctx ConsoleContext, namespace string, search map[string]repos.MatchFilter, pq repos.CursorPagination) (*repos.PaginatedRecord[*entities.Secret], error)
	GetSecret(ctx ConsoleContext, namespace, name string) (*entities.Secret, error)

	CreateSecret(ctx ConsoleContext, secret entities.Secret) (*entities.Secret, error)
	UpdateSecret(ctx ConsoleContext, secret entities.Secret) (*entities.Secret, error)
	DeleteSecret(ctx ConsoleContext, namespace, name string) error

	OnSecretApplyError(ctx ConsoleContext, errMsg, namespace, name string, opts UpdateAndDeleteOpts) error
	OnSecretDeleteMessage(ctx ConsoleContext, secret entities.Secret) error
	OnSecretUpdateMessage(ctx ConsoleContext, secret entities.Secret, status types.ResourceStatus, opts UpdateAndDeleteOpts) error

	ResyncSecret(ctx ConsoleContext, namespace, name string) error

	ListRouters(ctx ConsoleContext, namespace string, search map[string]repos.MatchFilter, pq repos.CursorPagination) (*repos.PaginatedRecord[*entities.Router], error)
	GetRouter(ctx ConsoleContext, namespace, name string) (*entities.Router, error)

	CreateRouter(ctx ConsoleContext, router entities.Router) (*entities.Router, error)
	UpdateRouter(ctx ConsoleContext, router entities.Router) (*entities.Router, error)
	DeleteRouter(ctx ConsoleContext, namespace, name string) error

	OnRouterApplyError(ctx ConsoleContext, errMsg string, namespace string, name string, opts UpdateAndDeleteOpts) error
	OnRouterDeleteMessage(ctx ConsoleContext, router entities.Router) error
	OnRouterUpdateMessage(ctx ConsoleContext, router entities.Router, status types.ResourceStatus, opts UpdateAndDeleteOpts) error

	ResyncRouter(ctx ConsoleContext, namespace, name string) error

	ListManagedResources(ctx ConsoleContext, namespace string, search map[string]repos.MatchFilter, pq repos.CursorPagination) (*repos.PaginatedRecord[*entities.ManagedResource], error)
	GetManagedResource(ctx ConsoleContext, namespace, name string) (*entities.ManagedResource, error)

	CreateManagedResource(ctx ConsoleContext, mres entities.ManagedResource) (*entities.ManagedResource, error)
	UpdateManagedResource(ctx ConsoleContext, mres entities.ManagedResource) (*entities.ManagedResource, error)
	DeleteManagedResource(ctx ConsoleContext, namespace, name string) error

	OnManagedResourceApplyError(ctx ConsoleContext, errMsg string, namespace string, name string, opts UpdateAndDeleteOpts) error
	OnManagedResourceDeleteMessage(ctx ConsoleContext, mres entities.ManagedResource) error
	OnManagedResourceUpdateMessage(ctx ConsoleContext, mres entities.ManagedResource, status types.ResourceStatus, opts UpdateAndDeleteOpts) error

	ResyncManagedResource(ctx ConsoleContext, namespace, name string) error

	// image pull secrets

	ListImagePullSecrets(ctx ConsoleContext, namespace string, search map[string]repos.MatchFilter, pagination repos.CursorPagination) (*repos.PaginatedRecord[*entities.ImagePullSecret], error)
	GetImagePullSecret(ctx ConsoleContext, namespace string, name string) (*entities.ImagePullSecret, error)
	CreateImagePullSecret(ctx ConsoleContext, secret entities.ImagePullSecret) (*entities.ImagePullSecret, error)
	DeleteImagePullSecret(ctx ConsoleContext, namespace string, name string) error

	OnImagePullSecretApplyError(ctx ConsoleContext, errMsg string, namespace string, name string, opts UpdateAndDeleteOpts) error
	OnImagePullSecretDeleteMessage(ctx ConsoleContext, ips entities.ImagePullSecret) error
	OnImagePullSecretUpdateMessage(ctx ConsoleContext, ips entities.ImagePullSecret, status types.ResourceStatus, opts UpdateAndDeleteOpts) error

	ResyncImagePullSecret(ctx ConsoleContext, namespace, name string) error
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
	PublishRouterEvent(router *entities.Router, msg PublishMsg)
	PublishWorkspaceEvent(workspace *entities.Workspace, msg PublishMsg)
}
