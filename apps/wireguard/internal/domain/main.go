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
	deviceRepo  repos.DbRepo[*entities.Device]
	clusterRepo repos.DbRepo[*entities.Cluster]
}

func (d *domain) SetupCluster(
	ctx context.Context,
	clusterId repos.ID,
	address string,
	port uint16,
	netInterface string,
) (cluster *entities.Cluster, e error) {
	defer errors.HandleErr(&e)
	c, err := d.clusterRepo.FindById(ctx, clusterId)

	if err != nil {
		errors.AssertNoError(e, fmt.Errorf("cluster not found"))
	}
	c.ListenPort = &port
	c.NetInterface = &netInterface
	c.Address = &address
	updatedCluster, err := d.clusterRepo.UpdateById(ctx, clusterId, c)
	if err != nil {
		errors.AssertNoError(e, fmt.Errorf("failed to update cluster"))
	}
	return updatedCluster, err
}

func (d *domain) CreateCluster(ctx context.Context, data entities.Cluster) (cluster *entities.Cluster, e error) {
	defer errors.HandleErr(&e)
	pk, e := wgtypes.GeneratePrivateKey()
	errors.AssertNoError(e, fmt.Errorf("could not generate wg privateKey"))
	s := pk.String()
	sPub := pk.PublicKey().String()
	data.PrivateKey = &s
	data.PublicKey = &sPub
	c, err := d.clusterRepo.Create(ctx, &data)
	return c, err
}

func (d *domain) DeleteCluster(ctx context.Context, clusterId repos.ID) error {
	//_, e := d.clusterRepo.DeleteById(ctx, clusterId)
	//return e
	return nil
}

func (d *domain) ListClusters(ctx context.Context) ([]*entities.Cluster, error) {
	return d.clusterRepo.Find(ctx, repos.Query{})
}

func (d *domain) AddDevice(ctx context.Context, deviceName string, clusterId repos.ID, userId repos.ID) (dev *entities.Device, e error) {
	defer errors.HandleErr(&e)
	cluster, e := d.clusterRepo.FindById(ctx, clusterId)
	fmt.Println(cluster)
	errors.AssertNoError(e, fmt.Errorf("cluster is not ready"))
	pk, e := wgtypes.GeneratePrivateKey()
	pkString := pk.String()
	pbKeyString := pk.PublicKey().String()
	errors.AssertNoError(e, fmt.Errorf("could not generate wg private key"))
	newDevice, e := d.deviceRepo.Create(ctx, &entities.Device{
		Name:       deviceName,
		ClusterId:  clusterId,
		UserId:     userId,
		PrivateKey: &pkString,
		PublicKey:  &pbKeyString,
		Peers: map[string]entities.Peer{
			string(cluster.Id): {
				Id:        cluster.Id,
				Address:   cluster.Address,
				PublicKey: cluster.PublicKey,
			},
		},
	})
	errors.AssertNoError(e, fmt.Errorf("unable to create new device"))
	cluster.Peers[newDevice.Id] = entities.Peer{
		Id:        newDevice.Id,
		PublicKey: newDevice.PublicKey,
	}
	_, e = d.clusterRepo.UpdateById(ctx, cluster.Id, cluster)
	errors.AssertNoError(e, fmt.Errorf("unable to update cluster"))
	return newDevice, e
}

func (d *domain) RemoveDevice(ctx context.Context, deviceId repos.ID) error {
	//_, e := d.deviceRepo.DeleteById(ctx, deviceId)
	return nil
}

func (d *domain) ListDevices(ctx context.Context) ([]*entities.Device, error) {
	return d.deviceRepo.Find(ctx, repos.Query{})
}

var Module = fx.Module(
	"domain",
	fx.Provide(
		func(deviceRepo repos.DbRepo[*entities.Device], clusterRepo repos.DbRepo[*entities.Cluster]) Domain {
			return &domain{
				deviceRepo,
				clusterRepo,
			}
		}),
)
