package domain

import (
	"context"
	"fmt"
	"go.uber.org/fx"
	"golang.zx2c4.com/wireguard/wgctrl/wgtypes"
	"kloudlite.io/apps/console/internal/domain/entities"
	"kloudlite.io/pkg/config"
	"kloudlite.io/pkg/errors"
	"kloudlite.io/pkg/logger"
	"kloudlite.io/pkg/messaging"
	"kloudlite.io/pkg/repos"
)

type domain struct {
	deviceRepo      repos.DbRepo[*entities.Device]
	clusterRepo     repos.DbRepo[*entities.Cluster]
	messageProducer messaging.Producer[messaging.Json]
	messageTopic    string
	logger          logger.Logger
	messenger       InfraMessenger
}

func (d *domain) UpdateClusterState(ctx context.Context, id repos.ID, status entities.ClusterStatus, PublicIp string, PublicKey string) (bool, error) {
	byId, err := d.clusterRepo.FindById(ctx, id)
	if err != nil {
		return false, err
	}
	byId.Status = status
	updateById, err := d.clusterRepo.UpdateById(ctx, id, byId)
	if err != nil {
		return false, err
	}
	return updateById.Status == status, nil
}

func (d *domain) UpdateDeviceState(ctx context.Context, id repos.ID, status entities.DeviceStatus) (bool, error) {
	byId, err := d.deviceRepo.FindById(ctx, id)
	if err != nil {
		return false, err
	}
	byId.Status = status
	updateById, err := d.deviceRepo.UpdateById(ctx, id, byId)
	if err != nil {
		return false, err
	}
	return updateById.Status == status, nil
}

func (d *domain) ClusterDown(ctx context.Context, id repos.ID) (bool, error) {
	byId, err := d.clusterRepo.FindById(ctx, id)
	if err != nil {
		return false, err
	}
	byId.Status = entities.ClusterStateDown
	updateById, err := d.clusterRepo.UpdateById(ctx, id, byId)
	if err != nil {
		return false, err
	}
	return updateById.Status == entities.ClusterStateDown, nil
}

func (d *domain) ClusterUp(ctx context.Context, id repos.ID) (bool, error) {
	byId, err := d.clusterRepo.FindById(ctx, id)
	if err != nil {
		return false, err
	}
	byId.Status = entities.ClusterStateSyncing
	updateById, err := d.clusterRepo.UpdateById(ctx, id, byId)
	if err != nil {
		return false, err
	}
	// TODO Send Message
	return updateById.Status == entities.ClusterStateSyncing, nil
}

func (d *domain) CreateCluster(ctx context.Context, data entities.Cluster) (cluster *entities.Cluster, e error) {
	c, err := d.clusterRepo.Create(ctx, &data)
	if err != nil {
		return nil, err
	}
	return c, e
}

func (d *domain) UpdateCluster(
	ctx context.Context,
	id repos.ID,
	name *string,
	nodeCount *int,
) (*entities.Cluster, error) {
	c, err := d.clusterRepo.FindById(ctx, id)
	if err != nil {
		return nil, err
	}
	if name != nil {
		c.Name = *name
	}
	if nodeCount != nil {
		c.NodesCount = *nodeCount
	}
	updated, err := d.clusterRepo.UpdateById(ctx, id, c)
	if err != nil {
		return nil, err
	}
	return updated, nil
}

func (d *domain) DeleteCluster(ctx context.Context, clusterId repos.ID) error {
	// TODO
	fmt.Println(clusterId)
	return d.clusterRepo.DeleteById(ctx, clusterId)
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
	KafkaInfraTopic string `env:"KAFKA_INFRA_TOPIC" required:"true"`
}

func fxDomain(
	deviceRepo repos.DbRepo[*entities.Device],
	clusterRepo repos.DbRepo[*entities.Cluster],
	msgP messaging.Producer[messaging.Json],
	env *Env,
	logger logger.Logger,
	messenger InfraMessenger,
) Domain {
	return &domain{
		messenger:       messenger,
		deviceRepo:      deviceRepo,
		clusterRepo:     clusterRepo,
		messageProducer: msgP,
		messageTopic:    env.KafkaInfraTopic,
		logger:          logger,
	}
}

var Module = fx.Module(
	"domain",
	fx.Provide(config.LoadEnv[Env]()),
	fx.Provide(fxDomain),
)
