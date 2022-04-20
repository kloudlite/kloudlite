package domain

import (
	"context"
	"kloudlite.io/apps/console/internal/domain/entities"
	"kloudlite.io/pkg/repos"
)

type Domain interface {
	GetDevice(ctx context.Context, id repos.ID) (*entities.Device, error)
	GetCluster(ctx context.Context, id repos.ID) (*entities.Cluster, error)
	CreateCluster(
		ctx context.Context,
		data *entities.Cluster,
	) (*entities.Cluster, error)
	UpdateCluster(
		ctx context.Context,
		id repos.ID,
		name *string,
		nodeCount *int,
	) (*entities.Cluster, error)
	DeleteCluster(
		ctx context.Context,
		clusterId repos.ID,
	) error

	ListClusters(ctx context.Context, accountId repos.ID) ([]*entities.Cluster, error)

	AddDevice(
		ctx context.Context,
		deviceName string,
		clusterId repos.ID,
		userId repos.ID,
	) (dev *entities.Device, e error)

	RemoveDevice(
		ctx context.Context,
		deviceId repos.ID,
	) error

	ListClusterDevices(ctx context.Context, clusterId repos.ID) ([]*entities.Device, error)

	ListUserDevices(ctx context.Context, userId repos.ID) ([]*entities.Device, error)

	OnSetupCluster(cxt context.Context, response entities.SetupClusterResponse) error
	OnDeleteCluster(cxt context.Context, response entities.DeleteClusterResponse) error
	OnUpdateCluster(cxt context.Context, response entities.UpdateClusterResponse) error
	OnAddPeer(cxt context.Context, response entities.AddPeerResponse) error
	OnDeletePeer(cxt context.Context, response entities.DeletePeerResponse) error
	CreateProject(ctx context.Context, id repos.ID, projectName string, displayName string, logo *string, description *string) (*entities.Project, error)
	GetAccountProjects(ctx context.Context, id repos.ID) ([]*entities.Project, error)
	GetProjectWithID(ctx context.Context, projectId repos.ID) (*entities.Project, error)
	CreateConfig(ctx context.Context, id repos.ID, configName string, desc *string, configData []*entities.Entry) (*entities.Config, error)
	UpdateConfig(ctx context.Context, configId repos.ID, desc *string, configData []*entities.Entry) (bool, error)

	CreateSecret(ctx context.Context, projectId repos.ID, secretName string, desc *string, secretData []*entities.Entry) (*entities.Secret, error)
	UpdateSecret(ctx context.Context, secretId repos.ID, desc *string, secretData []*entities.Entry) (bool, error)

	GetConfigs(ctx context.Context, projectId repos.ID) ([]*entities.Config, error)
	GetConfig(ctx context.Context, configId repos.ID) (*entities.Config, error)

	GetSecrets(ctx context.Context, projectId repos.ID) ([]*entities.Secret, error)
	GetSecret(ctx context.Context, secretId repos.ID) (*entities.Secret, error)
	CreateRouter(ctx context.Context, id repos.ID, routerName string, domains []string, routes []*entities.Route) (*entities.Router, error)
	GetRouters(ctx context.Context, projectID repos.ID) ([]*entities.Router, error)
	GetRouter(ctx context.Context, routerID repos.ID) (*entities.Router, error)
	UpdateRouter(ctx context.Context, id repos.ID, domains []string, entries []*entities.Route) (bool, error)
}

type InfraActionMessage interface {
	entities.SetupClusterAction | entities.DeleteClusterAction | entities.UpdateClusterAction | entities.AddPeerAction | entities.DeletePeerAction
}

type InfraMessenger interface {
	SendAction(action any) error
}

func SendAction[T InfraActionMessage](i InfraMessenger, action T) error {
	return i.SendAction(action)
}
