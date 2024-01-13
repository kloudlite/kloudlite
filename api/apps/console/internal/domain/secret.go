package domain

import (
	"time"

	"github.com/kloudlite/api/apps/console/internal/entities"
	"github.com/kloudlite/api/common"
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
	filters := ctx.DBFilters()
	filters.Add("metadata.name", name)

	xSecret, err := d.secretRepo.FindOne(ctx, filters)
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
		"metadata.name": {
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

	s, err := d.secretRepo.Create(ctx, &secret)
	if err != nil {
		if d.secretRepo.ErrAlreadyExists(err) {
			// TODO: better insights into error, when it is being caused by duplicated indexes
			return nil, errors.NewE(err)
		}
		return nil, errors.NewE(err)
	}

	if s.Annotations == nil {
		s.Annotations = make(map[string]string)
	}

	for k, v := range types.SecretWatchingAnnotation {
		s.Annotations[k] = v
	}

	if err := d.applyK8sResource(ctx, s.ProjectName, &s.Secret, s.RecordVersion); err != nil {
		return s, errors.NewE(err)
	}

	return s, nil
}

func (d *domain) UpdateSecret(ctx ResourceContext, secret entities.Secret) (*entities.Secret, error) {
	if err := d.canMutateResourcesInEnvironment(ctx); err != nil {
		return nil, errors.NewE(err)
	}

	xSecret, err := d.findSecret(ctx, secret.Name)
	if err != nil {
		return nil, errors.NewE(err)
	}

	if xSecret.Type != secret.Type {
		return nil, errors.Newf("updating secret.type is forbidden")
	}

	xSecret.IncrementRecordVersion()
	xSecret.LastUpdatedBy = common.CreatedOrUpdatedBy{
		UserId:    ctx.UserId,
		UserName:  ctx.UserName,
		UserEmail: ctx.UserEmail,
	}

	xSecret.DisplayName = secret.DisplayName

	xSecret.Labels = secret.Labels
	xSecret.Annotations = secret.Annotations
	xSecret.Data = secret.Data
	xSecret.StringData = secret.StringData
	xSecret.SyncStatus = t.GenSyncStatus(t.SyncActionApply, xSecret.RecordVersion)

	upSecret, err := d.secretRepo.UpdateById(ctx, xSecret.Id, xSecret)
	if err != nil {
		return nil, errors.NewE(err)
	}

	if upSecret.Annotations == nil {
		upSecret.Annotations = make(map[string]string)
	}

	for k, v := range types.SecretWatchingAnnotation {
		upSecret.Annotations[k] = v
	}

	if err := d.applyK8sResource(ctx, upSecret.ProjectName, &upSecret.Secret, upSecret.RecordVersion); err != nil {
		return upSecret, errors.NewE(err)
	}

	return upSecret, nil
}

func (d *domain) DeleteSecret(ctx ResourceContext, name string) error {
	if err := d.canMutateResourcesInEnvironment(ctx); err != nil {
		return errors.NewE(err)
	}

	secret, err := d.findSecret(ctx, name)
	if err != nil {
		return errors.NewE(err)
	}

	patch := repos.Document{
		"markedForDeletion":          true,
		"syncStatus.syncScheduledAt": time.Now(),
		"syncStatus.action":          t.SyncActionDelete,
		"syncStatus.state":           t.SyncStateInQueue,
	}

	if _, err := d.secretRepo.PatchById(ctx, secret.Id, patch); err != nil {
		return errors.NewE(err)
	}

	return d.deleteK8sResource(ctx, secret.ProjectName, &secret.Secret)
}

func (d *domain) OnSecretDeleteMessage(ctx ResourceContext, secret entities.Secret) error {
	exSecret, err := d.findSecret(ctx, secret.Name)
	if err != nil {
		return errors.NewE(err)
	}

	return d.secretRepo.DeleteById(ctx, exSecret.Id)
}

func (d *domain) OnSecretUpdateMessage(ctx ResourceContext, secretIn entities.Secret, status types.ResourceStatus, opts UpdateAndDeleteOpts) error {
	xSecret, err := d.findSecret(ctx, secretIn.Name)
	if err != nil {
		return errors.NewE(err)
	}

	if err := d.MatchRecordVersion(secretIn.Annotations, xSecret.RecordVersion); err != nil {
		return d.resyncK8sResource(ctx, xSecret.ProjectName, xSecret.SyncStatus.Action, &xSecret.Secret, xSecret.RecordVersion)
	}

	patch := repos.Document{
		"metadata.creationTimestamp": secretIn.CreationTimestamp,
		"metadata.labels":            secretIn.Labels,
		"metadata.annotations":       secretIn.Annotations,
		"metadata.generation":        secretIn.Generation,

		"syncStatus.state": func() t.SyncState {
			if status == types.ResourceStatusDeleting {
				return t.SyncStateDeletingAtAgent
			}
			return t.SyncStateUpdatedAtAgent
		}(),
		"syncStatus.recordVersion": xSecret.RecordVersion,
		"syncStatus.lastSyncedAt":  opts.MessageTimestamp,
		"syncStatus.error":         nil,
	}

	_, err = d.secretRepo.PatchById(ctx, xSecret.Id, patch)
	return errors.NewE(err)
}

func (d *domain) OnSecretApplyError(ctx ResourceContext, errMsg, name string, opts UpdateAndDeleteOpts) error {
	xsecret, err := d.findSecret(ctx, name)
	if err != nil {
		return err
	}

	patch := repos.Document{
		"syncStatus.state":        t.SyncStateErroredAtAgent,
		"syncStatus.lastSyncedAt": opts.MessageTimestamp,
		"syncStatus.error":        errMsg,
	}

	usecret, err := d.secretRepo.PatchById(ctx, xsecret.Id, patch)
	if err != nil {
		return err
	}

	d.resourceEventPublisher.PublishSecretEvent(usecret, PublishUpdate)

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
