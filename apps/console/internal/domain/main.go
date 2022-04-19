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

func (d *domain) OnSetupCluster(ctx context.Context, response entities.SetupClusterResponse) error {
	byId, err := d.clusterRepo.FindById(ctx, response.ClusterID)
	if err != nil {
		return err
	}
	if response.Done {
		byId.Status = entities.ClusterStateLive
	} else {
		byId.Status = entities.ClusterStateError
	}

	if response.PublicIp != "" {
		byId.Ip = &response.PublicIp
	}
	if response.PublicKey != "" {
		byId.PublicKey = &response.PublicKey
	}
	_, err = d.clusterRepo.UpdateById(ctx, response.ClusterID, byId)
	if err != nil {
		return err
	}
	return nil
}

func (d *domain) OnDeleteCluster(cxt context.Context, response entities.DeleteClusterResponse) error {
	byId, err := d.clusterRepo.FindById(cxt, response.ClusterID)
	if err != nil {
		return err
	}
	if response.Done {
		byId.Status = entities.ClusterStateDown
	} else {
		byId.Status = entities.ClusterStateError
	}
	_, err = d.clusterRepo.UpdateById(cxt, response.ClusterID, byId)
	if err != nil {
		return err
	}
	return nil
}

func (d *domain) OnUpdateCluster(cxt context.Context, response entities.UpdateClusterResponse) error {
	byId, err := d.clusterRepo.FindById(cxt, response.ClusterID)
	if err != nil {
		return err
	}
	if response.Done {
		byId.Status = entities.ClusterStateLive
	} else {
		byId.Status = entities.ClusterStateError
	}
	_, err = d.clusterRepo.UpdateById(cxt, response.ClusterID, byId)
	if err != nil {
		return err
	}
	return nil
}

func (d *domain) OnAddPeer(cxt context.Context, response entities.AddPeerResponse) error {
	byId, err := d.deviceRepo.FindById(cxt, response.DeviceID)
	if err != nil {
		return err
	}
	if response.Done {
		byId.Status = entities.DeviceStateAttached
	} else {
		byId.Status = entities.DeviceStateError
	}
	_, err = d.deviceRepo.UpdateById(cxt, response.ClusterID, byId)
	if err != nil {
		return err
	}
	return nil
}

func (d *domain) OnDeletePeer(cxt context.Context, response entities.DeletePeerResponse) error {
	byId, err := d.deviceRepo.FindById(cxt, response.DeviceID)
	if err != nil {
		return err
	}
	if response.Done {
		byId.Status = entities.DeviceStateDeleted
	} else {
		byId.Status = entities.DeviceStateError
	}
	_, err = d.deviceRepo.UpdateById(cxt, response.ClusterID, byId)
	if err != nil {
		return err
	}
	return nil
}

func (d *domain) CreateCluster(ctx context.Context, data *entities.Cluster) (cluster *entities.Cluster, e error) {
	data.Status = entities.ClusterStateSyncing
	c, err := d.clusterRepo.Create(ctx, data)
	if err != nil {
		return nil, err
	}
	err = SendAction(
		d.infraMessenger,
		entities.SetupClusterAction{
			ClusterID:  c.Id,
			Region:     c.Region,
			Provider:   c.Provider,
			NodesCount: c.NodesCount,
		},
	)
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
		err = SendAction(d.infraMessenger, entities.UpdateClusterAction{
			ClusterID:  id,
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
	err = SendAction(d.infraMessenger, entities.DeleteClusterAction{
		ClusterID: clusterId,
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

	e = SendAction(d.infraMessenger, entities.AddPeerAction{
		ClusterID: clusterId,
		PublicKey: pbKeyString,
		PeerIp:    ip,
	})
	if e != nil {
		return nil, e
	}

	return newDevice, e
}

func (d *domain) RemoveDevice(ctx context.Context, deviceId repos.ID) error {
	device, err := d.deviceRepo.FindById(ctx, deviceId)
	if err != nil {
		return err
	}
	device.Status = entities.DeviceStateSyncing
	_, err = d.deviceRepo.UpdateById(ctx, deviceId, device)
	if err != nil {
		return err
	}
	err = SendAction(d.infraMessenger, entities.DeletePeerAction{
		ClusterID: device.ClusterId,
		DeviceID:  device.Id,
		PublicKey: *device.PublicKey,
	})
	if err != nil {
		return err
	}
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
