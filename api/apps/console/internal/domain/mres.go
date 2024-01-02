package domain

import (
	"github.com/kloudlite/api/apps/console/internal/entities"
	"github.com/kloudlite/api/common"
	"github.com/kloudlite/api/pkg/errors"
	"github.com/kloudlite/api/pkg/repos"
	t "github.com/kloudlite/api/pkg/types"
	"github.com/kloudlite/operator/operators/resource-watcher/types"
)

// query

func (d *domain) ListManagedResources(ctx ConsoleContext, namespace string, search map[string]repos.MatchFilter, pq repos.CursorPagination) (*repos.PaginatedRecord[*entities.ManagedResource], error) {
	if err := d.canReadResourcesInWorkspace(ctx, namespace); err != nil {
		return nil, errors.NewE(err)
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
		return nil, errors.NewE(err)
	}
	if mres == nil {
		return nil, errors.Newf(
			"no managed resource with name=%q,namespace=%q found",
			name,
			namespace,
		)
	}
	return mres, nil
}

func (d *domain) GetManagedResource(ctx ConsoleContext, namespace string, name string) (*entities.ManagedResource, error) {
	if err := d.canReadResourcesInWorkspace(ctx, namespace); err != nil {
		return nil, errors.NewE(err)
	}

	return d.findMRes(ctx, namespace, name)
}

// mutations

func (d *domain) CreateManagedResource(ctx ConsoleContext, mres entities.ManagedResource) (*entities.ManagedResource, error) {
	if err := d.canMutateResourcesInWorkspace(ctx, mres.Namespace); err != nil {
		return nil, errors.NewE(err)
	}

	mres.EnsureGVK()
	if err := d.k8sClient.ValidateObject(ctx, &mres.ManagedResource); err != nil {
		return nil, errors.NewE(err)
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
			return nil, errors.NewE(err)
		}
		return nil, errors.NewE(err)
	}
	d.resourceEventPublisher.PublishMresEvent(&mres, PublishAdd)

	if err := d.applyK8sResource(ctx, &m.ManagedResource, 0); err != nil {
		return m, errors.NewE(err)
	}

	return m, nil
}

func (d *domain) UpdateManagedResource(ctx ConsoleContext, mres entities.ManagedResource) (*entities.ManagedResource, error) {
	if err := d.canReadResourcesInWorkspace(ctx, mres.Namespace); err != nil {
		return nil, errors.NewE(err)
	}

	mres.EnsureGVK()
	if err := d.k8sClient.ValidateObject(ctx, &mres.ManagedResource); err != nil {
		return nil, errors.NewE(err)
	}

	m, err := d.findMRes(ctx, mres.Namespace, mres.Name)
	if err != nil {
		return nil, errors.NewE(err)
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
		return nil, errors.NewE(err)
	}

	d.resourceEventPublisher.PublishMresEvent(upMres, PublishUpdate)

	if err := d.applyK8sResource(ctx, &upMres.ManagedResource, upMres.RecordVersion); err != nil {
		return upMres, errors.NewE(err)
	}

	return upMres, nil
}

func (d *domain) DeleteManagedResource(ctx ConsoleContext, namespace string, name string) error {
	if err := d.canMutateResourcesInWorkspace(ctx, namespace); err != nil {
		return errors.NewE(err)
	}

	m, err := d.findMRes(ctx, namespace, name)
	if err != nil {
		return errors.NewE(err)
	}

	m.SyncStatus = t.GenSyncStatus(t.SyncActionDelete, m.RecordVersion)
	if _, err := d.mresRepo.UpdateById(ctx, m.Id, m); err != nil {
		return errors.NewE(err)
	}
	d.resourceEventPublisher.PublishMresEvent(m, PublishUpdate)

	return d.deleteK8sResource(ctx, &m.ManagedResource)
}

func (d *domain) OnManagedResourceDeleteMessage(ctx ConsoleContext, mres entities.ManagedResource) error {
	exMres, err := d.findMRes(ctx, mres.Namespace, mres.Name)
	if err != nil {
		return errors.NewE(err)
	}

	if err := d.MatchRecordVersion(mres.Annotations, exMres.RecordVersion); err != nil {
		return d.resyncK8sResource(ctx, mres.SyncStatus.Action, &mres.ManagedResource, mres.RecordVersion)
	}

	err = d.mresRepo.DeleteById(ctx, exMres.Id)
	if err != nil {
		return errors.NewE(err)
	}
	d.resourceEventPublisher.PublishMresEvent(exMres, PublishDelete)
	return nil
}

func (d *domain) OnManagedResourceUpdateMessage(ctx ConsoleContext, mres entities.ManagedResource, status types.ResourceStatus, opts UpdateAndDeleteOpts) error {
	exMres, err := d.findMRes(ctx, mres.Namespace, mres.Name)
	if err != nil {
		return errors.NewE(err)
	}

	if err := d.MatchRecordVersion(mres.Annotations, exMres.RecordVersion); err != nil {
		return d.resyncK8sResource(ctx, mres.SyncStatus.Action, &mres.ManagedResource, mres.RecordVersion)
	}

	exMres.CreationTimestamp = mres.CreationTimestamp
	exMres.Labels = mres.Labels
	exMres.Annotations = mres.Annotations
	exMres.Generation = mres.Generation

	exMres.Status = mres.Status

	exMres.SyncStatus.State = func() t.SyncState {
		if status == types.ResourceStatusDeleting {
			return t.SyncStateDeletingAtAgent
		}
		return t.SyncStateUpdatedAtAgent
	}()
	exMres.SyncStatus.RecordVersion = exMres.RecordVersion
	exMres.SyncStatus.Error = nil
	exMres.SyncStatus.LastSyncedAt = opts.MessageTimestamp

	_, err = d.mresRepo.UpdateById(ctx, exMres.Id, exMres)
	d.resourceEventPublisher.PublishMresEvent(exMres, PublishUpdate)
	return errors.NewE(err)
}

func (d *domain) OnManagedResourceApplyError(ctx ConsoleContext, errMsg string, namespace string, name string, opts UpdateAndDeleteOpts) error {
	m, err2 := d.findMRes(ctx, namespace, name)
	if err2 != nil {
		return err2
	}

	m.SyncStatus.State = t.SyncStateErroredAtAgent
	m.SyncStatus.LastSyncedAt = opts.MessageTimestamp
	m.SyncStatus.Error = &errMsg
	_, err := d.mresRepo.UpdateById(ctx, m.Id, m)
	return errors.NewE(err)
}

func (d *domain) ResyncManagedResource(ctx ConsoleContext, namespace, name string) error {
	if err := d.canMutateResourcesInWorkspace(ctx, namespace); err != nil {
		return errors.NewE(err)
	}

	m, err := d.findMRes(ctx, namespace, name)
	if err != nil {
		return errors.NewE(err)
	}
	return d.resyncK8sResource(ctx, m.SyncStatus.Action, &m.ManagedResource, m.RecordVersion)
}
