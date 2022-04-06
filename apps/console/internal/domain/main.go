package domain

import (
	"context"
	"fmt"

	"go.uber.org/fx"
	"golang.zx2c4.com/wireguard/wgctrl/wgtypes"
	"kloudlite.io/apps/console/internal/domain/entities"
	"kloudlite.io/pkg/config"
	"kloudlite.io/pkg/errors"
	"kloudlite.io/pkg/messaging"
	"kloudlite.io/pkg/repos"
)

type domain struct {
	deviceRepo      repos.DbRepo[*entities.Device]
	clusterRepo     repos.DbRepo[*entities.Cluster]
	messageProducer messaging.Producer[messaging.Json]
	messageTopic    string
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

	d.messageProducer.SendMessage(d.messageTopic, string(c.Id), messaging.Json{
		"cluster_id":    c.Id,
		"region":        c.Region,
		"provider":      c.Provider,
		"masters_count": 1,
		"nodes_count":   2,
	})

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
	})
	errors.AssertNoError(e, fmt.Errorf("unable to create new device"))

	return newDevice, e
}

func (d *domain) RemoveDevice(ctx context.Context, deviceId repos.ID) error {
	return d.deviceRepo.DeleteById(ctx, deviceId)
}

func (d *domain) ListClusterDevices(ctx context.Context, clusterId repos.ID) ([]*entities.Device, error) {
	return d.deviceRepo.Find(ctx, repos.Query{
		Filter: repos.Filter{
			"cluster_id": clusterId,
		},
	})
}

func (d *domain) ListUserDevices(ctx context.Context, userId repos.ID) ([]*entities.Device, error) {
	fmt.Println(userId)
	return d.deviceRepo.Find(ctx, repos.Query{
		Filter: repos.Filter{
			"user_id": userId,
		},
	})
}

func (d *domain) GetCluster(ctx context.Context, id repos.ID) (*entities.Cluster, error) {
	return d.clusterRepo.FindById(ctx, id)
}

func (d *domain) GetDevice(ctx context.Context, id repos.ID) (*entities.Device, error) {
	return d.deviceRepo.FindById(ctx, id)
}

type Env struct {
	KafkaInfraTopic string `env:KAFKA_INFRA_TOPIC required:"true"`
}

var Module = fx.Module(
	"domain",
	fx.Provide(func() (*Env, error) {
		var envC Env
		err := config.LoadConfigFromEnv(&envC)
		if err != nil {
			fmt.Println(err, "failed to load env")
			return nil, fmt.Errorf("not able to load ENV: %v", err)
		}
		return &envC, err
	}),
	fx.Provide(
		func(deviceRepo repos.DbRepo[*entities.Device], clusterRepo repos.DbRepo[*entities.Cluster], msgP messaging.Producer[messaging.Json], env *Env) Domain {
			return &domain{
				deviceRepo:      deviceRepo,
				clusterRepo:     clusterRepo,
				messageProducer: msgP,
				messageTopic:    env.KafkaInfraTopic,
			}
		}),
)
