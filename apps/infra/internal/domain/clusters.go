package domain

import (
	"fmt"
	iamT "kloudlite.io/apps/iam/types"
	"time"

	"kloudlite.io/apps/infra/internal/entities"

	fn "kloudlite.io/pkg/functions"
	"kloudlite.io/pkg/repos"
	t "kloudlite.io/pkg/types"
)

func (d *domain) CreateCluster(ctx InfraContext, cluster entities.Cluster) (*entities.Cluster, error) {
	if err := d.canPerformActionInAccount(ctx, iamT.CreateCluster); err != nil {
		return nil, err
	}

	cluster.EnsureGVK()
	cluster.Namespace = d.getAccountNamespace(ctx.AccountName)

	if err := d.k8sExtendedClient.ValidateStruct(ctx, &cluster.Cluster); err != nil {
		return nil, err
	}

	cps, err := d.findProviderSecret(ctx, cluster.Spec.CredentialsRef.Name)
	if err != nil {
		return nil, err
	}

	if cps.IsMarkedForDeletion() {
		return nil, fmt.Errorf("cloud provider secret %q is marked for deletion, aborting cluster creation ...", cps.Name)
	}

	cluster.Spec.CredentialsRef.Namespace = cps.Namespace

	cluster.IncrementRecordVersion()
	cluster.AccountName = ctx.AccountName
	cluster.SyncStatus = t.GenSyncStatus(t.SyncActionApply, cluster.RecordVersion)

	nCluster, err := d.clusterRepo.Create(ctx, &cluster)
	if err != nil {
		if d.clusterRepo.ErrAlreadyExists(err) {
			return nil, fmt.Errorf("cluster with name %q already exists in namespace %q", cluster.Name, cluster.Namespace)
		}
		return nil, err
	}

	if err := d.ensureNamespaceForAccount(ctx, ctx.AccountName); err != nil {
		return nil, err
	}

	if err := d.applyK8sResource(ctx, &nCluster.Cluster, nCluster.RecordVersion); err != nil {
		return nil, err
	}

	return nCluster, nil
}

func (d *domain) ListClusters(ctx InfraContext, mf map[string]repos.MatchFilter, pagination repos.CursorPagination) (*repos.PaginatedRecord[*entities.Cluster], error) {
	if err := d.canPerformActionInAccount(ctx, iamT.ListClusters); err != nil {
		return nil, err
	}
	f := repos.Filter{
		"accountName":        ctx.AccountName,
		"metadata.namespace": d.getAccountNamespace(ctx.AccountName),
	}

	return d.clusterRepo.FindPaginated(ctx, d.secretRepo.MergeMatchFilters(f, mf), pagination)
}

func (d *domain) GetCluster(ctx InfraContext, name string) (*entities.Cluster, error) {
	if err := d.canPerformActionInAccount(ctx, iamT.GetCluster); err != nil {
		return nil, err
	}
	return d.clusterRepo.FindOne(ctx, repos.Filter{
		"accountName":        ctx.AccountName,
		"metadata.name":      name,
		"metadata.namespace": d.getAccountNamespace(ctx.AccountName),
	})
}

func (d *domain) UpdateCluster(ctx InfraContext, cluster entities.Cluster) (*entities.Cluster, error) {
	if err := d.canPerformActionInAccount(ctx, iamT.UpdateCluster); err != nil {
		return nil, err
	}
	cluster.EnsureGVK()
	clus, err := d.findCluster(ctx, cluster.Name)
	if err != nil {
		return nil, err
	}

	if clus.IsMarkedForDeletion() {
		return nil, fmt.Errorf("cluster %q in namespace %q is marked for deletion, could not perform any update operation", clus.Name, clus.Namespace)
	}

	cps, err := d.findProviderSecret(ctx, cluster.Spec.CredentialsRef.Name)
	if err != nil {
		return nil, err
	}

	if cps.IsMarkedForDeletion() {
		return nil, fmt.Errorf("cloud provider secret %q is marked for deletion, aborting cluster update ...", cps.Name)
	}

	cluster.Spec.CredentialsRef.Namespace = cps.Namespace

	clus.IncrementRecordVersion()

	clus.Spec = cluster.Cluster.Spec
	clus.Labels = cluster.Labels
	clus.Annotations = cluster.Annotations
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
	if err := d.canPerformActionInAccount(ctx, iamT.DeleteCluster); err != nil {
		return err
	}
	c, err := d.findCluster(ctx, name)
	if err != nil {
		return err
	}

	c.MarkedForDeletion = fn.New(true)
	c.SyncStatus = t.GetSyncStatusForDeletion(c.Generation)
	upC, err := d.clusterRepo.UpdateById(ctx, c.Id, c)
	if err != nil {
		return err
	}

	return d.deleteK8sResource(ctx, &upC.Cluster)
}

func (d *domain) OnDeleteClusterMessage(ctx InfraContext, cluster entities.Cluster) error {
	return d.clusterRepo.DeleteOne(ctx, repos.Filter{
		"accountName":        ctx.AccountName,
		"metadata.name":      cluster.Name,
		"metadata.namespace": d.getAccountNamespace(ctx.AccountName),
	})
}

func (d *domain) OnUpdateClusterMessage(ctx InfraContext, cluster entities.Cluster) error {
	c, err := d.findCluster(ctx, cluster.Name)
	if err != nil {
		return err
	}

	if err := d.matchRecordVersion(cluster.Annotations, c.RecordVersion); err != nil {
		return nil
	}

	c.Cluster.Labels = cluster.Labels
	c.Cluster.Annotations = cluster.Annotations
	c.Cluster.Spec = cluster.Spec

	c.SyncStatus.LastSyncedAt = time.Now()
	c.SyncStatus.Error = nil
	c.SyncStatus.RecordVersion = c.RecordVersion
	c.SyncStatus.State = t.SyncStateReceivedUpdateFromAgent

	c.Status = cluster.Status

	_, err = d.clusterRepo.UpdateById(ctx, c.Id, c)
	return err
}

func (d *domain) findCluster(ctx InfraContext, clusterName string) (*entities.Cluster, error) {
	cluster, err := d.clusterRepo.FindOne(ctx, repos.Filter{
		"accountName":        ctx.AccountName,
		"metadata.name":      clusterName,
		"metadata.namespace": d.getAccountNamespace(ctx.AccountName),
	})
	if err != nil {
		return nil, err
	}
	if cluster == nil {
		return nil, fmt.Errorf("cluster with name %q not found", clusterName)
	}
	return cluster, nil
}
