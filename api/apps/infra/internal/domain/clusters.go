package domain

import (
	"fmt"
	"time"

	"kloudlite.io/apps/infra/internal/domain/entities"
	"kloudlite.io/pkg/repos"
	t "kloudlite.io/pkg/types"
)

func (d *domain) CreateCluster(ctx InfraContext, cluster entities.Cluster) (*entities.Cluster, error) {
	cluster.EnsureGVK()
	if err := d.k8sExtendedClient.ValidateStruct(ctx, &cluster.Cluster); err != nil {
		return nil, err
	}

	cluster.IncrementRecordVersion()
	cluster.AccountName = ctx.AccountName
	cluster.SyncStatus = t.GenSyncStatus(t.SyncActionApply, cluster.RecordVersion)

	nCluster, err := d.clusterRepo.Create(ctx, &cluster)
	if err != nil {
		if d.clusterRepo.ErrAlreadyExists(err) {
			return nil, fmt.Errorf("cluster with name %q already exists", cluster.Name)
		}
		return nil, err
	}

	if err := d.applyK8sResource(ctx, &nCluster.Cluster, nCluster.RecordVersion); err != nil {
		return nil, err
	}

	return nCluster, nil
}

func (d *domain) ListClusters(ctx InfraContext, pagination t.CursorPagination) (*repos.PaginatedRecord[*entities.Cluster], error) {
	return d.clusterRepo.FindPaginated(ctx, repos.Filter{
		"accountName": ctx.AccountName,
	}, pagination)
}

func (d *domain) GetCluster(ctx InfraContext, name string) (*entities.Cluster, error) {
	return d.clusterRepo.FindOne(ctx, repos.Filter{
		"accountName":   ctx.AccountName,
		"metadata.name": name,
	})
}

func (d *domain) UpdateCluster(ctx InfraContext, cluster entities.Cluster) (*entities.Cluster, error) {
	cluster.EnsureGVK()
	clus, err := d.findCluster(ctx, cluster.Name)
	if err != nil {
		return nil, err
	}

	clus.Cluster = cluster.Cluster
	clus.SyncStatus = t.GenSyncStatus(t.SyncActionApply, clus.RecordVersion)

	uCluster, err := d.clusterRepo.UpdateById(ctx, clus.Id, clus)
	if err != nil {
		return nil, err
	}

	if err := d.applyK8sResource(ctx, &uCluster.Cluster, uCluster.RecordVersion); err != nil {
		return nil, err
	}

	return uCluster, nil
}

func (d *domain) DeleteCluster(ctx InfraContext, name string) error {
	c, err := d.findCluster(ctx, name)
	if err != nil {
		return err
	}

	c.SyncStatus = t.GetSyncStatusForDeletion(c.Generation)
	upC, err := d.clusterRepo.UpdateById(ctx, c.Id, c)
	if err != nil {
		return err
	}
	return d.deleteK8sResource(ctx, &upC.Cluster)
}

func (d *domain) OnDeleteClusterMessage(ctx InfraContext, cluster entities.Cluster) error {
	return d.clusterRepo.DeleteOne(ctx, repos.Filter{
		"accountName":   ctx.AccountName,
		"metadata.name": cluster.Name,
	})
}

func (d *domain) OnUpdateClusterMessage(ctx InfraContext, cluster entities.Cluster) error {
	c, err := d.findCluster(ctx, cluster.Name)
	if err != nil {
		return err
	}

	c.Cluster = cluster.Cluster
	c.SyncStatus.LastSyncedAt = time.Now()
	c.SyncStatus.State = t.SyncStateReceivedUpdateFromAgent

	_, err = d.clusterRepo.UpdateById(ctx, c.Id, c)
	return err
}

func (d *domain) findCluster(ctx InfraContext, clusterName string) (*entities.Cluster, error) {
	cluster, err := d.clusterRepo.FindOne(ctx, repos.Filter{
		"accountName":   ctx.AccountName,
		"metadata.name": clusterName,
	})
	if err != nil {
		return nil, err
	}
	if cluster == nil {
		return nil, fmt.Errorf("cluster with name %q not found", clusterName)
	}
	return cluster, nil
}

// func (d *domain) OnUpdateBYOCClusterMessage(ctx InfraContext, cluster entities.BYOCCluster) error {
// 	d.findBYOCCluster(ctx, cluster.Name)
// }
