package domain

import (
	"fmt"
	"time"

	"kloudlite.io/apps/infra/internal/entities"
	fn "kloudlite.io/pkg/functions"
	"kloudlite.io/pkg/repos"
	t "kloudlite.io/pkg/types"
)

func (d *domain) CreateNodePool(ctx InfraContext, clusterName string, nodePool entities.NodePool) (*entities.NodePool, error) {
	nodePool.EnsureGVK()
	if err := d.k8sExtendedClient.ValidateStruct(ctx, &nodePool.NodePool); err != nil {
		return nil, err
	}

	nodePool.IncrementRecordVersion()
	nodePool.AccountName = ctx.AccountName
	nodePool.ClusterName = clusterName
	nodePool.SyncStatus = t.GenSyncStatus(t.SyncActionApply, nodePool.RecordVersion)

	np, err := d.nodePoolRepo.Create(ctx, &nodePool)
	if err != nil {
		if d.nodePoolRepo.ErrAlreadyExists(err) {
			return nil, fmt.Errorf("nodepool with name %q already exists", nodePool.Name)
		}
		return nil, err
	}

	if err := d.applyToTargetCluster(ctx, clusterName, &np.NodePool, np.RecordVersion); err != nil {
		return nil, err
	}

	return np, nil
}

func (d *domain) UpdateNodePool(ctx InfraContext, clusterName string, nodePool entities.NodePool) (*entities.NodePool, error) {
	nodePool.EnsureGVK()
	if err := d.k8sExtendedClient.ValidateStruct(ctx, &nodePool.NodePool); err != nil {
		return nil, err
	}

	np, err := d.findNodePool(ctx, clusterName, nodePool.Name)
	if err != nil {
		return nil, err
	}

	if np.IsMarkedForDeletion() {
		return nil, fmt.Errorf("nodepool %q (clusterName=%q) is marked for deletion, aborting update ...", nodePool.Name, clusterName)
	}

	np.Labels = nodePool.Labels
	np.Annotations = nodePool.Annotations
	np.Spec = nodePool.Spec

	np.IncrementRecordVersion()
	np.SyncStatus = t.GenSyncStatus(t.SyncActionApply, np.RecordVersion)

	unp, err := d.nodePoolRepo.UpdateById(ctx, np.Id, np)
	if err != nil {
		return nil, err
	}

	if err := d.applyToTargetCluster(ctx, clusterName, &unp.NodePool, unp.RecordVersion); err != nil {
		return nil, err
	}

	return unp, nil
}

func (d *domain) DeleteNodePool(ctx InfraContext, clusterName string, poolName string) error {
	np, err := d.findNodePool(ctx, clusterName, poolName)
	if err != nil {
		return err
	}

	if np.IsMarkedForDeletion() {
		return fmt.Errorf("nodepool %q (clusterName=%q) is already marked for deletion", poolName, clusterName)
	}

	np.MarkedForDeletion = fn.New(true)
	np.SyncStatus = t.GetSyncStatusForDeletion(np.Generation)
	upC, err := d.nodePoolRepo.UpdateById(ctx, np.Id, np)
	if err != nil {
		return err
	}
	return d.deleteFromTargetCluster(ctx, clusterName, &upC.NodePool)
}

func (d *domain) GetNodePool(ctx InfraContext, clusterName string, poolName string) (*entities.NodePool, error) {
	np, err := d.findNodePool(ctx, clusterName, poolName)
	if err != nil {
		return nil, err
	}
	return np, nil
}

func (d *domain) ListNodePools(ctx InfraContext, clusterName string, pagination t.CursorPagination) (*repos.PaginatedRecord[*entities.NodePool], error) {
	return d.nodePoolRepo.FindPaginated(ctx, repos.Filter{
		"accountName": ctx.AccountName,
		"clusterName": clusterName,
	}, pagination)
}

func (d *domain) findNodePool(ctx InfraContext, clusterName string, poolName string) (*entities.NodePool, error) {
	np, err := d.nodePoolRepo.FindOne(ctx, repos.Filter{
		"accountName":   ctx.AccountName,
		"clusterName":   clusterName,
		"metadata.name": poolName,
	})
	if err != nil {
		return nil, err
	}
	if np == nil {
		return nil, fmt.Errorf("nodepool with name %q not found", clusterName)
	}
	return np, nil
}

func (d *domain) ResyncNodePool(ctx InfraContext, clusterName string, poolName string) error {
	np, err := d.findNodePool(ctx, clusterName, poolName)
	if err != nil {
		return err
	}

	return d.resyncToTargetCluster(ctx, np.SyncStatus.Action, clusterName, &np.NodePool, np.RecordVersion)
}

// on message events

func (d *domain) OnDeleteNodePoolMessage(ctx InfraContext, clusterName string, nodePool entities.NodePool) error {
	np, _ := d.findNodePool(ctx, clusterName, nodePool.Name)
	if np == nil {
		// does not exist, (maybe already deleted)
		return nil
	}

	if err := d.matchRecordVersion(nodePool.Annotations, np.RecordVersion); err != nil {
		return d.resyncToTargetCluster(ctx, np.SyncStatus.Action, clusterName, &np.NodePool, np.RecordVersion)
	}

	return d.nodePoolRepo.DeleteById(ctx, np.Id)
}

func (d *domain) OnUpdateNodePoolMessage(ctx InfraContext, clusterName string, nodePool entities.NodePool) error {
	np, err := d.findNodePool(ctx, clusterName, nodePool.Name)
	if err != nil {
		return err
	}

	if err := d.matchRecordVersion(nodePool.Annotations, np.RecordVersion); err != nil {
		return d.resyncToTargetCluster(ctx, np.SyncStatus.Action, clusterName, &np.NodePool, np.RecordVersion)
	}

	np.Status = nodePool.Status

	np.SyncStatus.State = t.SyncStateReceivedUpdateFromAgent
	np.SyncStatus.LastSyncedAt = time.Now()
	np.SyncStatus.Error = nil
	np.SyncStatus.RecordVersion = np.RecordVersion

	if _, err := d.nodePoolRepo.UpdateById(ctx, np.Id, np); err != nil {
		return err
	}
	return nil
}
