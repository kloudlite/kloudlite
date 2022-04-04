package domain

import (
	"context"

	"kloudlite.io/apps/wireguard/internal/domain/entities"
	"kloudlite.io/pkg/repos"
)

type Domain interface {
	CreateCluster(ctx context.Context, data entities.Cluster) (entities.Cluster, error)
	DeleteCluster(ctx context.Context, clusterId repos.ID) error
	ListClusters(ctx context.Context) ([]entities.Cluster, error)
	AddDevice(ctx context.Context, data entities.Device) (entities.Device, error)
	RemoveDevice(ctx context.Context, deviceId repos.ID) error
	ListDevices(ctx context.Context) ([]entities.Device, error)
}
