package domain

import (
	"fmt"
	"time"

	"kloudlite.io/apps/console/internal/entities"
	"kloudlite.io/common"
	"kloudlite.io/pkg/repos"
	t "kloudlite.io/pkg/types"
)

func (d *domain) ListImagePullSecrets(ctx ConsoleContext, namespace string, search map[string]repos.MatchFilter, pagination repos.CursorPagination) (*repos.PaginatedRecord[*entities.ImagePullSecret], error) {
	if err := d.canReadSecretsFromAccount(ctx, string(ctx.UserId), ctx.AccountName); err != nil {
		return nil, err
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
		return nil, err
	}

	if ips == nil {
		return nil, fmt.Errorf("no image-pull-secret with name=%q, namespace=%q found", name, namespace)
	}
	return ips, nil
}

func (d *domain) GetImagePullSecret(ctx ConsoleContext, namespace, name string) (*entities.ImagePullSecret, error) {
	if err := d.canReadSecretsFromAccount(ctx, string(ctx.UserId), ctx.AccountName); err != nil {
		return nil, err
	}

	return d.findImagePullSecret(ctx, namespace, name)
}

func (d *domain) CreateImagePullSecret(ctx ConsoleContext, ips entities.ImagePullSecret) (*entities.ImagePullSecret, error) {
	if err := d.canMutateSecretsInAccount(ctx, string(ctx.UserId), ctx.AccountName); err != nil {
		return nil, err
	}

	ips.EnsureGVK()
	if err := d.k8sClient.ValidateObject(ctx, &ips.ImagePullSecret); err != nil {
		return nil, err
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
			return nil, err
		}
		return nil, err
	}

	return nIps, nil
}

func (d *domain) UpdateImagePullSecret(ctx ConsoleContext, ips entities.ImagePullSecret) (*entities.ImagePullSecret, error) {
	if err := d.canMutateSecretsInAccount(ctx, string(ctx.UserId), ctx.AccountName); err != nil {
		return nil, err
	}

	ips.EnsureGVK()
	if err := d.k8sClient.ValidateObject(ctx, &ips.ImagePullSecret); err != nil {
		return nil, err
	}

	currScrt, err := d.findImagePullSecret(ctx, ips.Namespace, ips.Name)
	if err != nil {
		return nil, err
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
		return nil, err
	}

	if err := d.applyK8sResource(ctx, &upIps.ImagePullSecret, ips.RecordVersion); err != nil {
		return nil, err
	}

	return upIps, err
}

func (d *domain) DeleteImagePullSecret(ctx ConsoleContext, namespace, name string) error {
	if err := d.canMutateSecretsInAccount(ctx, string(ctx.UserId), ctx.AccountName); err != nil {
		return err
	}

	ips, err := d.findImagePullSecret(ctx, namespace, name)
	if err != nil {
		return err
	}

	ips.SyncStatus = t.GenSyncStatus(t.SyncActionDelete, ips.RecordVersion)

	if _, err := d.pullSecretsRepo.UpdateById(ctx, ips.Id, ips); err != nil {
		return err
	}

	return d.deleteK8sResource(ctx, &ips.ImagePullSecret)
}

func (d *domain) OnUpdateImagePullSecretMessage(ctx ConsoleContext, ips entities.ImagePullSecret) error {
	exIps, err := d.findImagePullSecret(ctx, ips.Namespace, ips.Name)
	if err != nil {
		return err
	}

	annotatedVersion, err := d.parseRecordVersionFromAnnotations(ips.Annotations)
	if err != nil {
		return d.resyncK8sResource(ctx, exIps.SyncStatus.Action, &exIps.ImagePullSecret, exIps.RecordVersion)
	}

	if annotatedVersion != exIps.RecordVersion {
		return d.resyncK8sResource(ctx, exIps.SyncStatus.Action, &exIps.ImagePullSecret, exIps.RecordVersion)
	}

	exIps.CreationTimestamp = ips.CreationTimestamp
	exIps.Labels = ips.Labels
	exIps.Annotations = ips.Annotations
	exIps.Generation = ips.Generation

	exIps.Status = ips.Status

	exIps.SyncStatus.State = t.SyncStateReceivedUpdateFromAgent
	exIps.SyncStatus.RecordVersion = annotatedVersion
	exIps.SyncStatus.Error = nil
	exIps.SyncStatus.LastSyncedAt = time.Now()

	_, err = d.pullSecretsRepo.UpdateById(ctx, exIps.Id, exIps)
	return err
}

func (d *domain) OnDeleteImagePullSecretMessage(ctx ConsoleContext, ips entities.ImagePullSecret) error {
	a, err := d.findImagePullSecret(ctx, ips.Namespace, ips.Name)
	if err != nil {
		return err
	}

	return d.pullSecretsRepo.DeleteById(ctx, a.Id)
}

func (d *domain) OnApplyImagePullSecretError(ctx ConsoleContext, errMsg string, namespace string, name string) error {
	a, err2 := d.findImagePullSecret(ctx, namespace, name)
	if err2 != nil {
		return err2
	}

	a.SyncStatus.State = t.SyncStateErroredAtAgent
	a.SyncStatus.LastSyncedAt = time.Now()
	a.SyncStatus.Error = &errMsg
	_, err := d.pullSecretsRepo.UpdateById(ctx, a.Id, a)
	return err
}

func (d *domain) ResyncImagePullSecret(ctx ConsoleContext, namespace, name string) error {
	if err := d.canMutateResourcesInWorkspace(ctx, namespace); err != nil {
		return err
	}

	exIps, err := d.findImagePullSecret(ctx, namespace, name)
	if err != nil {
		return err
	}
	return d.resyncK8sResource(ctx, exIps.SyncStatus.Action, &exIps.ImagePullSecret, exIps.RecordVersion)
}
