package domain

import (
	"time"

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

	if mres.Spec.ResourceTemplate.TypeMeta.GroupVersionKind().GroupKind().Empty() {
		return nil, errors.New(".spec.resourceTemplate.apiVersion, and .spec.resourceTemplate.kind must be set")
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

	if _, err := d.upsertEnvironmentResourceMapping(ctx, &mres); err != nil {
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

	if err := d.applyK8sResource(ctx, ctx.ProjectName, &m.ManagedResource, mres.RecordVersion); err != nil {
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

	xmres, err := d.findMRes(ctx, mres.Name)
	if err != nil {
		return nil, errors.NewE(err)
	}

	patch := repos.Document{
		"recordVersion": xmres.RecordVersion + 1,
		"displayName":   mres.DisplayName,
		"lastUpdatedBy": common.CreatedOrUpdatedBy{
			UserId:    ctx.UserId,
			UserName:  ctx.UserName,
			UserEmail: ctx.UserEmail,
		},

		"metadata.labels":      mres.Labels,
		"metadata.annotations": mres.Annotations,

		"spec": mres.Spec,

		"syncStatus.state":           t.SyncStateInQueue,
		"syncStatus.syncScheduledAt": time.Now(),
		"syncStatus.action":          t.SyncActionApply,
	}

	upMres, err := d.mresRepo.PatchById(ctx, xmres.Id, patch)
	if err != nil {
		return nil, errors.NewE(err)
	}

	d.resourceEventPublisher.PublishMresEvent(upMres, PublishUpdate)

	if err := d.applyK8sResource(ctx, ctx.ProjectName, &upMres.ManagedResource, upMres.RecordVersion); err != nil {
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

	patch := repos.Document{
		"markedForDeletion":          true,
		"syncStatus.syncScheduledAt": time.Now(),
		"syncStatus.action":          t.SyncActionDelete,
		"syncStatus.state":           t.SyncStateInQueue,
	}

	umres, err := d.mresRepo.PatchById(ctx, mres.Id, patch)
	if err != nil {
		return errors.NewE(err)
	}
	d.resourceEventPublisher.PublishMresEvent(umres, PublishUpdate)

	return d.deleteK8sResource(ctx, mres.ProjectName, &mres.ManagedResource)
}

func (d *domain) OnManagedResourceDeleteMessage(ctx ResourceContext, mres entities.ManagedResource) error {
	xmres, err := d.findMRes(ctx, mres.Name)
	if err != nil {
		return errors.NewE(err)
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

	patch := repos.Document{
		"metadata.creationTimestamp": mres.CreationTimestamp,
		"metadata.labels":            mres.Labels,
		"metadata.annotations":       mres.Annotations,
		"metadata.generation":        mres.Generation,

		"status": mres.Status,

		"syncStatus.state": func() t.SyncState {
			if status == types.ResourceStatusDeleting {
				return t.SyncStateDeletingAtAgent
			}
			return t.SyncStateUpdatedAtAgent
		}(),
		"syncStatus.recordVersion": xmres.RecordVersion,
		"syncStatus.lastSyncedAt":  opts.MessageTimestamp,
		"syncStatus.error":         nil,
	}

	umres, err := d.mresRepo.PatchById(ctx, xmres.Id, patch)
	if err != nil {
		return err
	}
	d.resourceEventPublisher.PublishMresEvent(umres, PublishUpdate)
	return errors.NewE(err)
}

func (d *domain) OnManagedResourceApplyError(ctx ResourceContext, errMsg string, name string, opts UpdateAndDeleteOpts) error {
	m, err2 := d.findMRes(ctx, name)
	if err2 != nil {
		return err2
	}

	patch := repos.Document{
		"syncStatus.state":        t.SyncStateErroredAtAgent,
		"syncStatus.lastSyncedAt": opts.MessageTimestamp,
		"syncStatus.error":        errMsg,
	}

	umres, err := d.mresRepo.PatchById(ctx, m.Id, patch)
	if err != nil {
		return err
	}
	d.resourceEventPublisher.PublishMresEvent(umres, PublishUpdate)
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
