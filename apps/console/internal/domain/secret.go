package domain

import (
	"github.com/kloudlite/api/apps/console/internal/entities"
	fc "github.com/kloudlite/api/apps/console/internal/entities/field-constants"
	"github.com/kloudlite/api/common"
	"github.com/kloudlite/api/common/fields"
	"github.com/kloudlite/api/pkg/errors"
	fn "github.com/kloudlite/api/pkg/functions"
	"github.com/kloudlite/api/pkg/repos"
	t "github.com/kloudlite/api/pkg/types"
	"github.com/kloudlite/operator/operators/resource-watcher/types"
)

func (d *domain) ListSecrets(ctx ResourceContext, search map[string]repos.MatchFilter, pq repos.CursorPagination) (*repos.PaginatedRecord[*entities.Secret], error) {
	if err := d.canReadResourcesInEnvironment(ctx); err != nil {
		return nil, errors.NewE(err)
	}

	filters := ctx.DBFilters()
	return d.secretRepo.FindPaginated(ctx, d.secretRepo.MergeMatchFilters(filters, search), pq)
}

func (d *domain) findSecret(ctx ResourceContext, name string) (*entities.Secret, error) {
	xSecret, err := d.secretRepo.FindOne(
		ctx,
		ctx.DBFilters().Add(fields.MetadataName, name),
	)
	if err != nil {
		return nil, errors.NewE(err)
	}
	if xSecret == nil {
		return nil, errors.Newf("no secret with name (%s) found", name)
	}
	return xSecret, nil
}

func (d *domain) GetSecret(ctx ResourceContext, name string) (*entities.Secret, error) {
	if err := d.canReadResourcesInEnvironment(ctx); err != nil {
		return nil, errors.NewE(err)
	}
	return d.findSecret(ctx, name)
}

// GetSecretEntries implements Domain.
func (d *domain) GetSecretEntries(ctx ResourceContext, keyrefs []SecretKeyRef) ([]*SecretKeyValueRef, error) {
	filters := ctx.DBFilters()

	names := make([]any, 0, len(keyrefs))
	for i := range keyrefs {
		names = append(names, keyrefs[i].SecretName)
	}

	filters = d.secretRepo.MergeMatchFilters(filters, map[string]repos.MatchFilter{
		fields.MetadataName: {
			MatchType: repos.MatchTypeArray,
			Array:     names,
		},
	})

	secrets, err := d.secretRepo.Find(ctx, repos.Query{Filter: filters})
	if err != nil {
		return nil, errors.NewE(err)
	}

	results := make([]*SecretKeyValueRef, 0, len(secrets))

	data := make(map[string]map[string]string)

	for i := range secrets {
		m := make(map[string]string, len(secrets[i].Data))
		for k, v := range secrets[i].Data {
			m[k] = string(v)
		}

		for k, v := range secrets[i].StringData {
			m[k] = string(v)
		}

		data[secrets[i].Name] = m
	}

	for i := range keyrefs {
		results = append(results, &SecretKeyValueRef{
			SecretName: keyrefs[i].SecretName,
			Key:        keyrefs[i].Key,
			Value:      data[keyrefs[i].SecretName][keyrefs[i].Key],
		})
	}

	return results, nil
}

func (d *domain) CreateSecret(ctx ResourceContext, secret entities.Secret) (*entities.Secret, error) {
	if err := d.canMutateResourcesInEnvironment(ctx); err != nil {
		return nil, errors.NewE(err)
	}

	targetNamespace, err := d.envTargetNamespace(ctx.ConsoleContext, ctx.ProjectName, ctx.EnvironmentName)
	if err != nil {
		return nil, errors.NewE(err)
	}

	secret.SetGroupVersionKind(fn.GVK("v1", "Secret"))

	secret.Namespace = targetNamespace

	secret.IncrementRecordVersion()
	secret.CreatedBy = common.CreatedOrUpdatedBy{
		UserId:    ctx.UserId,
		UserName:  ctx.UserName,
		UserEmail: ctx.UserEmail,
	}
	secret.LastUpdatedBy = secret.CreatedBy

	secret.AccountName = ctx.AccountName
	secret.ProjectName = ctx.ProjectName
	secret.EnvironmentName = ctx.EnvironmentName
	secret.SyncStatus = t.GenSyncStatus(t.SyncActionApply, secret.RecordVersion)

	if _, err := d.upsertEnvironmentResourceMapping(ctx, &secret); err != nil {
		return nil, errors.NewE(err)
	}

	nsecret, err := d.secretRepo.Create(ctx, &secret)
	if err != nil {
		if d.secretRepo.ErrAlreadyExists(err) {
			// TODO: better insights into error, when it is being caused by duplicated indexes
			return nil, errors.NewE(err)
		}
		return nil, errors.NewE(err)
	}

	if nsecret.Annotations == nil {
		nsecret.Annotations = make(map[string]string)
	}

	for k, v := range types.SecretWatchingAnnotation {
		nsecret.Annotations[k] = v
	}

	if err := d.applyK8sResource(ctx, nsecret.ProjectName, &nsecret.Secret, nsecret.RecordVersion); err != nil {
		return nsecret, errors.NewE(err)
	}

	return nsecret, nil
}

