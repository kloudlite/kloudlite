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

func (d *domain) ListManagedResources(ctx ResourceContext, search map[string]repos.MatchFilter, pq repos.CursorPagination) (*repos.PaginatedRecord[*entities.ManagedResource], error) {
	if err := d.canReadResourcesInEnvironment(ctx); err != nil {
		return nil, errors.NewE(err)
	}

	filters := ctx.DBFilters()
	return d.mresRepo.FindPaginated(ctx, d.mresRepo.MergeMatchFilters(filters, search), pq)
}

func (d *domain) findMRes(ctx ResourceContext, name string) (*entities.ManagedResource, error) {
	filters := ctx.DBFilters()
	filters.Add("metadata.name", name)

	mres, err := d.mresRepo.FindOne(ctx, filters)
	if err != nil {
		return nil, errors.NewE(err)
	}
	if mres == nil {
		return nil, errors.Newf("no managed resource with name (%s) found", name)
	}
	return mres, nil
}

func (d *domain) GetManagedResource(ctx ResourceContext, name string) (*entities.ManagedResource, error) {
	if err := d.canReadResourcesInEnvironment(ctx); err != nil {
		return nil, errors.NewE(err)
	}

	return d.findMRes(ctx, name)
}

// GetManagedResourceOutputKVs implements Domain.
func (d *domain) GetManagedResourceOutputKVs(ctx ResourceContext, keyrefs []ManagedResourceKeyRef) ([]*ManagedResourceKeyValueRef, error) {
	filters := ctx.DBFilters()

	names := make([]any, 0, len(keyrefs))
	for i := range keyrefs {
		names = append(names, keyrefs[i].MresName)
	}

	filters = d.mresRepo.MergeMatchFilters(filters, map[string]repos.MatchFilter{
		"metadata.name": {
			MatchType: repos.MatchTypeArray,
			Array:     names,
		},
	})

	mresSecrets, err := d.mresRepo.Find(ctx, repos.Query{Filter: filters})
	if err != nil {
		return nil, errors.NewE(err)
	}

	results := make([]*ManagedResourceKeyValueRef, 0, len(mresSecrets))

	data := make(map[string]map[string]string)

	for i := range mresSecrets {
		m := make(map[string]string, len(mresSecrets[i].SyncedOutputSecretRef.Data))
		for k, v := range mresSecrets[i].SyncedOutputSecretRef.Data {
			m[k] = string(v)
		}

		for k, v := range mresSecrets[i].SyncedOutputSecretRef.StringData {
			m[k] = v
		}

		data[mresSecrets[i].Name] = m
	}

	for i := range keyrefs {
		results = append(results, &ManagedResourceKeyValueRef{
			MresName: keyrefs[i].MresName,
			Key:      keyrefs[i].Key,
			Value:    data[keyrefs[i].MresName][keyrefs[i].Key],
		})
	}

	return results, nil
}

// GetManagedResourceOutputKeys implements Domain.
func (d *domain) GetManagedResourceOutputKeys(ctx ResourceContext, name string) ([]string, error) {
	filters := ctx.DBFilters()
	filters.Add("metadata.name", name)
	mresSecrets, err := d.mresRepo.FindOne(ctx, filters)
	if err != nil {
		return nil, errors.NewE(err)
	}

	results := make([]string, 0, len(mresSecrets.SyncedOutputSecretRef.Data))

	for _, data := range mresSecrets.SyncedOutputSecretRef.Data {
		results = append(results, string(data))
	}

	return results, nil
}

// mutations

func (d *domain) CreateManagedResource(ctx ResourceContext, mres entities.ManagedResource) (*entities.ManagedResource, error) {
	if err := d.canMutateResourcesInEnvironment(ctx); err != nil {
		return nil, errors.NewE(err)
	}

	env, err := d.findEnvironment(ctx.ConsoleContext, ctx.ProjectName, ctx.EnvironmentName)
	if err != nil {
		return nil, errors.NewE(err)
	}

	mres.Namespace = env.Spec.TargetNamespace

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
	mres.ProjectName = ctx.ProjectName
	mres.EnvironmentName = ctx.EnvironmentName

	mres.SyncStatus = t.GenSyncStatus(t.SyncActionApply, mres.RecordVersion)

	if _, err := d.upsertResourceMapping(ctx, &mres); err != nil {
		return nil, errors.NewE(err)
	}

	m, err := d.mresRepo.Create(ctx, &mres)
	if err != nil {
		if d.mresRepo.ErrAlreadyExists(err) {
			// TODO: better insights into error, when it is being caused by duplicated indexes
			return nil, errors.NewE(err)
		}
		return nil, errors.NewE(err)
	}
	d.resourceEventPublisher.PublishMresEvent(&mres, PublishAdd)

	if err := d.applyK8sResource(nil, "", &m.ManagedResource, 0); err != nil {
		return m, errors.NewE(err)
	}

	return m, nil
}

