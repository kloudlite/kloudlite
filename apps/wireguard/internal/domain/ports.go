package domain

import (
	"context"

	"kloudlite.io/apps/wireguard/internal/domain/entities"
	"kloudlite.io/pkg/repos"
)

type Domain interface {
	CreateCluster(
		ctx context.Context,
		data entities.Cluster,
	) (*entities.Cluster, error)
	DeleteCluster(
		ctx context.Context,
		clusterId repos.ID,
	) error
	ListClusters(ctx context.Context) ([]*entities.Cluster, error)
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
	ListDevices(ctx context.Context) ([]*entities.Device, error)
	SetupCluster(
		ctx context.Context,
		clusterId repos.ID,
		address string,
		listenPort uint16,
		netInterface string,
	) (*entities.Cluster, error)
}
