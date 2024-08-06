package domain

import (
	"github.com/kloudlite/api/apps/console/internal/entities"
	fc "github.com/kloudlite/api/apps/console/internal/entities/field-constants"
	iamT "github.com/kloudlite/api/apps/iam/types"
	"github.com/kloudlite/api/common"
	"github.com/kloudlite/api/pkg/errors"
	"github.com/kloudlite/api/pkg/repos"
	t "github.com/kloudlite/api/pkg/types"
	common_types "github.com/kloudlite/operator/apis/common-types"
	"github.com/kloudlite/operator/operators/resource-watcher/types"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (d *domain) ImportManagedResource(ctx ManagedResourceContext, mresName string, importName string) (*entities.ImportedManagedResource, error) {
	mr, err := d.findMRes(ctx, mresName)
	if err != nil {
		return nil, err
	}

	if mr.SyncedOutputSecretRef == nil {
		return nil, errors.Newf("synced output secret not found")
	}

	return d.createAndApplyImportedManagedResource(
		ResourceContext{ConsoleContext: ctx.ConsoleContext, EnvironmentName: *ctx.EnvironmentName},
		CreateAndApplyImportedManagedResourceArgs{
			ImportedManagedResourceName: importName,
			ManagedResourceRefID:        mr.Id,
		})
}

func (d *domain) DeleteImportedManagedResource(ctx ResourceContext, importName string) error {
	if err := d.canMutateResourcesInEnvironment(ctx); err != nil {
		return errors.NewE(err)
	}

	impMres, err := d.findImportedMRes(ctx, importName)
	if err != nil {
		return errors.NewE(err)
	}

	if err := d.deleteSecret(ctx, impMres.SecretRef.Name); err != nil {
		return errors.NewE(err)
	}

	if _, err := d.importedMresRepo.PatchById(ctx, impMres.Id, repos.Document{fc.MarkedForDeletion: true}); err != nil {
		return errors.NewE(err)
	}

	d.resourceEventPublisher.PublishResourceEvent(ctx, entities.ResourceTypeManagedResource, impMres.ManagedResourceRef.Name, PublishDelete)

	return nil
}

func (d *domain) deleteImportedManagedResources(ctx ConsoleContext, mresNamespace string) error {
	if err := d.canPerformActionInAccount(ctx, iamT.DeleteManagedResource); err != nil {
		return errors.NewE(err)
	}

	records, err := d.importedMresRepo.Find(ctx, repos.Query{
		Filter: repos.Filter{
			fc.AccountName: ctx.AccountName,
			fc.ImportedManagedResourceManagedResourceRefNamespace: mresNamespace,
		},
	})
	if err != nil {
		return errors.NewE(err)
	}

	for i := range records {
		if err := d.deleteSecret(ResourceContext{ConsoleContext: ctx, EnvironmentName: records[i].EnvironmentName}, records[i].SecretRef.Name); err != nil {
			return errors.NewE(err)
		}

		if err := d.importedMresRepo.DeleteById(ctx, records[i].Id); err != nil {
			return errors.NewE(err)
		}
	}

	return nil
}

func (d *domain) findImportedMRes(ctx ResourceContext, importName string) (*entities.ImportedManagedResource, error) {
	imr, err := d.importedMresRepo.FindOne(ctx, repos.Filter{
		fc.AccountName:                 ctx.AccountName,
		fc.EnvironmentName:             ctx.EnvironmentName,
		fc.ImportedManagedResourceName: importName,
	})
	if err != nil {
		return nil, err
	}

	if imr == nil {
		return nil, errors.Newf("no imported managed resource found")
	}

	return imr, nil
}

type CreateAndApplyImportedManagedResourceArgs struct {
	ImportedManagedResourceName string
	ManagedResourceRefID        repos.ID
}

func (d *domain) createAndApplyImportedManagedResource(ctx ResourceContext, args CreateAndApplyImportedManagedResourceArgs) (*entities.ImportedManagedResource, error) {
	mr, err := d.mresRepo.FindById(ctx, args.ManagedResourceRefID)
	if err != nil {
		return nil, err
	}

	if mr.SyncedOutputSecretRef == nil {
		return nil, errors.Newf("synced output secret not found")
	}

	outputSecret := mr.SyncedOutputSecretRef

	envTargetNamespace := d.getEnvironmentTargetNamespace(ctx.EnvironmentName)

	outputSecret.ObjectMeta = metav1.ObjectMeta{
		Name:      args.ImportedManagedResourceName,
		Namespace: envTargetNamespace,
	}

	imr, err := d.importedMresRepo.Create(ctx, &entities.ImportedManagedResource{
		Name: args.ImportedManagedResourceName,
		ManagedResourceRef: entities.ManagedResourceRef{
			ID:        mr.Id,
			Name:      mr.Name,
			Namespace: mr.Namespace,
		},
		SecretRef: common_types.SecretRef{
			Name:      args.ImportedManagedResourceName,
			Namespace: envTargetNamespace,
		},
		ResourceMetadata: common.ResourceMetadata{
			DisplayName: args.ImportedManagedResourceName,
			CreatedBy: common.CreatedOrUpdatedBy{
				UserId:    ctx.UserId,
				UserName:  ctx.UserName,
				UserEmail: ctx.UserEmail,
			},
			LastUpdatedBy: common.CreatedOrUpdatedBy{
				UserId:    ctx.UserId,
				UserName:  ctx.UserName,
				UserEmail: ctx.UserEmail,
			},
		},
		AccountName:     ctx.AccountName,
		EnvironmentName: ctx.EnvironmentName,
		SyncStatus:      t.GenSyncStatus(t.SyncActionApply, mr.RecordVersion),
	})
	if err != nil {
		return nil, errors.NewE(err)
	}

	if _, err := d.createSecret(ctx, entities.Secret{
		Secret:          *outputSecret,
		AccountName:     ctx.AccountName,
		EnvironmentName: ctx.EnvironmentName,
		For: &entities.SecretCreatedFor{
			RefId:        imr.Id,
			ResourceType: entities.ResourceTypeImportedManagedResource,
			Name:         imr.Name,
			Namespace:    mr.Namespace,
		},
		IsReadOnly: true,
	}); err != nil {
		return nil, errors.NewE(err)
	}

	d.resourceEventPublisher.PublishEnvironmentResourceEvent(ctx.ConsoleContext, imr.EnvironmentName, entities.ResourceTypeImportedManagedResource, imr.Name, PublishUpdate)

	return imr, nil
}

func (d *domain) OnImportedManagedResourceDeleteMessage(ctx ConsoleContext, imrId repos.ID) error {
	return d.importedMresRepo.DeleteById(ctx, imrId)
}

func (d *domain) ListImportedManagedResources(ctx ConsoleContext, envName string, search map[string]repos.MatchFilter, pq repos.CursorPagination) (*repos.PaginatedRecord[*entities.ImportedManagedResource], error) {
	if err := d.canPerformActionInAccount(ctx, iamT.ListManagedResources); err != nil {
		return nil, errors.NewE(err)
	}

	filters := d.mresRepo.MergeMatchFilters(repos.Filter{
		fc.AccountName:     ctx.AccountName,
		fc.EnvironmentName: envName,
	}, search)

	pr, err := d.importedMresRepo.FindPaginated(ctx, filters, pq)
	if err != nil {
		return nil, errors.NewE(err)
	}
	return pr, nil
}

func (d *domain) OnImportedManagedResourceUpdateMessage(ctx ConsoleContext, imrID repos.ID, status types.ResourceStatus, opts UpdateAndDeleteOpts) error {
	imr, err := d.importedMresRepo.FindById(ctx, imrID)
	if err != nil {
		return errors.NewE(err)
	}

	if imr == nil {
		return errors.Newf("no imported managed resource found")
	}

	patch := repos.Document{
		fc.SyncStatusState: func() t.SyncState {
			if status == types.ResourceStatusDeleting {
				return t.SyncStateDeletingAtAgent
			}
			return t.SyncStateUpdatedAtAgent
		}(),
		fc.SyncStatusRecordVersion: imr.RecordVersion,
		fc.SyncStatusLastSyncedAt:  opts.MessageTimestamp,
		fc.SyncStatusError:         nil,
	}

	umres, err := d.importedMresRepo.PatchById(ctx, imr.Id, patch)
	if err != nil {
		return err
	}

	d.resourceEventPublisher.PublishEnvironmentResourceEvent(ctx, umres.EnvironmentName, entities.ResourceTypeManagedResource, umres.ManagedResourceRef.Name, PublishUpdate)
	return nil
}

func (d *domain) GetImportedManagedResourceOutputKeys(ctx ResourceContext, name string) ([]string, error) {
	imr, err := d.findImportedMRes(ctx, name)
	if err != nil {
		return nil, errors.NewE(err)
	}

	mr, err := d.mresRepo.FindById(ctx, imr.ManagedResourceRef.ID)
	if err != nil {
		return nil, errors.NewE(err)
	}

	if mr.SyncedOutputSecretRef == nil {
		return nil, errors.Newf("waiting for managed resource output to be ready")
	}

	results := make([]string, 0, len(mr.SyncedOutputSecretRef.Data))

	for k := range mr.SyncedOutputSecretRef.Data {
		results = append(results, k)
	}

	return results, nil
}

func (d *domain) GetImportedManagedResourceOutputKVs(ctx ResourceContext, keyrefs []ManagedResourceKeyRef) ([]*ManagedResourceKeyValueRef, error) {
	uniqMres := make(map[string]struct{})
	for i := range keyrefs {
		uniqMres[keyrefs[i].MresName] = struct{}{}
	}

	uniqMresNames := make([]any, 0, len(uniqMres))
	for k := range uniqMres {
		uniqMresNames = append(uniqMresNames, k)
	}

	filters := d.importedMresRepo.MergeMatchFilters(repos.Filter{
		fc.AccountName:     ctx.AccountName,
		fc.EnvironmentName: ctx.EnvironmentName,
	}, map[string]repos.MatchFilter{
		fc.ImportedManagedResourceName: {
			MatchType: repos.MatchTypeArray,
			Array:     uniqMresNames,
		},
	})

	importedResources, err := d.importedMresRepo.Find(ctx, repos.Query{Filter: filters})
	if err != nil {
		return nil, errors.NewE(err)
	}

	importNameMap := make(map[string]string, len(importedResources))
	for i := range importedResources {
		importNameMap[importedResources[i].Name] = importedResources[i].ManagedResourceRef.Name
	}

	mresIDs := make([]any, 0, len(importedResources))
	for i := range importedResources {
		mresIDs = append(mresIDs, importedResources[i].ManagedResourceRef.ID)
	}

	mresFilter := d.mresRepo.MergeMatchFilters(repos.Filter{}, map[string]repos.MatchFilter{
		fc.Id: {
			MatchType: repos.MatchTypeArray,
			Array:     mresIDs,
		},
	})

	mresSecrets, err := d.mresRepo.Find(ctx, repos.Query{Filter: mresFilter})
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
			Value:    data[importNameMap[keyrefs[i].MresName]][keyrefs[i].Key],
		})
	}

	return results, nil
}