func (d *domain) UpdateManagedResource(ctx ResourceContext, mres entities.ManagedResource) (*entities.ManagedResource, error) {
	if err := d.canReadResourcesInEnvironment(ctx); err != nil {
		return nil, errors.NewE(err)
	}

	mres.EnsureGVK()
	if err := d.k8sClient.ValidateObject(ctx, &mres.ManagedResource); err != nil {
		return nil, errors.NewE(err)
	}

	m, err := d.findMRes(ctx, mres.Name)
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

	if err := d.applyK8sResource(nil, "", &upMres.ManagedResource, upMres.RecordVersion); err != nil {
		return upMres, errors.NewE(err)
	}

	return upMres, nil
}

func (d *domain) DeleteManagedResource(ctx ResourceContext, name string) error {
	if err := d.canMutateResourcesInEnvironment(ctx); err != nil {
		return errors.NewE(err)
	}

	mres, err := d.findMRes(ctx, name)
	if err != nil {
		return errors.NewE(err)
	}

	mres.SyncStatus = t.GenSyncStatus(t.SyncActionDelete, mres.RecordVersion)
	if _, err := d.mresRepo.UpdateById(ctx, mres.Id, mres); err != nil {
		return errors.NewE(err)
	}
	d.resourceEventPublisher.PublishMresEvent(mres, PublishUpdate)

	return d.deleteK8sResource(ctx, mres.ProjectName, &mres.ManagedResource)
}

func (d *domain) OnManagedResourceDeleteMessage(ctx ResourceContext, mres entities.ManagedResource) error {
	xmres, err := d.findMRes(ctx, mres.Name)
	if err != nil {
		return errors.NewE(err)
	}

	if err := d.MatchRecordVersion(mres.Annotations, xmres.RecordVersion); err != nil {
		return d.resyncK8sResource(ctx, xmres.ProjectName, mres.SyncStatus.Action, &mres.ManagedResource, mres.RecordVersion)
	}

	err = d.mresRepo.DeleteById(ctx, xmres.Id)
	if err != nil {
		return errors.NewE(err)
	}
	d.resourceEventPublisher.PublishMresEvent(xmres, PublishDelete)
	return nil
}

func (d *domain) OnManagedResourceUpdateMessage(ctx ResourceContext, mres entities.ManagedResource, status types.ResourceStatus, opts UpdateAndDeleteOpts) error {
	xmres, err := d.findMRes(ctx, mres.Name)
	if err != nil {
		return errors.NewE(err)
	}

	if err := d.MatchRecordVersion(mres.Annotations, xmres.RecordVersion); err != nil {
		return d.resyncK8sResource(ctx, xmres.ProjectName, mres.SyncStatus.Action, &mres.ManagedResource, mres.RecordVersion)
	}

	xmres.CreationTimestamp = mres.CreationTimestamp
	xmres.Labels = mres.Labels
	xmres.Annotations = mres.Annotations
	xmres.Generation = mres.Generation

	xmres.Status = mres.Status

	xmres.SyncStatus.State = func() t.SyncState {
		if status == types.ResourceStatusDeleting {
			return t.SyncStateDeletingAtAgent
		}
		return t.SyncStateUpdatedAtAgent
	}()
	xmres.SyncStatus.RecordVersion = xmres.RecordVersion
	xmres.SyncStatus.Error = nil
	xmres.SyncStatus.LastSyncedAt = opts.MessageTimestamp

	_, err = d.mresRepo.UpdateById(ctx, xmres.Id, xmres)
	d.resourceEventPublisher.PublishMresEvent(xmres, PublishUpdate)
	return errors.NewE(err)
}

func (d *domain) OnManagedResourceApplyError(ctx ResourceContext, errMsg string, name string, opts UpdateAndDeleteOpts) error {
	m, err2 := d.findMRes(ctx, name)
	if err2 != nil {
		return err2
	}

	m.SyncStatus.State = t.SyncStateErroredAtAgent
	m.SyncStatus.LastSyncedAt = opts.MessageTimestamp
	m.SyncStatus.Error = &errMsg
	_, err := d.mresRepo.UpdateById(ctx, m.Id, m)
	return errors.NewE(err)
}

func (d *domain) ResyncManagedResource(ctx ResourceContext, name string) error {
	if err := d.canMutateResourcesInEnvironment(ctx); err != nil {
		return errors.NewE(err)
	}

	mres, err := d.findMRes(ctx, name)
	if err != nil {
		return errors.NewE(err)
	}
	return d.resyncK8sResource(ctx, mres.ProjectName, mres.SyncStatus.Action, &mres.ManagedResource, mres.RecordVersion)
}
