package domain

import (
	"fmt"
	"time"

	"kloudlite.io/apps/console/internal/entities"
	"kloudlite.io/common"
	"kloudlite.io/pkg/repos"
	t "kloudlite.io/pkg/types"
)

// query

func (d *domain) ListManagedResources(ctx ConsoleContext, namespace string, search map[string]repos.MatchFilter, pq repos.CursorPagination) (*repos.PaginatedRecord[*entities.ManagedResource], error) {
	if err := d.canReadResourcesInWorkspace(ctx, namespace); err != nil {
		return nil, err
	}

	filter := repos.Filter{
		"accountName":        ctx.AccountName,
		"clusterName":        ctx.ClusterName,
		"metadata.namespace": namespace,
	}

	return d.mresRepo.FindPaginated(ctx, d.mresRepo.MergeMatchFilters(filter, search), pq)
}

func (d *domain) findMRes(ctx ConsoleContext, namespace string, name string) (*entities.ManagedResource, error) {
	mres, err := d.mresRepo.FindOne(ctx, repos.Filter{
		"accountName":        ctx.AccountName,
		"clusterName":        ctx.ClusterName,
		"metadata.namespace": namespace,
		"metadata.name":      name,
	})
	if err != nil {
		return nil, err
	}
	if mres == nil {
		return nil, fmt.Errorf(
			"no managed resource with name=%q,namespace=%q found",
			name,
			namespace,
		)
	}
	return mres, nil
}

func (d *domain) GetManagedResource(ctx ConsoleContext, namespace string, name string) (*entities.ManagedResource, error) {
	if err := d.canReadResourcesInWorkspace(ctx, namespace); err != nil {
		return nil, err
	}

	return d.findMRes(ctx, namespace, name)
}

// mutations

func (d *domain) CreateManagedResource(ctx ConsoleContext, mres entities.ManagedResource) (*entities.ManagedResource, error) {
	if err := d.canMutateResourcesInWorkspace(ctx, mres.Namespace); err != nil {
		return nil, err
	}

	mres.EnsureGVK()
	if err := d.k8sClient.ValidateObject(ctx, &mres.ManagedResource); err != nil {
		return nil, err
	}

	mres.IncrementRecordVersion()

	mres.CreatedBy = common.CreatedOrUpdatedBy{
		UserId:    ctx.UserId,
		UserName:  ctx.UserName,
		UserEmail: ctx.UserEmail,
	}
	mres.LastUpdatedBy = mres.CreatedBy

	mres.AccountName = ctx.AccountName
	mres.ClusterName = ctx.ClusterName
	mres.SyncStatus = t.GenSyncStatus(t.SyncActionApply, mres.RecordVersion)

	m, err := d.mresRepo.Create(ctx, &mres)
	if err != nil {
		if d.mresRepo.ErrAlreadyExists(err) {
			// TODO: better insights into error, when it is being caused by duplicated indexes
			return nil, err
		}
		return nil, err
	}

	if err := d.applyK8sResource(ctx, &m.ManagedResource, 0); err != nil {
		return m, err
	}

	return m, nil
}

func (d *domain) UpdateManagedResource(ctx ConsoleContext, mres entities.ManagedResource) (*entities.ManagedResource, error) {
	if err := d.canReadResourcesInWorkspace(ctx, mres.Namespace); err != nil {
		return nil, err
	}

	mres.EnsureGVK()
	if err := d.k8sClient.ValidateObject(ctx, &mres.ManagedResource); err != nil {
		return nil, err
	}

	m, err := d.findMRes(ctx, mres.Namespace, mres.Name)
	if err != nil {
		return nil, err
	}

	m.IncrementRecordVersion()
	m.LastUpdatedBy = common.CreatedOrUpdatedBy{
		UserId:    ctx.UserId,
		UserName:  ctx.UserName,
		UserEmail: ctx.UserEmail,
	}
	m.DisplayName = mres.DisplayName

	m.Labels = mres.Labels
	m.Annotations = mres.Annotations

	m.Spec = mres.Spec
	m.SyncStatus = t.GenSyncStatus(t.SyncActionApply, m.RecordVersion)

	upMres, err := d.mresRepo.UpdateById(ctx, m.Id, m)
	if err != nil {
		return nil, err
	}

	if err := d.applyK8sResource(ctx, &upMres.ManagedResource, upMres.RecordVersion); err != nil {
		return upMres, err
	}

	return upMres, nil
}

