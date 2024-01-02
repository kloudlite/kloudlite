package domain

import (
	"github.com/kloudlite/api/apps/console/internal/entities"
	"github.com/kloudlite/api/common"
	"github.com/kloudlite/api/pkg/errors"
	"github.com/kloudlite/api/pkg/repos"
	t "github.com/kloudlite/api/pkg/types"
	"github.com/kloudlite/operator/operators/resource-watcher/types"
)

func (d *domain) ListImagePullSecrets(ctx ConsoleContext, namespace string, search map[string]repos.MatchFilter, pagination repos.CursorPagination) (*repos.PaginatedRecord[*entities.ImagePullSecret], error) {
	if err := d.canReadSecretsFromAccount(ctx, string(ctx.UserId), ctx.AccountName); err != nil {
		return nil, errors.NewE(err)
	}

	filter := repos.Filter{
		"accountName":        ctx.AccountName,
		"clusterName":        ctx.ClusterName,
		"metadata.namespace": namespace,
	}

	return d.pullSecretsRepo.FindPaginated(ctx, d.pullSecretsRepo.MergeMatchFilters(filter, search), pagination)
}

func (d *domain) findImagePullSecret(ctx ConsoleContext, namespace, name string) (*entities.ImagePullSecret, error) {
	ips, err := d.pullSecretsRepo.FindOne(ctx, repos.Filter{"accountName": ctx.AccountName, "name": name, "namespace": namespace})
	if err != nil {
		return nil, errors.NewE(err)
	}

	if ips == nil {
		return nil, errors.Newf("no image-pull-secret with name=%q, namespace=%q found", name, namespace)
	}
	return ips, nil
}

func (d *domain) GetImagePullSecret(ctx ConsoleContext, namespace, name string) (*entities.ImagePullSecret, error) {
	if err := d.canReadSecretsFromAccount(ctx, string(ctx.UserId), ctx.AccountName); err != nil {
		return nil, errors.NewE(err)
	}

	return d.findImagePullSecret(ctx, namespace, name)
}

func (d *domain) CreateImagePullSecret(ctx ConsoleContext, ips entities.ImagePullSecret) (*entities.ImagePullSecret, error) {
	if err := d.canMutateSecretsInAccount(ctx, string(ctx.UserId), ctx.AccountName); err != nil {
		return nil, errors.NewE(err)
	}

	ips.EnsureGVK()
	if err := d.k8sClient.ValidateObject(ctx, &ips.ImagePullSecret); err != nil {
		return nil, errors.NewE(err)
	}

	ips.IncrementRecordVersion()

	ips.CreatedBy = common.CreatedOrUpdatedBy{
		UserId:    ctx.UserId,
		UserName:  ctx.UserName,
		UserEmail: ctx.UserEmail,
	}
	ips.LastUpdatedBy = ips.CreatedBy

	ips.AccountName = ctx.AccountName
	ips.ClusterName = ctx.ClusterName
	ips.SyncStatus = t.GenSyncStatus(t.SyncActionApply, ips.RecordVersion)

	nIps, err := d.pullSecretsRepo.Create(ctx, &ips)
	if err != nil {
		if d.pullSecretsRepo.ErrAlreadyExists(err) {
			// TODO: better insights into error, when it is being caused by duplicated indexes
			return nil, errors.NewE(err)
		}
		return nil, errors.NewE(err)
	}

	return nIps, nil
}

func (d *domain) UpdateImagePullSecret(ctx ConsoleContext, ips entities.ImagePullSecret) (*entities.ImagePullSecret, error) {
	if err := d.canMutateSecretsInAccount(ctx, string(ctx.UserId), ctx.AccountName); err != nil {
		return nil, errors.NewE(err)
	}

	ips.EnsureGVK()
	if err := d.k8sClient.ValidateObject(ctx, &ips.ImagePullSecret); err != nil {
		return nil, errors.NewE(err)
	}

	currScrt, err := d.findImagePullSecret(ctx, ips.Namespace, ips.Name)
	if err != nil {
		return nil, errors.NewE(err)
	}

	currScrt.IncrementRecordVersion()

	currScrt.LastUpdatedBy = common.CreatedOrUpdatedBy{
		UserId:    ctx.UserId,
		UserName:  ctx.UserName,
		UserEmail: ctx.UserEmail,
	}
	currScrt.DisplayName = ips.DisplayName

	currScrt.Annotations = ips.Annotations
	currScrt.Labels = ips.Labels

	currScrt.Spec = ips.Spec
	currScrt.SyncStatus = t.GenSyncStatus(t.SyncActionApply, currScrt.RecordVersion)

	upIps, err := d.pullSecretsRepo.UpdateById(ctx, currScrt.Id, currScrt)
	if err != nil {
		return nil, errors.NewE(err)
	}

	if err := d.applyK8sResource(ctx, &upIps.ImagePullSecret, ips.RecordVersion); err != nil {
		return nil, errors.NewE(err)
	}

	return upIps, errors.NewE(err)
}

