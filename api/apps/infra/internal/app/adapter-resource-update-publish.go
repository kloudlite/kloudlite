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

func (r *ResourceEventPublisherImpl) PublishClusterEvent(cluster *entities.Cluster, msg domain.PublishMsg) {
	subject := clusterResUpdateSubject(cluster)
	if err := r.cli.Conn.Publish(subject, []byte(msg)); err != nil {
		r.logger.Errorf(err, "failed to publish message to subject %q", subject)
	}
}

func (r *ResourceEventPublisherImpl) PublishNodePoolEvent(np *entities.NodePool, msg domain.PublishMsg) {
	subject := nodePoolResUpdateSubject(np)
	if err := r.cli.Conn.Publish(subject, []byte(msg)); err != nil {
		r.logger.Errorf(err, "failed to publish message to subject %q", subject)
	}
}

func (r *ResourceEventPublisherImpl) PublishVpnDeviceEvent(dev *entities.VPNDevice, msg domain.PublishMsg) {
	subject := vpnDeviceResUpdateSubject(dev)
	if err := r.cli.Conn.Publish(subject, []byte(msg)); err != nil {
		r.logger.Errorf(err, "failed to publish message to subject %q", subject)
	}
}

func (r *ResourceEventPublisherImpl) PublishDomainResEvent(domain *entities.DomainEntry, msg domain.PublishMsg) {
	subject := domainResUpdateSubject(domain)
	if err := r.cli.Conn.Publish(subject, []byte(msg)); err != nil {
		r.logger.Errorf(err, "failed to publish message to subject %q", subject)
	}
}

func (r *ResourceEventPublisherImpl) PublishPvcResEvent(pvc *entities.PersistentVolumeClaim, msg domain.PublishMsg) {
	subject := pvcResUpdateSubject(pvc)
	if err := r.cli.Conn.Publish(subject, []byte(msg)); err != nil {
		r.logger.Errorf(err, "failed to publish message to subject %q", subject)
	}
}

func (r *ResourceEventPublisherImpl) PublishCMSEvent(cms *entities.ClusterManagedService, msg domain.PublishMsg) {
	subject := clusterManagedServiceUpdateSubject(cms)
	if err := r.cli.Conn.Publish(subject, []byte(msg)); err != nil {
		r.logger.Errorf(err, "failed to publish message to subject %q", subject)
	}
}

func NewResourceEventPublisher(cli *nats.Client, logger logging.Logger) domain.ResourceEventPublisher {
	return &ResourceEventPublisherImpl{
		cli,
		logger,
	}
}

func clusterResUpdateSubject(cluster *entities.Cluster) string {
	return fmt.Sprint(
		"res-updates.",
		"account.",
		cluster.Cluster.Spec.AccountName, ".",
		"cluster.",
		cluster.Cluster.Name)
}

func nodePoolResUpdateSubject(nodePool *entities.NodePool) string {
	return fmt.Sprint(
		"res-updates.",
		"account.",
		nodePool.AccountName, ".",
		"cluster.",
		nodePool.ClusterName, ".",
		"node-pool.", nodePool.Name,
	)
}

func domainResUpdateSubject(domainEntry *entities.DomainEntry) string {
	return fmt.Sprint(
		"res-updates.",
		"account.",
		domainEntry.AccountName, ".",
		"cluster.",
		domainEntry.ClusterName, ".",
		"domain.", domainEntry.DomainName,
	)
}

func vpnDeviceResUpdateSubject(device *entities.VPNDevice) string {
	return fmt.Sprint(
		"res-updates.",
		"account.",
		device.AccountName, ".",
		"cluster.",
		device.ClusterName, ".",
		"vpn-device.", device.Name,
	)
}

func pvcResUpdateSubject(pvc *entities.PersistentVolumeClaim) string {
	return fmt.Sprint(
		"res-updates.",
		"account.",
		pvc.AccountName, ".",
		"cluster.",
		pvc.ClusterName, ".",
		"vpn-device.", pvc.Name,
	)
}

func clusterManagedServiceUpdateSubject(cmsvc *entities.ClusterManagedService) string {
	return fmt.Sprint(
		"res-updates.",
		"account.",
		cmsvc.AccountName, ".",
		"cluster.",
		cmsvc.ClusterName, ".",
		"cluster-managed-service.", cmsvc.Name,
	)
}