func (d *domain) DeleteManagedResource(ctx ConsoleContext, namespace string, name string) error {
	if err := d.canMutateResourcesInWorkspace(ctx, namespace); err != nil {
		return err
	}

	m, err := d.findMRes(ctx, namespace, name)
	if err != nil {
		return err
	}

	m.SyncStatus = t.GenSyncStatus(t.SyncActionDelete, m.RecordVersion)
	if _, err := d.mresRepo.UpdateById(ctx, m.Id, m); err != nil {
		return err
	}

	return d.deleteK8sResource(ctx, &m.ManagedResource)
}

func (d *domain) OnDeleteManagedResourceMessage(ctx ConsoleContext, mres entities.ManagedResource) error {
	exMres, err := d.findMRes(ctx, mres.Namespace, mres.Name)
	if err != nil {
		return err
	}

	if err := d.MatchRecordVersion(mres.Annotations, exMres.RecordVersion); err != nil {
		return d.resyncK8sResource(ctx, mres.SyncStatus.Action, &mres.ManagedResource, mres.RecordVersion)
	}

	return d.mresRepo.DeleteById(ctx, exMres.Id)
}

func (d *domain) OnUpdateManagedResourceMessage(ctx ConsoleContext, mres entities.ManagedResource) error {
	exMres, err := d.findMRes(ctx, mres.Namespace, mres.Name)
	if err != nil {
		return err
	}

	annotatedVersion, err := d.parseRecordVersionFromAnnotations(mres.Annotations)
	if err != nil {
		return d.resyncK8sResource(ctx, mres.SyncStatus.Action, &mres.ManagedResource, mres.RecordVersion)
	}

	if annotatedVersion != exMres.RecordVersion {
		return d.resyncK8sResource(ctx, mres.SyncStatus.Action, &mres.ManagedResource, mres.RecordVersion)
	}

	exMres.CreationTimestamp = mres.CreationTimestamp
	exMres.Labels = mres.Labels
	exMres.Annotations = mres.Annotations
	exMres.Generation = mres.Generation

	exMres.Status = mres.Status

	exMres.SyncStatus.State = t.SyncStateReceivedUpdateFromAgent
	exMres.SyncStatus.RecordVersion = annotatedVersion
	exMres.SyncStatus.Error = nil
	exMres.SyncStatus.LastSyncedAt = time.Now()

	_, err = d.mresRepo.UpdateById(ctx, exMres.Id, exMres)
	return err
}

func (d *domain) OnApplyManagedResourceError(ctx ConsoleContext, errMsg string, namespace string, name string) error {
	m, err2 := d.findMRes(ctx, namespace, name)
	if err2 != nil {
		return err2
	}

	m.SyncStatus.State = t.SyncStateErroredAtAgent
	m.SyncStatus.LastSyncedAt = time.Now()
	m.SyncStatus.Error = &errMsg
	_, err := d.mresRepo.UpdateById(ctx, m.Id, m)
	return err
}

func (d *domain) ResyncManagedResource(ctx ConsoleContext, namespace, name string) error {
	if err := d.canMutateResourcesInWorkspace(ctx, namespace); err != nil {
		return err
	}

	m, err := d.findMRes(ctx, namespace, name)
	if err != nil {
		return err
	}
	return d.resyncK8sResource(ctx, m.SyncStatus.Action, &m.ManagedResource, m.RecordVersion)
}
