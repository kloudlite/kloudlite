package domain

import (
	"fmt"
	"strings"

	"github.com/kloudlite/api/apps/console/internal/entities"
	fc "github.com/kloudlite/api/apps/console/internal/entities/field-constants"
	"github.com/kloudlite/api/common"
	"github.com/kloudlite/api/common/fields"
	"github.com/kloudlite/api/pkg/errors"
	fn "github.com/kloudlite/api/pkg/functions"
	"github.com/kloudlite/api/pkg/repos"
	t "github.com/kloudlite/api/pkg/types"
	"github.com/kloudlite/operator/operators/resource-watcher/types"
	corev1 "k8s.io/api/core/v1"
)

func (d *domain) ListSecrets(ctx ResourceContext, search map[string]repos.MatchFilter, pq repos.CursorPagination) (*repos.PaginatedRecord[*entities.Secret], error) {
	if err := d.canReadResourcesInEnvironment(ctx); err != nil {
		return nil, errors.NewE(err)
	}

	filters := ctx.DBFilters()
	pr, err := d.secretRepo.FindPaginated(ctx, d.secretRepo.MergeMatchFilters(filters, search), pq)
	if err != nil {
		return nil, errors.NewE(err)
	}

	for i := range pr.Edges {
		fromDataToStringData(&pr.Edges[i].Node.Secret)
		pr.Edges[i].Node.StringData = filterOutHiddenKeysFromSecret(pr.Edges[i].Node)
	}

	return pr, nil
}

func filterOutHiddenKeysFromSecret(secret *entities.Secret) map[string]string {
	if secret.For != nil {
		// means, this is a secret created by something other than a secret
		fdata := make(map[string]string, len(secret.StringData))
		for k, v := range secret.StringData {
			if !strings.HasPrefix(k, ".") {
				fdata[k] = v
			}
		}
		return fdata
	}
	return secret.StringData
}

