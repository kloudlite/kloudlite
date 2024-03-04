package domain

import (
	"fmt"

	"github.com/kloudlite/api/apps/console/internal/entities"
	fc "github.com/kloudlite/api/apps/console/internal/entities/field-constants"
	"github.com/kloudlite/api/common"
	"github.com/kloudlite/api/common/fields"
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
	mres, err := d.mresRepo.FindOne(
		ctx,
		ctx.DBFilters().Add(fields.MetadataName, name),
	)
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
		fields.MetadataName: {
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
	filters.Add(fields.MetadataName, name)

	mresSecret, err := d.findMRes(ctx, name)
	if err != nil {
		return nil, errors.NewE(err)
	}

	if mresSecret.SyncedOutputSecretRef == nil {
		return nil, errors.Newf("waiting for managed resource output to sync")
	}

	results := make([]string, 0, len(mresSecret.SyncedOutputSecretRef.Data))

	for k := range mresSecret.SyncedOutputSecretRef.Data {
		results = append(results, k)
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

	mres.Spec.ResourceName = genMresResourceName(ctx.EnvironmentName, mres.Name)

	return d.createAndApplyManagedResource(ctx, &mres)
}

func genMresResourceName(envName string, mresName string) string {
	return fmt.Sprintf("env-%s-%s", envName, mresName)
}

func (d *domain) createAndApplyManagedResource(ctx ResourceContext, mres *entities.ManagedResource) (*entities.ManagedResource, error) {
	mres.SyncStatus = t.GenSyncStatus(t.SyncActionApply, 0)

	if _, err := d.upsertEnvironmentResourceMapping(ctx, mres); err != nil {
		return nil, errors.NewE(err)
	}

	m, err := d.mresRepo.Create(ctx, mres)
	if err != nil {
		if d.mresRepo.ErrAlreadyExists(err) {
			// TODO: better insights into error, when it is being caused by duplicated indexes
			return nil, errors.NewE(err)
		}
		return nil, errors.NewE(err)
	}
	d.resourceEventPublisher.PublishResourceEvent(ctx, entities.ResourceTypeManagedResource, m.Name, PublishAdd)

	if err := d.applyK8sResource(ctx, ctx.ProjectName, &m.ManagedResource, m.RecordVersion); err != nil {
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

	patchForUpdate := common.PatchForUpdate(
		ctx,
		&mres,
		common.PatchOpts{
			XPatch: repos.Document{
				fc.ManagedResourceSpec: mres.Spec,
			},
		})

	upMres, err := d.mresRepo.Patch(
		ctx,
		ctx.DBFilters().Add(fields.MetadataName, mres.Name),
		patchForUpdate,
	)
	if err != nil {
		return nil, errors.NewE(err)
	}

	d.resourceEventPublisher.PublishResourceEvent(ctx, entities.ResourceTypeManagedResource, upMres.Name, PublishUpdate)

	if err := d.applyK8sResource(ctx, ctx.ProjectName, &upMres.ManagedResource, upMres.RecordVersion); err != nil {
		return upMres, errors.NewE(err)
	}

	return upMres, nil
}

func (d *domain) DeleteManagedResource(ctx ResourceContext, name string) error {
	if err := d.canMutateResourcesInEnvironment(ctx); err != nil {
		return errors.NewE(err)
	}

	umres, err := d.mresRepo.Patch(
		ctx,
		ctx.DBFilters().Add(fields.MetadataName, name),
		common.PatchForMarkDeletion(),
	)
	if err != nil {
		return errors.NewE(err)
	}
	d.resourceEventPublisher.PublishResourceEvent(ctx, entities.ResourceTypeManagedResource, umres.Name, PublishUpdate)
	if err := d.deleteK8sResource(ctx, umres.ProjectName, &umres.ManagedResource); err != nil {
		if errors.Is(err, ErrNoClusterAttached) {
			return d.mresRepo.DeleteById(ctx, umres.Id)
		}
		return errors.NewE(err)
	}
	return nil
}

func (d *domain) OnManagedResourceDeleteMessage(ctx ResourceContext, mres entities.ManagedResource) error {
	err := d.mresRepo.DeleteOne(
		ctx,
		ctx.DBFilters().Add(fields.MetadataName, mres.Name),
	)
	if err != nil {
		return errors.NewE(err)
	}
	d.resourceEventPublisher.PublishResourceEvent(ctx, entities.ResourceTypeManagedResource, mres.Name, PublishDelete)
	return nil
}

func (d *domain) OnManagedResourceUpdateMessage(ctx ResourceContext, mres entities.ManagedResource, status types.ResourceStatus, opts UpdateAndDeleteOpts) error {
	xmres, err := d.findMRes(ctx, mres.Name)
	if err != nil {
		return errors.NewE(err)
	}

	if xmres == nil {
		return errors.Newf("no manage resource found")
	}

	recordVersion, err := d.MatchRecordVersion(mres.Annotations, xmres.RecordVersion)
	if err != nil {
		return d.resyncK8sResource(ctx, xmres.ProjectName, mres.SyncStatus.Action, &mres.ManagedResource, mres.RecordVersion)
	}

	umres, err := d.mresRepo.PatchById(
		ctx,
		xmres.Id,
		common.PatchForSyncFromAgent(&mres, recordVersion, status, common.PatchOpts{
			MessageTimestamp: opts.MessageTimestamp,
			XPatch: repos.Document{
				fc.ManagedResourceSyncedOutputSecretRef: mres.SyncedOutputSecretRef,
			},
		}))
	if err != nil {
		return err
	}

	d.resourceEventPublisher.PublishResourceEvent(ctx, umres.GetResourceType(), umres.GetName(), PublishUpdate)

	if mres.SyncedOutputSecretRef != nil {
		if mres.SyncedOutputSecretRef.Labels == nil {
			mres.SyncedOutputSecretRef.Labels = map[string]string{}
		}
		mres.SyncedOutputSecretRef.Labels["kloudlite.io/secret.synced-by"] = fmt.Sprintf("%s/%s", umres.GetNamespace(), umres.GetName())

		secretData := make(map[string]string, len(mres.SyncedOutputSecretRef.Data))

		for k, v := range mres.SyncedOutputSecretRef.Data {
			secretData[k] = string(v)
		}

		mres.SyncedOutputSecretRef.Data = nil
		mres.SyncedOutputSecretRef.StringData = secretData

		if _, err = d.secretRepo.Upsert(ctx, repos.Filter{
			fc.AccountName:       ctx.AccountName,
			fc.ProjectName:       ctx.ProjectName,
			fc.EnvironmentName:   ctx.EnvironmentName,
			fc.MetadataName:      mres.SyncedOutputSecretRef.GetName(),
			fc.MetadataNamespace: mres.SyncedOutputSecretRef.GetNamespace(),
		}, &entities.Secret{
			Secret:          *mres.SyncedOutputSecretRef,
			AccountName:     ctx.AccountName,
			ProjectName:     ctx.ProjectName,
			EnvironmentName: ctx.EnvironmentName,
			ResourceMetadata: common.ResourceMetadata{
				DisplayName:   umres.GetName(),
				CreatedBy:     common.CreatedOrUpdatedByResourceSync,
				LastUpdatedBy: common.CreatedOrUpdatedByResourceSync,
			},
			SyncStatus: t.SyncStatus{
				LastSyncedAt:  opts.MessageTimestamp,
				Action:        t.SyncActionApply,
				RecordVersion: recordVersion,
				State:         t.SyncStateUpdatedAtAgent,
				Error:         nil,
			},
			IsReadOnly: true,
		}); err != nil {
			return errors.NewE(err)
		}
	}

	return nil
}

func (d *domain) OnManagedResourceApplyError(ctx ResourceContext, errMsg string, name string, opts UpdateAndDeleteOpts) error {
	umres, err := d.mresRepo.Patch(
		ctx,
		ctx.DBFilters().Add(fields.MetadataName, name),
		common.PatchForErrorFromAgent(
			errMsg,
			common.PatchOpts{
				MessageTimestamp: opts.MessageTimestamp,
			},
		),
	)
	if err != nil {
		return errors.NewE(err)
	}
	d.resourceEventPublisher.PublishResourceEvent(ctx, entities.ResourceTypeManagedResource, umres.Name, PublishDelete)
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
