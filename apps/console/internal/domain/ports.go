package domain

import (
	"context"
	"kloudlite.io/apps/console/internal/domain/entities"
	op_crds "kloudlite.io/apps/console/internal/domain/op-crds"
	"kloudlite.io/pkg/repos"
)

type Domain interface {
	CreateCluster(ctx context.Context, data *entities.Cluster) (*entities.Cluster, error)
	UpdateCluster(ctx context.Context, id repos.ID, name *string, nodeCount *int) (bool, error)
	DeleteCluster(ctx context.Context, clusterId repos.ID) error
	GetCluster(ctx context.Context, id repos.ID) (*entities.Cluster, error)
	ListClusters(ctx context.Context, accountId repos.ID) ([]*entities.Cluster, error)
	OnSetupCluster(cxt context.Context, response entities.SetupClusterResponse) error
	OnUpdateCluster(cxt context.Context, response entities.UpdateClusterResponse) error

	GetDevice(ctx context.Context, id repos.ID) (*entities.Device, error)
	GetDeviceConfig(ctx context.Context, id repos.ID) (string, error)
	AddDevice(ctx context.Context, deviceName string, clusterId repos.ID, userId repos.ID) (dev *entities.Device, e error)
	RemoveDevice(ctx context.Context, deviceId repos.ID) error
	ListClusterDevices(ctx context.Context, clusterId repos.ID) ([]*entities.Device, error)
	ListUserDevices(ctx context.Context, userId repos.ID, clusterId *repos.ID) ([]*entities.Device, error)
	OnAddPeer(cxt context.Context, response entities.AddPeerResponse) error

	CreateProject(ctx context.Context, id repos.ID, projectName string, displayName string, logo *string, description *string) (*entities.Project, error)
	GetAccountProjects(ctx context.Context, id repos.ID) ([]*entities.Project, error)
	GetProjectWithID(ctx context.Context, projectId repos.ID) (*entities.Project, error)
	OnUpdateProject(ctx context.Context, response *op_crds.Project) error

	CreateConfig(ctx context.Context, id repos.ID, configName string, desc *string, configData []*entities.Entry) (*entities.Config, error)
	UpdateConfig(ctx context.Context, configId repos.ID, desc *string, configData []*entities.Entry) (bool, error)
	PatchConfig(ctx context.Context, configId repos.ID, desc *string, configData []*entities.Entry) (bool, error)
	GetConfigs(ctx context.Context, projectId repos.ID) ([]*entities.Config, error)
	GetConfig(ctx context.Context, configId repos.ID) (*entities.Config, error)
	DeleteConfig(ctx context.Context, configId repos.ID) (bool, error)
	OnUpdateConfig(ctx context.Context, configId repos.ID) error

	CreateSecret(ctx context.Context, projectId repos.ID, secretName string, desc *string, secretData []*entities.Entry) (*entities.Secret, error)
	UpdateSecret(ctx context.Context, secretId repos.ID, desc *string, secretData []*entities.Entry) (bool, error)
	PatchSecret(ctx context.Context, secretId repos.ID, desc *string, secretData []*entities.Entry) (bool, error)
	DeleteSecret(ctx context.Context, secretId repos.ID) (bool, error)
	GetSecrets(ctx context.Context, projectId repos.ID) ([]*entities.Secret, error)
	GetSecret(ctx context.Context, secretId repos.ID) (*entities.Secret, error)
	OnUpdateSecret(ctx context.Context, secretId repos.ID) error

	GetRouters(ctx context.Context, projectID repos.ID) ([]*entities.Router, error)
	GetRouter(ctx context.Context, routerID repos.ID) (*entities.Router, error)
	DeleteRouter(ctx context.Context, routerID repos.ID) (bool, error)
	CreateRouter(ctx context.Context, id repos.ID, routerName string, domains []string, routes []*entities.Route) (*entities.Router, error)
	UpdateRouter(ctx context.Context, id repos.ID, domains []string, entries []*entities.Route) (bool, error)
	OnUpdateRouter(ctx context.Context, r *op_crds.Router) error

	GetManagedSvc(ctx context.Context, managedSvcID repos.ID) (*entities.ManagedService, error)
	GetManagedSvcs(ctx context.Context, projectID repos.ID) ([]*entities.ManagedService, error)
	InstallManagedSvc(ctx context.Context, projectID repos.ID, templateID repos.ID, name string, values map[string]interface{}) (*entities.ManagedService, error)
	UpdateManagedSvc(ctx context.Context, managedServiceId repos.ID, values map[string]interface{}) (bool, error)
	UnInstallManagedSvc(ctx context.Context, managedServiceId repos.ID) (bool, error)
	OnUpdateManagedSvc(ctx context.Context, r *op_crds.ManagedService) error

	GetManagedRes(ctx context.Context, managedResID repos.ID) (*entities.ManagedResource, error)
	GetManagedResources(ctx context.Context, projectID repos.ID) ([]*entities.ManagedResource, error)
	GetManagedResourcesOfService(ctx context.Context, installationId repos.ID) ([]*entities.ManagedResource, error)

	InstallManagedRes(
		ctx context.Context,
		installationId repos.ID,
		name string,
		resourceType string,
		values map[string]string,
	) (*entities.ManagedResource, error)
	UpdateManagedRes(ctx context.Context, managedResID repos.ID, values map[string]string) (bool, error)
	UnInstallManagedRes(ctx context.Context, managedResID repos.ID) (bool, error)
	OnUpdateManagedRes(ctx context.Context, r *op_crds.ManagedResource) error

	GetApps(ctx context.Context, projectId repos.ID) ([]*entities.App, error)
	GetApp(ctx context.Context, projectID repos.ID) (*entities.App, error)
	UpdateApp(ctx context.Context, managedResID repos.ID, values map[string]interface{}) (bool, error)
	DeleteApp(ctx context.Context, appID repos.ID) (bool, error)
	OnUpdateApp(ctx context.Context, r *op_crds.App) error
	GetManagedServiceTemplates(ctx context.Context) ([]*entities.ManagedServiceCategory, error)
	InstallAppFlow(
		ctx context.Context,
		userId repos.ID,
		id repos.ID,
		app entities.AppIn,
	) (bool, error)

	GetResourceOutputs(ctx context.Context, managedResID repos.ID) (map[string]string, error)

	GetProjectMemberships(ctx context.Context, projectID repos.ID) ([]*entities.ProjectMembership, error)
	InviteProjectMember(ctx context.Context, projectID repos.ID, email string, role string) (bool, error)

	UpdateResourceStatus(ctx context.Context, resourceType string, resourceNamespace string, resourceName string, status ResourceStatus) (bool, error)
}

type InfraActionMessage interface {
	entities.SetupClusterAction | entities.DeleteClusterAction | entities.UpdateClusterAction | entities.AddPeerAction | entities.DeletePeerAction
}

type InfraMessenger interface {
	SendAction(action any) error
}

type WorkloadMessenger interface {
	SendAction(action string, resId string, res any) error
}

func SendAction[T InfraActionMessage](i InfraMessenger, action T) error {
	return i.SendAction(action)
}