func fromDataToStringData(secret *corev1.Secret) {
	if secret.StringData == nil {
		secret.StringData = make(map[string]string, len(secret.Data))
	}

	for k, v := range secret.Data {
		secret.StringData[k] = string(v)
	}
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

	fromDataToStringData(&xSecret.Secret)
	xSecret.StringData = filterOutHiddenKeysFromSecret(xSecret)
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
		fromDataToStringData(&secrets[i].Secret)
		if secrets[i].For != nil {
			secrets[i].StringData = filterOutHiddenKeysFromSecret(secrets[i])
		}
		m := make(map[string]string, len(secrets[i].StringData))
		for k, v := range secrets[i].StringData {
			m[k] = v
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
	return d.createSecret(ctx, secret)
}

func (d *domain) createSecret(ctx ResourceContext, secret entities.Secret) (*entities.Secret, error) {
	if err := d.canMutateResourcesInEnvironment(ctx); err != nil {
		return nil, errors.NewE(err)
	}

	targetNamespace, err := d.envTargetNamespace(ctx.ConsoleContext, ctx.EnvironmentName)
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
	secret.EnvironmentName = ctx.EnvironmentName
	secret.SyncStatus = t.GenSyncStatus(t.SyncActionApply, secret.RecordVersion)

	if secret.Annotations == nil {
		secret.Annotations = make(map[string]string, len(types.SecretWatchingAnnotation))
	}

	for k, v := range types.SecretWatchingAnnotation {
		secret.Annotations[k] = v
	}

	return d.createAndApplySecret(ctx, &secret)
}

func (d *domain) createAndApplySecret(ctx ResourceContext, secret *entities.Secret) (*entities.Secret, error) {
	if _, err := d.upsertEnvironmentResourceMapping(ctx, secret); err != nil {
		return nil, errors.NewE(err)
	}
	nsecret, err := d.secretRepo.Create(ctx, secret)
	if err != nil {
		if d.secretRepo.ErrAlreadyExists(err) {
			// TODO: better insights into error, when it is being caused by duplicated indexes
			return nil, errors.NewE(err)
		}
		return nil, errors.NewE(err)
	}

	if err := d.applyK8sResource(ctx, nsecret.EnvironmentName, &nsecret.Secret, nsecret.RecordVersion); err != nil {
		return nsecret, errors.NewE(err)
	}

	return nsecret, nil
}

func (d *domain) UpdateSecret(ctx ResourceContext, secret entities.Secret) (*entities.Secret, error) {
	if err := d.canMutateResourcesInEnvironment(ctx); err != nil {
		return nil, errors.NewE(err)
	}

	if secret.Annotations == nil {
		secret.Annotations = make(map[string]string, len(types.SecretWatchingAnnotation))
	}

	for k, v := range types.SecretWatchingAnnotation {
		secret.Annotations[k] = v
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

	if err := d.applyK8sResource(ctx, upSecret.EnvironmentName, &upSecret.Secret, upSecret.RecordVersion); err != nil {
		return upSecret, errors.NewE(err)
	}

	return upSecret, nil
}

func (d *domain) DeleteSecret(ctx ResourceContext, name string) error {
	return d.deleteSecret(ctx, name)
}

func (d *domain) deleteSecret(ctx ResourceContext, name string) error {
	if err := d.canMutateResourcesInEnvironment(ctx); err != nil {
		return errors.NewE(err)
	}

	defer func() {
		d.resourceEventPublisher.PublishResourceEvent(ctx, entities.ResourceTypeSecret, name, PublishUpdate)
	}()

	filters := ctx.DBFilters().Add(fields.MetadataName, name)

	usecret, err := d.secretRepo.Patch(ctx, filters, common.PatchForMarkDeletion())
	if err != nil {
		if errors.Is(err, repos.ErrNoDocuments) {
			return nil
		}
		return errors.NewE(err)
	}

	if err := d.deleteK8sResource(ctx, usecret.EnvironmentName, &usecret.Secret); err != nil {
		if errors.Is(err, ErrNoClusterAttached) {
			return d.secretRepo.DeleteById(ctx, usecret.Id)
		}
		return errors.NewE(err)
	}
	return nil
}

func (d *domain) OnSecretDeleteMessage(ctx ResourceContext, secret entities.Secret) error {
	if v, ok := secret.GetLabels()["app.kubernetes.io/managed-by"]; ok && v == "Helm" {
		err := d.secretRepo.DeleteOne(ctx, ctx.DBFilters().Add(fields.MetadataName, secret.Name).Add(fields.MetadataNamespace, secret.Namespace))
		if err != nil {
			return errors.NewE(err)
		}
		d.resourceEventPublisher.PublishResourceEvent(ctx, entities.ResourceTypeConfig, secret.Name, PublishDelete)
	}

	s, err := d.findSecret(ctx, secret.Name)
	if err != nil {
		return errors.NewE(err)
	}

	// if _, err := d.MatchRecordVersion(secret.Annotations, s.RecordVersion); err != nil {
	// 	return d.resyncK8sResource(ctx, s.EnvironmentName, s.SyncStatus.Action, &s.Secret, s.RecordVersion)
	// }

	if s.For != nil {
		switch s.For.ResourceType {
		case entities.ResourceTypeImportedManagedResource:
			{
				if err := d.OnImportedManagedResourceDeleteMessage(ctx.ConsoleContext, s.For.RefId); err != nil {
					return errors.NewE(err)
				}
			}
		default:
			{
				d.logger.Warnf("unknown resource type %s", s.For.ResourceType)
			}
		}
	}

	if err := d.secretRepo.DeleteOne(ctx, ctx.DBFilters().Add(fields.MetadataName, secret.Name)); err != nil {
		return errors.NewE(err)
	}

	d.resourceEventPublisher.PublishResourceEvent(ctx, entities.ResourceTypeSecret, secret.Name, PublishDelete)

	return nil
}

func (d *domain) OnSecretUpdateMessage(ctx ResourceContext, secretIn entities.Secret, status types.ResourceStatus, opts UpdateAndDeleteOpts) error {
	if v, ok := secretIn.GetLabels()["app.kubernetes.io/managed-by"]; ok && v == "Helm" {
		// INFO: configmap created with Helm, we should just upsert it

		ctx.DBFilters().Add(fc.MetadataName, secretIn.Name).Add(fc.MetadataNamespace, secretIn.Namespace)

		secretIn.AccountName = ctx.AccountName
		secretIn.EnvironmentName = ctx.EnvironmentName
		secretIn.ResourceMetadata = common.ResourceMetadata{
			DisplayName:   secretIn.Name,
			CreatedBy:     common.CreatedOrUpdatedByResourceSync,
			LastUpdatedBy: common.CreatedOrUpdatedByResourceSync,
		}
		secretIn.CreatedByHelm = fn.New(fmt.Sprintf("%s/%s", secretIn.GetAnnotations()["meta.helm.sh/release-namespace"], secretIn.GetAnnotations()["meta.helm.sh/release-name"]))
		secretIn.SyncStatus = t.SyncStatus{
			LastSyncedAt:  opts.MessageTimestamp,
			Action:        t.SyncActionApply,
			RecordVersion: 0,
			State:         t.SyncStateAppliedAtAgent,
			Error:         nil,
		}

		_, err := d.secretRepo.Upsert(ctx, repos.Filter{
			fc.MetadataName:      secretIn.Name,
			fc.MetadataNamespace: secretIn.Namespace,
			fc.EnvironmentName:   ctx.EnvironmentName,
		}, &secretIn)
		if err != nil {
			return errors.NewE(err)
		}

		return nil
	}

	xSecret, err := d.findSecret(ctx, secretIn.Name)
	if err != nil {
		return errors.NewE(err)
	}

	recordVersion, err := d.MatchRecordVersion(secretIn.Annotations, xSecret.RecordVersion)
	if err != nil {
		return d.resyncK8sResource(ctx, xSecret.EnvironmentName, xSecret.SyncStatus.Action, &xSecret.Secret, xSecret.RecordVersion)
	}

	if xSecret.For != nil {
		switch xSecret.For.ResourceType {
		case entities.ResourceTypeImportedManagedResource:
			{
				if err := d.OnImportedManagedResourceUpdateMessage(ctx.ConsoleContext, xSecret.For.RefId, status, opts); err != nil {
					return errors.NewE(err)
				}
			}
		default:
			{
				d.logger.Warnf("unknown resource type %s", xSecret.For.ResourceType)
			}
		}
	}

	usecret, err := d.secretRepo.PatchById(ctx, xSecret.Id, common.PatchForSyncFromAgent(&secretIn, recordVersion, status, common.PatchOpts{
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

	return d.resyncK8sResource(ctx, secret.EnvironmentName, secret.SyncStatus.Action, &secret.Secret, secret.RecordVersion)
}
