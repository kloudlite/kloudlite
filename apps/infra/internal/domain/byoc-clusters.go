package domain

import (
	"fmt"
	iamT "kloudlite.io/apps/iam/types"
	"kloudlite.io/apps/infra/internal/entities"
	fn "kloudlite.io/pkg/functions"

	redpandaMsvcv1 "github.com/kloudlite/operator/apis/redpanda.msvc/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"kloudlite.io/common"
	"kloudlite.io/pkg/repos"
	t "kloudlite.io/pkg/types"
)

func (d *domain) findBYOCCluster(ctx InfraContext, clusterName string) (*entities.BYOCCluster, error) {
	cluster, err := d.byocClusterRepo.FindOne(ctx, repos.Filter{
		"spec.accountName":   ctx.AccountName,
		"metadata.name":      clusterName,
		"metadata.namespace": d.getAccountNamespace(ctx.AccountName),
	})
	if err != nil {
		return nil, err
	}
	if cluster == nil {
		return nil, fmt.Errorf("BYOC cluster with name %q not found", clusterName)
	}
	return cluster, nil
}

func (d *domain) CreateBYOCCluster(ctx InfraContext, cluster entities.BYOCCluster) (*entities.BYOCCluster, error) {
	if err := d.canPerformActionInAccount(ctx, iamT.CreateCluster); err != nil {
		return nil, err
	}

	cluster.EnsureGVK()
	cluster.IncomingKafkaTopicName = common.GetKafkaTopicName(ctx.AccountName, cluster.Name)

	if err := d.k8sExtendedClient.ValidateStruct(ctx, &cluster.BYOC); err != nil {
		return nil, err
	}

	cluster.IncrementRecordVersion()
	cluster.IsConnected = false
	cluster.Spec.AccountName = ctx.AccountName
	cluster.SyncStatus = t.GenSyncStatus(t.SyncActionApply, cluster.RecordVersion)

	nCluster, err := d.byocClusterRepo.Create(ctx, &cluster)
	if err != nil {
		if d.clusterRepo.ErrAlreadyExists(err) {
			return nil, err
		}
	}

	if err := d.ensureNamespaceForAccount(ctx, ctx.AccountName); err != nil {
		return nil, err
	}

	if err := d.applyK8sResource(ctx, &nCluster.BYOC, nCluster.RecordVersion); err != nil {
		return nil, err
	}

	redpandaTopic := redpandaMsvcv1.Topic{
		TypeMeta:   metav1.TypeMeta{},
		ObjectMeta: metav1.ObjectMeta{Name: cluster.IncomingKafkaTopicName, Namespace: d.env.ProviderSecretNamespace},
	}

	redpandaTopic.EnsureGVK()

	if err := d.applyK8sResource(ctx, &redpandaTopic, nCluster.RecordVersion); err != nil {
		return nil, err
	}

	return nCluster, nil
}

func (d *domain) ListBYOCClusters(ctx InfraContext, search *repos.SearchFilter, pagination t.CursorPagination) (*repos.PaginatedRecord[*entities.BYOCCluster], error) {
	if err := d.canPerformActionInAccount(ctx, iamT.ListClusters); err != nil {
		return nil, err
	}
	filter := repos.Filter{
		"accountName":        ctx.AccountName,
		"metadata.namespace": d.getAccountNamespace(ctx.AccountName),
	}
	return d.byocClusterRepo.FindPaginated(ctx, d.byocClusterRepo.MergeSearchFilter(filter, search), pagination)
}

func (d *domain) GetBYOCCluster(ctx InfraContext, name string) (*entities.BYOCCluster, error) {
	if err := d.canPerformActionInAccount(ctx, iamT.GetCluster); err != nil {
		return nil, err
	}
	return d.findBYOCCluster(ctx, name)
}

func (d *domain) UpdateBYOCCluster(ctx InfraContext, cluster entities.BYOCCluster) (*entities.BYOCCluster, error) {
	if err := d.canPerformActionInAccount(ctx, iamT.UpdateCluster); err != nil {
		return nil, err
	}

	cluster.EnsureGVK()
	if err := d.k8sExtendedClient.ValidateStruct(ctx, &cluster.BYOC); err != nil {
		return nil, err
	}

	c, err := d.findBYOCCluster(ctx, cluster.Name)
	if err != nil {
		return nil, err
	}

	c.IncrementRecordVersion()
	c.BYOC = cluster.BYOC
	c.SyncStatus = t.GenSyncStatus(t.SyncActionApply, c.RecordVersion)

	uCluster, err := d.byocClusterRepo.UpdateById(ctx, c.Id, c)
	if err != nil {
		return nil, err
	}

	if err := d.applyK8sResource(ctx, &uCluster.BYOC, uCluster.RecordVersion); err != nil {
		return nil, err
	}

	return uCluster, nil
}

func (d *domain) DeleteBYOCCluster(ctx InfraContext, name string) error {
	if err := d.canPerformActionInAccount(ctx, iamT.DeleteCluster); err != nil {
		return err
	}

	clus, err := d.findBYOCCluster(ctx, name)
	if err != nil {
		return err
	}

	if clus.IsMarkedForDeletion() {
		return fmt.Errorf("BYOC cluster %q is already marked for deletion", name)
	}

	clus.MarkedForDeletion = fn.New(true)
	clus.SyncStatus = t.GetSyncStatusForDeletion(clus.Generation)
	upC, err := d.byocClusterRepo.UpdateById(ctx, clus.Id, clus)
	if err != nil {
		return err
	}
	return d.deleteK8sResource(ctx, &upC.BYOC)
}

func (d *domain) ResyncBYOCCluster(ctx InfraContext, name string) error {
	clus, err := d.findBYOCCluster(ctx, name)
	if err != nil {
		return err
	}

	if err := d.applyK8sResource(ctx, &clus.BYOC, clus.RecordVersion); err != nil {
		return err
	}

	redpandaTopic := redpandaMsvcv1.Topic{
		TypeMeta: metav1.TypeMeta{},
		ObjectMeta: metav1.ObjectMeta{
			Name:      clus.IncomingKafkaTopicName,
			Namespace: d.env.ProviderSecretNamespace,
		},
	}

	redpandaTopic.EnsureGVK()
	return d.applyK8sResource(ctx, &redpandaTopic, clus.RecordVersion)
}

func (d *domain) OnDeleteBYOCClusterMessage(ctx InfraContext, cluster entities.BYOCCluster) error {
	return d.clusterRepo.DeleteOne(ctx, repos.Filter{
		"accountName":        ctx.AccountName,
		"metadata.name":      cluster.Name,
		"metadata.namespace": d.getAccountNamespace(ctx.AccountName),
	})
}

func (d *domain) OnUpdateBYOCClusterMessage(ctx InfraContext, cluster entities.BYOCCluster) error {
	c, err := d.findBYOCCluster(ctx, cluster.Name)
	if err != nil {
		return err
	}

	c.SyncStatus.State = t.SyncStateReceivedUpdateFromAgent

	_, err = d.byocClusterRepo.UpdateById(ctx, c.Id, &cluster)
	if err != nil {
		return err
	}
	return nil
}
