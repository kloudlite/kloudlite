package app

import (
	"fmt"
	"github.com/kloudlite/api/apps/infra/internal/domain"
	"github.com/kloudlite/api/apps/infra/internal/entities"
	"github.com/kloudlite/api/pkg/logging"
	"github.com/kloudlite/api/pkg/nats"
)

type ResourceEventPublisherImpl struct {
	cli    *nats.Client
	logger logging.Logger
}

func (r *ResourceEventPublisherImpl) publish(subject string, msg domain.PublishMsg) {
	if err := r.cli.Conn.Publish(subject, []byte(msg)); err != nil {
		r.logger.Errorf(err, "failed to publish message to subject %q", subject)
	}
}

func (r *ResourceEventPublisherImpl) PublishHelmReleaseEvent(hr *entities.HelmRelease, msg domain.PublishMsg) {
	subject := fmt.Sprintf(
		"res-updates.account.%s.cluster.%s",
		hr.AccountName, hr.ClusterName,
	)

	r.publish(subject, msg)
}

func (r *ResourceEventPublisherImpl) PublishClusterEvent(cluster *entities.Cluster, msg domain.PublishMsg) {
	subject := fmt.Sprintf(
		"res-updates.account.%s.cluster.%s",
		cluster.AccountName, cluster.Cluster.Name,
	)

	r.publish(subject, msg)
}

func (r *ResourceEventPublisherImpl) PublishNodePoolEvent(np *entities.NodePool, msg domain.PublishMsg) {
	subject := fmt.Sprintf(
		"res-updates.account.%s.cluster.%s.node-pool.%s",
		np.AccountName, np.ClusterName, np.Name,
	)

	r.publish(subject, msg)
}

func (r *ResourceEventPublisherImpl) PublishVpnDeviceEvent(dev *entities.VPNDevice, msg domain.PublishMsg) {
	subject := fmt.Sprintf(
		"res-updates.account.%s.cluster.%s.vpn-device.%s",
		dev.AccountName, dev.ClusterName, dev.Name,
	)

	r.publish(subject, msg)
}

func (r *ResourceEventPublisherImpl) PublishDomainResEvent(domain *entities.DomainEntry, msg domain.PublishMsg) {
	subject := fmt.Sprintf(
		"res-updates.account.%s.cluster.%s.domain.%s",
		domain.AccountName, domain.ClusterName, domain.DomainName,
	)

	r.publish(subject, msg)
}

func (r *ResourceEventPublisherImpl) PublishPvcResEvent(pvc *entities.PersistentVolumeClaim, msg domain.PublishMsg) {
	subject := fmt.Sprintf(
		"res-updates.account.%s.cluster.%s.vpn-device.%s",
		pvc.AccountName, pvc.ClusterName, pvc.Name,
	)

	r.publish(subject, msg)
}

func (r *ResourceEventPublisherImpl) PublishCMSEvent(cms *entities.ClusterManagedService, msg domain.PublishMsg) {
	subject := fmt.Sprintf(
		"res-updates.account.%s.cluster.%s.cluster-managed-service.%s",
		cms.AccountName, cms.ClusterName, cms.Name,
	)

	r.publish(subject, msg)
}

func NewResourceEventPublisher(cli *nats.Client, logger logging.Logger) domain.ResourceEventPublisher {
	return &ResourceEventPublisherImpl{
		cli,
		logger,
	}
}