func (d *domain) DeleteImagePullSecret(ctx ConsoleContext, namespace, name string) error {
	if err := d.canMutateSecretsInAccount(ctx, string(ctx.UserId), ctx.AccountName); err != nil {
		return errors.NewE(err)
	}

	ips, err := d.findImagePullSecret(ctx, namespace, name)
	if err != nil {
		return errors.NewE(err)
	}

	ips.SyncStatus = t.GenSyncStatus(t.SyncActionDelete, ips.RecordVersion)

	if _, err := d.pullSecretsRepo.UpdateById(ctx, ips.Id, ips); err != nil {
		return errors.NewE(err)
	}

	return d.deleteK8sResource(ctx, &ips.ImagePullSecret)
}

func (d *domain) OnImagePullSecretUpdateMessage(ctx ConsoleContext, ips entities.ImagePullSecret, status types.ResourceStatus, opts UpdateAndDeleteOpts) error {
	exIps, err := d.findImagePullSecret(ctx, ips.Namespace, ips.Name)
	if err != nil {
		return errors.NewE(err)
	}

	if err := d.MatchRecordVersion(ips.Annotations, exIps.RecordVersion); err != nil {
		return d.resyncK8sResource(ctx, exIps.SyncStatus.Action, &exIps.ImagePullSecret, exIps.RecordVersion)
	}

	exIps.CreationTimestamp = ips.CreationTimestamp
	exIps.Labels = ips.Labels
	exIps.Annotations = ips.Annotations
	exIps.Generation = ips.Generation

	exIps.Status = ips.Status

	exIps.SyncStatus.State = func() t.SyncState {
		if status == types.ResourceStatusDeleting {
			return t.SyncStateDeletingAtAgent
		}
		return t.SyncStateUpdatedAtAgent
	}()
	exIps.SyncStatus.RecordVersion = exIps.RecordVersion
	exIps.SyncStatus.Error = nil
	exIps.SyncStatus.LastSyncedAt = opts.MessageTimestamp

	_, err = d.pullSecretsRepo.UpdateById(ctx, exIps.Id, exIps)
	return errors.NewE(err)
}

func (d *domain) OnImagePullSecretDeleteMessage(ctx ConsoleContext, ips entities.ImagePullSecret) error {
	a, err := d.findImagePullSecret(ctx, ips.Namespace, ips.Name)
	if err != nil {
		return errors.NewE(err)
	}

	if err := d.MatchRecordVersion(ips.Annotations, a.RecordVersion); err != nil {
		return d.resyncK8sResource(ctx, a.SyncStatus.Action, &a.ImagePullSecret, a.RecordVersion)
	}

	return d.pullSecretsRepo.DeleteById(ctx, a.Id)
}

func (d *domain) OnImagePullSecretApplyError(ctx ConsoleContext, errMsg string, namespace string, name string, opts UpdateAndDeleteOpts) error {
	a, err2 := d.findImagePullSecret(ctx, namespace, name)
	if err2 != nil {
		return err2
	}

	a.SyncStatus.State = t.SyncStateErroredAtAgent
	a.SyncStatus.LastSyncedAt = opts.MessageTimestamp
	a.SyncStatus.Error = &errMsg
	_, err := d.pullSecretsRepo.UpdateById(ctx, a.Id, a)
	return errors.NewE(err)
}

func (d *domain) ResyncImagePullSecret(ctx ConsoleContext, namespace, name string) error {
	if err := d.canMutateResourcesInWorkspace(ctx, namespace); err != nil {
		return errors.NewE(err)
	}

	exIps, err := d.findImagePullSecret(ctx, namespace, name)
	if err != nil {
		return errors.NewE(err)
	}
	return d.resyncK8sResource(ctx, exIps.SyncStatus.Action, &exIps.ImagePullSecret, exIps.RecordVersion)
}
