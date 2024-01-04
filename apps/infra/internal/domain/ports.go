package domain

import (
	"context"
	"github.com/kloudlite/api/apps/infra/internal/entities"
	"github.com/kloudlite/api/grpc-interfaces/kloudlite.io/rpc/accounts"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type AccountsSvc interface {
	GetAccount(ctx context.Context, userId string, accountName string) (*accounts.GetAccountOut, error)
}

type ResourceDispatcher interface {
	ApplyToTargetCluster(ctx InfraContext, clusterName string, obj client.Object, recordVersion int) error
	DeleteFromTargetCluster(ctx InfraContext, clusterName string, obj client.Object) error
}

type PublishMsg string

const (
	PublishAdd    PublishMsg = "added"
	PublishDelete PublishMsg = "deleted"
	PublishUpdate PublishMsg = "updated"
)

type ResourceEventPublisher interface {
	PublishClusterEvent(cluster *entities.Cluster, msg PublishMsg)
	PublishNodePoolEvent(np *entities.NodePool, msg PublishMsg)
	PublishVpnDeviceEvent(dev *entities.VPNDevice, msg PublishMsg)
	PublishDomainResEvent(domain *entities.DomainEntry, msg PublishMsg)
	PublishPvcResEvent(pvc *entities.PersistentVolumeClaim, msg PublishMsg)
	PublishCMSEvent(cms *entities.ClusterManagedService, msg PublishMsg)
	PublishHelmReleaseEvent(hr *entities.HelmRelease, msg PublishMsg)
	PublishPvResEvent(pv *entities.PersistentVolume, msg PublishMsg)
	PublishVolumeAttachmentEvent(volatt *entities.VolumeAttachment, msg PublishMsg)
}