func (d *domain) UpdateSecret(ctx ResourceContext, secret entities.Secret) (*entities.Secret, error) {
	if err := d.canMutateResourcesInEnvironment(ctx); err != nil {
		return nil, errors.NewE(err)
	}

	patchForUpdate := common.PatchForUpdate(
		ctx,
		&secret,
		common.PatchOpts{
			XPatch: repos.Document{
				fc.SecretData:       secret.Data,
				fc.SecretStringData: secret.StringData,
			},
		})

	upSecret, err := d.secretRepo.Patch(
		ctx,
		ctx.DBFilters().Add(fields.MetadataName, secret.Name),
		patchForUpdate,
	)

	if err != nil {
		return nil, errors.NewE(err)
	}

	if upSecret.Annotations == nil {
		upSecret.Annotations = make(map[string]string)
	}

	for k, v := range types.SecretWatchingAnnotation {
		upSecret.Annotations[k] = v
	}

	d.resourceEventPublisher.PublishResourceEvent(ctx, entities.ResourceTypeSecret, upSecret.Name, PublishUpdate)

	if err := d.applyK8sResource(ctx, upSecret.ProjectName, &upSecret.Secret, upSecret.RecordVersion); err != nil {
		return upSecret, errors.NewE(err)
	}

	return upSecret, nil
}

func (d *domain) DeleteSecret(ctx ResourceContext, name string) error {
	if err := d.canMutateResourcesInEnvironment(ctx); err != nil {
		return errors.NewE(err)
	}

	usecret, err := d.secretRepo.Patch(
		ctx,
		ctx.DBFilters().Add(fields.MetadataName, name),
		common.PatchForMarkDeletion(),
	)

	if err != nil {
		return errors.NewE(err)
	}

	d.resourceEventPublisher.PublishResourceEvent(ctx, entities.ResourceTypeSecret, usecret.Name, PublishUpdate)

	if err := d.deleteK8sResource(ctx, usecret.ProjectName, &usecret.Secret); err != nil {
		if errors.Is(err, ErrNoClusterAttached) {
			return d.secretRepo.DeleteById(ctx, usecret.Id)
		}
		return errors.NewE(err)
	}
	return nil
}

func (d *domain) OnSecretDeleteMessage(ctx ResourceContext, secret entities.Secret) error {
	err := d.secretRepo.DeleteOne(
		ctx,
		ctx.DBFilters().Add(fields.MetadataName, secret.Name),
	)

	if err != nil {
		return errors.NewE(err)
	}

	d.resourceEventPublisher.PublishResourceEvent(ctx, entities.ResourceTypeSecret, secret.Name, PublishDelete)

	return nil
}

func (d *domain) OnSecretUpdateMessage(ctx ResourceContext, secretIn entities.Secret, status types.ResourceStatus, opts UpdateAndDeleteOpts) error {
	xSecret, err := d.findSecret(ctx, secretIn.Name)

	if err != nil {
		return errors.NewE(err)
	}

	recordVersion, err := d.MatchRecordVersion(secretIn.Annotations, xSecret.RecordVersion)
	if err != nil {
		return d.resyncK8sResource(ctx, xSecret.ProjectName, xSecret.SyncStatus.Action, &xSecret.Secret, xSecret.RecordVersion)
	}

	usecret, err := d.secretRepo.PatchById(
		ctx,
		xSecret.Id,
		common.PatchForSyncFromAgent(&secretIn, recordVersion, status, common.PatchOpts{
			MessageTimestamp: opts.MessageTimestamp,
		}))

	d.resourceEventPublisher.PublishResourceEvent(ctx, usecret.GetResourceType(), usecret.GetName(), PublishUpdate)

	return errors.NewE(err)
}

func (d *domain) OnSecretApplyError(ctx ResourceContext, errMsg, name string, opts UpdateAndDeleteOpts) error {
	usecret, err := d.secretRepo.Patch(
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

	d.resourceEventPublisher.PublishResourceEvent(ctx, entities.ResourceTypeSecret, usecret.Name, PublishDelete)

	return errors.NewE(err)
}

func (d *domain) ResyncSecret(ctx ResourceContext, name string) error {
	if err := d.canMutateResourcesInEnvironment(ctx); err != nil {
		return errors.NewE(err)
	}

	secret, err := d.findSecret(ctx, name)
	if err != nil {
		return errors.NewE(err)
	}

	return d.resyncK8sResource(ctx, secret.ProjectName, secret.SyncStatus.Action, &secret.Secret, secret.RecordVersion)
}
