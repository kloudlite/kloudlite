package domain

import (
	"context"
	"fmt"

	"go.uber.org/fx"
	"golang.zx2c4.com/wireguard/wgctrl/wgtypes"
	"kloudlite.io/apps/wireguard/internal/domain/entities"
	"kloudlite.io/pkg/errors"
	"kloudlite.io/pkg/repos"
)

type domain struct {
	deviceRepo  repos.DbRepo[entities.Device]
	clusterRepo repos.DbRepo[entities.Cluster]
}

func (d *domain) CreateCluster(ctx context.Context, data entities.Cluster) (cluster entities.Cluster, e error) {
	defer errors.HandleErr(&e)
	pk, e := wgtypes.GeneratePrivateKey()
	errors.AssertNoError(e, fmt.Errorf("could not generate wg privateKey"))
	s := pk.String()
	data.PrivateKey = &s
	return d.clusterRepo.Create(ctx, data)
}

func (d *domain) DeleteCluster(ctx context.Context, clusterId repos.ID) error {
	_, e := d.clusterRepo.DeleteById(ctx, clusterId)
	return e
}

func (d *domain) ListClusters(ctx context.Context) ([]entities.Cluster, error) {
	return d.clusterRepo.Find(ctx, repos.Query{})
}

func (d *domain) AddDevice(ctx context.Context, data entities.Device) (dev entities.Device, e error) {
	defer errors.HandleErr(&e)
	pk, e := wgtypes.GeneratePrivateKey()
	errors.AssertNoError(e, fmt.Errorf("could not generate wg private key"))
	data.PrivateKey = pk.String()
	data.PublicKey = pk.PublicKey().String()
	return d.deviceRepo.Create(ctx, data)
}

func (d *domain) RemoveDevice(ctx context.Context, deviceId repos.ID) error {
	_, e := d.deviceRepo.DeleteById(ctx, deviceId)
	return e
}

func (d *domain) ListDevices(ctx context.Context) ([]entities.Device, error) {
	return d.deviceRepo.Find(ctx, repos.Query{})
}

var Module = fx.Module(
	"domain",
	fx.Provide(func(deviceRepo repos.DbRepo[entities.Device], clusterRepo repos.DbRepo[entities.Cluster]) Domain {
		return &domain{
			deviceRepo,
			clusterRepo,
		}
	}),
)
