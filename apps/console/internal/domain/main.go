package domain

import (
	"context"
	"fmt"
	"go.uber.org/fx"
	"golang.zx2c4.com/wireguard/wgctrl/wgtypes"
	"kloudlite.io/apps/console/internal/domain/entities"
	"kloudlite.io/pkg/config"
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
	infraMessenger  InfraMessenger
}

func (d *domain) UpdateClusterState(
	ctx context.Context,
	id repos.ID,
	status entities.ClusterStatus,
	PublicIp *string,
	PublicKey *string,
) (bool, error) {
	byId, err := d.clusterRepo.FindById(ctx, id)
	if err != nil {
		return false, err
	}
	byId.Status = status
	if PublicIp != nil {
		byId.Ip = PublicIp
	}
	if PublicKey != nil {
		byId.PublicKey = PublicKey
	}
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

func (d *domain) RemoveDeviceDone(ctx context.Context, id repos.ID) error {
	return d.deviceRepo.DeleteById(ctx, id)
}

func (d *domain) _ClusterDown(ctx context.Context, id repos.ID) (bool, error) {
	byId, err := d.clusterRepo.FindById(ctx, id)
	if err != nil {
		return false, err
	}
	byId.Status = entities.ClusterStateDown
	updateById, err := d.clusterRepo.UpdateById(ctx, id, byId)
	if err != nil {
		return false, err
	}
	err = d.infraMessenger.SendDeleteClusterAction(entities.DeleteClusterAction{
		ClusterID: string(updateById.Id),
		Provider:  updateById.Provider,
	})
	if err != nil {
		return false, err
	}
	return updateById.Status == entities.ClusterStateDown, nil
}

func (d *domain) _ClusterUp(ctx context.Context, id repos.ID) (bool, error) {
	byId, err := d.clusterRepo.FindById(ctx, id)
	if err != nil {
		return false, err
	}
	byId.Status = entities.ClusterStateSyncing
	updateById, err := d.clusterRepo.UpdateById(ctx, id, byId)
	if err != nil {
		return false, err
	}
	err = d.infraMessenger.SendUpdateClusterAction(entities.UpdateClusterAction{
		ClusterID:  string(updateById.Id),
		Region:     updateById.Region,
		Provider:   updateById.Provider,
		NodesCount: updateById.NodesCount,
	})
	if err != nil {
		return false, err
	}
	return updateById.Status == entities.ClusterStateSyncing, nil
}

func (d *domain) CreateCluster(ctx context.Context, data *entities.Cluster) (cluster *entities.Cluster, e error) {
	data.Status = entities.ClusterStateSyncing
	c, err := d.clusterRepo.Create(ctx, data)
	if err != nil {
		return nil, err
	}
	err = d.infraMessenger.SendAddClusterAction(entities.SetupClusterAction{
		ClusterID:  string(c.Id),
		Region:     c.Region,
		Provider:   c.Provider,
		NodesCount: c.NodesCount,
	})
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
		c.Status = entities.ClusterStateSyncing
	}
	updated, err := d.clusterRepo.UpdateById(ctx, id, c)
	if err != nil {
		return nil, err
	}
	if c.Status == entities.ClusterStateSyncing {
		err = d.infraMessenger.SendUpdateClusterAction(entities.UpdateClusterAction{
			ClusterID:  string(id),
			Region:     updated.Region,
			Provider:   updated.Provider,
			NodesCount: updated.NodesCount,
		})
		if err != nil {
			return nil, err
		}
	}
	return updated, nil
}

func (d *domain) RemoveClusterDone(ctx context.Context, id repos.ID) error {
	return d.clusterRepo.DeleteById(ctx, id)
}

func (d *domain) DeleteCluster(ctx context.Context, clusterId repos.ID) error {
	cluster, err := d.clusterRepo.FindById(ctx, clusterId)
	if err != nil {
		return err
	}
	cluster.Status = entities.ClusterStateSyncing
	_, err = d.clusterRepo.UpdateById(ctx, clusterId, cluster)
	if err != nil {
		return err
	}
	err = d.infraMessenger.SendDeleteClusterAction(entities.DeleteClusterAction{
		ClusterID: string(clusterId),
		Provider:  cluster.Provider,
	})
	if err != nil {
		return err
	}
	return nil
}

func (d *domain) ListClusters(ctx context.Context, accountId repos.ID) ([]*entities.Cluster, error) {
	return d.clusterRepo.Find(ctx, repos.Query{
		Filter: repos.Filter{
			"account_id": accountId,
		},
	})
}

func (d *domain) AddDevice(ctx context.Context, deviceName string, clusterId repos.ID, userId repos.ID) (*entities.Device, error) {

	cluster, e := d.clusterRepo.FindById(ctx, clusterId)
	if e != nil {
		return nil, fmt.Errorf("unable to fetch cluster %v", e)
	}

	if cluster.PublicKey == nil {
		return nil, fmt.Errorf("cluster is not ready")
	}

	pk, e := wgtypes.GeneratePrivateKey()
	if e != nil {
		return nil, fmt.Errorf("unable to generate private key because %v", e)
	}
	pkString := pk.String()
	pbKeyString := pk.PublicKey().String()

	devices, err := d.deviceRepo.Find(ctx, repos.Query{
		Filter: repos.Filter{
			"cluster_id": clusterId,
		},
		Sort: map[string]any{
			"index": 1,
		},
	})

	if err != nil {
		return nil, err
	}

	index := -1
	count := 0
	for i, d := range devices {
		count++
		if d.Index != i {
			index = i
			break
		}
	}
	if index == -1 {
		index = count
	}

	ip := fmt.Sprintf("10.13.13.%v", index+51)
	newDevice, e := d.deviceRepo.Create(ctx, &entities.Device{
		Name:       deviceName,
		ClusterId:  clusterId,
		UserId:     userId,
		PrivateKey: &pkString,
		PublicKey:  &pbKeyString,
		Ip:         ip,
		Status:     entities.DeviceStateSyncing,
		Index:      index,
	})

	if e != nil {
		return nil, fmt.Errorf("unable to persist in db %v", e)
	}

	d.infraMessenger.SendAddDeviceAction(entities.AddPeerAction{
		ClusterID: string(clusterId),
		PublicKey: pbKeyString,
		PeerIp:    ip,
	})

	return newDevice, e
}

func (d *domain) RemoveDevice(ctx context.Context, deviceId repos.ID) error {
	device, err := d.deviceRepo.FindById(ctx, deviceId)
	if err != nil {
		return err
	}
	_, err = d.UpdateDeviceState(ctx, deviceId, entities.DeviceStateSyncing)
	if err != nil {
		return err
	}
	d.infraMessenger.SendRemoveDeviceAction(entities.DeletePeerAction{
		ClusterID: string(device.ClusterId),
		DeviceID:  string(device.Id),
		PublicKey: *device.PublicKey,
	})
	return err
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
		infraMessenger:  messenger,
		deviceRepo:      deviceRepo,
		clusterRepo:     clusterRepo,
		messageProducer: msgP,
		messageTopic:    env.KafkaInfraTopic,
		logger:          logger,
	}
}

var Module = fx.Module(
	"domain",
	config.EnvFx[Env](),
	fx.Provide(fxDomain),
)
