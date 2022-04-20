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
