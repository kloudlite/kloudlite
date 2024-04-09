package domain

import (
	"context"
	"github.com/kloudlite/api/apps/iot-console/internal/entities"
	"github.com/kloudlite/api/common/fields"
	"github.com/kloudlite/api/pkg/repos"
)

type IotConsoleContext struct {
	context.Context
	AccountName string

	UserId    repos.ID
	UserEmail string
	UserName  string
}

type IotResourceContext struct {
	IotConsoleContext
	ProjectName string
	//EnvironmentName string
}

func (r IotResourceContext) IOTConsoleDBFilters() repos.Filter {
	return repos.Filter{
		fields.AccountName: r.AccountName,
		fields.ProjectName: r.ProjectName,
	}
}

func (i IotConsoleContext) GetUserId() repos.ID { return i.UserId }

func (i IotConsoleContext) GetUserEmail() string {
	return i.UserEmail
}

func (i IotConsoleContext) GetUserName() string {
	return i.UserName
}
func (i IotConsoleContext) GetAccountName() string { return i.AccountName }

type Domain interface {
	ListProjects(ctx IotConsoleContext, search map[string]repos.MatchFilter, pagination repos.CursorPagination) (*repos.PaginatedRecord[*entities.IOTProject], error)
	GetProject(ctx IotConsoleContext, name string) (*entities.IOTProject, error)

	CreateProject(ctx IotConsoleContext, project entities.IOTProject) (*entities.IOTProject, error)
	UpdateProject(ctx IotConsoleContext, project entities.IOTProject) (*entities.IOTProject, error)
	DeleteProject(ctx IotConsoleContext, name string) error

	ListDeployments(ctx IotResourceContext, search map[string]repos.MatchFilter, pagination repos.CursorPagination) (*repos.PaginatedRecord[*entities.IOTDeployment], error)
	GetDeployment(ctx IotResourceContext, name string) (*entities.IOTDeployment, error)

	CreateDeployment(ctx IotResourceContext, deployment entities.IOTDeployment) (*entities.IOTDeployment, error)
	UpdateDeployment(ctx IotResourceContext, deployment entities.IOTDeployment) (*entities.IOTDeployment, error)
	DeleteDeployment(ctx IotResourceContext, name string) error

	ListDevices(ctx IotResourceContext, deviceBlueprintName string, search map[string]repos.MatchFilter, pq repos.CursorPagination) (*repos.PaginatedRecord[*entities.IOTDevice], error)
	GetDevice(ctx IotResourceContext, name string, deviceBlueprintName string) (*entities.IOTDevice, error)
	ListDeploymentDevices(ctx IotResourceContext, deploymentName string, search map[string]repos.MatchFilter, pq repos.CursorPagination) (*repos.PaginatedRecord[*entities.IOTDevice], error)
	GetDeploymentDevice(ctx IotResourceContext, name string, deploymentName string) (*entities.IOTDevice, error)

	CreateDevice(ctx IotResourceContext, deviceBlueprintName string, device entities.IOTDevice) (*entities.IOTDevice, error)
	UpdateDevice(ctx IotResourceContext, deviceBlueprintName string, device entities.IOTDevice) (*entities.IOTDevice, error)
	DeleteDevice(ctx IotResourceContext, deviceBlueprintName string, name string) error

	AddDeviceToDeployment(ctx IotResourceContext, deploymentName string, deviceName string, deviceBlueprintName string) (*entities.IOTDevice, error)
	RemoveDeviceOfDeployment(ctx IotResourceContext, deploymentName string, deviceName string, deviceBlueprintName string) (*entities.IOTDevice, error)

	ListDeviceBlueprints(ctx IotResourceContext, search map[string]repos.MatchFilter, pq repos.CursorPagination) (*repos.PaginatedRecord[*entities.IOTDeviceBlueprint], error)
	GetDeviceBlueprint(ctx IotResourceContext, name string) (*entities.IOTDeviceBlueprint, error)

	CreateDeviceBlueprint(ctx IotResourceContext, deviceBlueprint entities.IOTDeviceBlueprint) (*entities.IOTDeviceBlueprint, error)
	UpdateDeviceBlueprint(ctx IotResourceContext, deviceBlueprint entities.IOTDeviceBlueprint) (*entities.IOTDeviceBlueprint, error)
	DeleteDeviceBlueprint(ctx IotResourceContext, name string) error

	ListApps(ctx IotResourceContext, deviceBlueprintName string, search map[string]repos.MatchFilter, pq repos.CursorPagination) (*repos.PaginatedRecord[*entities.IOTApp], error)
	GetApp(ctx IotResourceContext, deviceBlueprintName string, name string) (*entities.IOTApp, error)

	CreateApp(ctx IotResourceContext, deviceBlueprintName string, app entities.IOTApp) (*entities.IOTApp, error)
	UpdateApp(ctx IotResourceContext, deviceBlueprintName string, app entities.IOTApp) (*entities.IOTApp, error)
	DeleteApp(ctx IotResourceContext, deviceBlueprintName string, name string) error
}
