package domain

import (
	"fmt"
	"github.com/kloudlite/api/apps/infra/internal/entities"
)

func (d *domain) clusterResUpdateSubject(cluster *entities.Cluster) string {
	return fmt.Sprint(
		"res-updates.",
		"account.",
		cluster.Cluster.Spec.AccountName, ".",
		"cluster.",
		cluster.Cluster.Name)
}


func (d *domain) nodePoolResUpdateSubject(nodePool *entities.NodePool) string {
	return fmt.Sprint(
		"res-updates.",
		"account.",
		nodePool.AccountName, ".",
		"cluster.",
		nodePool.ClusterName, ".",
		"node-pool.", nodePool.Name,
	)
}

func (d *domain) domainResUpdateSubject(domainEntry *entities.DomainEntry) string {
	return fmt.Sprint(
		"res-updates.",
		"account.",
		domainEntry.AccountName, ".",
		"cluster.",
		domainEntry.ClusterName, ".",
		"domain.", domainEntry.DomainName,
	)
}


func (d *domain) vpnDeviceResUpdateSubject(device *entities.VPNDevice) string {
	return fmt.Sprint(
		"res-updates.",
		"account.",
		device.AccountName, ".",
		"cluster.",
		device.ClusterName, ".",
		"vpn-device.", device.Name,
	)
}


func (d *domain) pvcResUpdateSubject(pvc *entities.PersistentVolumeClaim) string {
	return fmt.Sprint(
		"res-updates.",
		"account.",
		pvc.AccountName, ".",
		"cluster.",
		pvc.ClusterName, ".",
		"vpn-device.", pvc.Name,
	)
}
