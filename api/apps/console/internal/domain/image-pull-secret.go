package domain

import (
	"encoding/base64"
	"encoding/json"
	"time"

	"github.com/kloudlite/api/apps/console/internal/entities"
	"github.com/kloudlite/api/common"
	"github.com/kloudlite/api/pkg/errors"
	"github.com/kloudlite/api/pkg/repos"
	t "github.com/kloudlite/api/pkg/types"
	"github.com/kloudlite/operator/operators/resource-watcher/types"
	"github.com/kloudlite/operator/pkg/constants"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (d *domain) ListImagePullSecrets(ctx ResourceContext, search map[string]repos.MatchFilter, pagination repos.CursorPagination) (*repos.PaginatedRecord[*entities.ImagePullSecret], error) {
	if err := d.canReadSecretsFromAccount(ctx, string(ctx.UserId), ctx.AccountName); err != nil {
		return nil, errors.NewE(err)
	}

	filters := ctx.DBFilters()

	return d.pullSecretsRepo.FindPaginated(ctx, d.pullSecretsRepo.MergeMatchFilters(filters, search), pagination)
}

func (d *domain) findImagePullSecret(ctx ResourceContext, name string) (*entities.ImagePullSecret, error) {
	filters := ctx.DBFilters()
	filters.Add("metadata.name", name)

	ips, err := d.pullSecretsRepo.FindOne(ctx, filters)
	if err != nil {
		return nil, errors.NewE(err)
	}

	if ips == nil {
		return nil, errors.Newf("no image-pull-secret with name (%s) found", name)
	}
	return ips, nil
}

func (d *domain) GetImagePullSecret(ctx ResourceContext, name string) (*entities.ImagePullSecret, error) {
	if err := d.canReadSecretsFromAccount(ctx, string(ctx.UserId), ctx.AccountName); err != nil {
		return nil, errors.NewE(err)
	}

	return d.findImagePullSecret(ctx, name)
}

func generateImagePullSecret(ips entities.ImagePullSecret) (corev1.Secret, error) {
	if err := ips.Validate(); err != nil {
		return corev1.Secret{}, errors.NewE(err)
	}

	data := map[string][]byte{}
	switch ips.Format {
	case entities.DockerConfigJsonFormat:
		data[corev1.DockerConfigJsonKey] = []byte(*ips.DockerConfigJson)
	case entities.ParamsFormat:
		m := map[string]any{
			"auths": map[string]any{
				*ips.RegistryURL: map[string]any{
					"username": *ips.RegistryUsername,
					"password": *ips.RegistryPassword,
				},
			},
		}
		b, err := json.Marshal(m)
		if err != nil {
			return corev1.Secret{}, err
		}

		data[corev1.DockerConfigJsonKey] = []byte(base64.StdEncoding.EncodeToString(b))
	}

	secret := corev1.Secret{
		TypeMeta: v1.TypeMeta{APIVersion: "v1", Kind: "Secret"},
		ObjectMeta: v1.ObjectMeta{
			Name:      ips.Name,
			Namespace: ips.Namespace,
			Labels:    map[string]string{},
			Annotations: map[string]string{
				constants.DescriptionKey: "This resource is managed by kloudlite.io control plane. This secret is created as part of image-pull-secret resource.",
			},
		},
		Data: data,
		Type: corev1.SecretTypeDockercfg,
	}

	return secret, nil
}

func (d *domain) CreateImagePullSecret(ctx ResourceContext, ips entities.ImagePullSecret) (*entities.ImagePullSecret, error) {
	if err := d.canMutateSecretsInAccount(ctx, string(ctx.UserId), ctx.AccountName); err != nil {
		return nil, errors.NewE(err)
	}

	if err := ips.Validate(); err != nil {
		return nil, errors.NewE(err)
	}

	env, err := d.findEnvironment(ctx.ConsoleContext, ctx.ProjectName, ctx.EnvironmentName)
	if err != nil {
		return nil, errors.NewE(err)
	}

	ips.Namespace = env.Spec.TargetNamespace

	ips.IncrementRecordVersion()

	ips.CreatedBy = common.CreatedOrUpdatedBy{
		UserId:    ctx.UserId,
		UserName:  ctx.UserName,
		UserEmail: ctx.UserEmail,
	}
	ips.LastUpdatedBy = ips.CreatedBy

	ips.AccountName = ctx.AccountName
	ips.ProjectName = ctx.ProjectName
	ips.EnvironmentName = ctx.EnvironmentName
	ips.SyncStatus = t.GenSyncStatus(t.SyncActionApply, ips.RecordVersion)

	pullSecret, err := generateImagePullSecret(ips)
	if err != nil {
		return nil, errors.NewEf(err, "failed to create a valid kubernetes secret")
	}
	ips.GeneratedK8sSecret = pullSecret

	if _, err := d.upsertEnvironmentResourceMapping(ctx, &ips); err != nil {
		return nil, errors.NewE(err)
	}

	nips, err := d.pullSecretsRepo.Create(ctx, &ips)
	if err != nil {
		if d.pullSecretsRepo.ErrAlreadyExists(err) {
			// TODO: better insights into error, when it is being caused by duplicated indexes
			return nil, errors.NewE(err)
		}
		return nil, errors.NewE(err)
	}

	if err := d.applyK8sResource(ctx, ips.ProjectName, &pullSecret, ips.RecordVersion); err != nil {
		return nil, errors.NewE(err)
	}

	return nips, nil
}

func (d *domain) UpdateImagePullSecret(ctx ResourceContext, ips entities.ImagePullSecret) (*entities.ImagePullSecret, error) {
	if err := d.canMutateSecretsInAccount(ctx, string(ctx.UserId), ctx.AccountName); err != nil {
		return nil, errors.NewE(err)
	}

	if err := ips.Validate(); err != nil {
		return nil, errors.NewE(err)
	}

	xips, err := d.findImagePullSecret(ctx, ips.Name)
	if err != nil {
		return nil, errors.NewE(err)
	}

	pullSecret, err := generateImagePullSecret(*xips)
	if err != nil {
		return nil, errors.NewE(err)
	}

	patch := repos.Document{
		"recordVersion": ips.RecordVersion + 1,
		"displayName":   ips.DisplayName,
		"lastUpdatedBy": common.CreatedOrUpdatedBy{
			UserId:    ctx.UserId,
			UserName:  ctx.UserName,
			UserEmail: ctx.UserEmail,
		},

		"format":           ips.Format,
		"dockerConfigJson": ips.DockerConfigJson,

		"registryURL":      ips.RegistryURL,
		"registryUsername": ips.RegistryUsername,
		"registryPassword": ips.RegistryPassword,

		"metadata.labels":      ips.Labels,
		"metadata.annotations": ips.Annotations,

		"generatedK8sSecret": pullSecret,

		"syncStatus.state":           t.SyncStateInQueue,
		"syncStatus.syncScheduledAt": time.Now(),
		"syncStatus.action":          t.SyncActionApply,
	}

	upIps, err := d.pullSecretsRepo.PatchById(ctx, xips.Id, patch)
	if err != nil {
		return nil, errors.NewE(err)
	}

	if err := d.applyK8sResource(ctx, xips.ProjectName, &pullSecret, xips.RecordVersion); err != nil {
		return nil, errors.NewE(err)
	}

	d.resourceEventPublisher.PublishImagePullSecretEvent(upIps, PublishUpdate)

	return upIps, errors.NewE(err)
}

func (d *domain) DeleteImagePullSecret(ctx ResourceContext, name string) error {
	if err := d.canMutateResourcesInEnvironment(ctx); err != nil {
		return errors.NewE(err)
	}

	ips, err := d.findImagePullSecret(ctx, name)
	if err != nil {
		return errors.NewE(err)
	}

	patch := repos.Document{
		"markedForDeletion":          true,
		"syncStatus.syncScheduledAt": time.Now(),
		"syncStatus.action":          t.SyncActionDelete,
		"syncStatus.state":           t.SyncStateInQueue,
	}

	uips, err := d.pullSecretsRepo.PatchById(ctx, ips.Id, patch)
	if err != nil {
		return errors.NewE(err)
	}

	if err := d.deleteK8sResource(ctx, ips.ProjectName, &corev1.Secret{
		TypeMeta:   v1.TypeMeta{APIVersion: "v1", Kind: "Secret"},
		ObjectMeta: v1.ObjectMeta{Name: ips.Name, Namespace: ips.Namespace},
	}); err != nil {
		if errors.Is(err, ErrNoClusterAttached) {
			return d.pullSecretsRepo.DeleteById(ctx, ips.Id)
		}
		return err
	}

	d.resourceEventPublisher.PublishImagePullSecretEvent(uips, PublishDelete)

	return nil
}

func (d *domain) OnImagePullSecretUpdateMessage(ctx ResourceContext, ips entities.ImagePullSecret, status types.ResourceStatus, opts UpdateAndDeleteOpts) error {
	xips, err := d.findImagePullSecret(ctx, ips.Name)
	if err != nil {
		return errors.NewE(err)
	}

	if err := d.MatchRecordVersion(ips.Annotations, xips.RecordVersion); err != nil {
		return d.resyncK8sResource(ctx, xips.ProjectName, xips.SyncStatus.Action, &xips.GeneratedK8sSecret, xips.RecordVersion)
	}

	patch := repos.Document{
		"metadata.creationTimestamp": ips.CreationTimestamp,
		"metadata.labels":            ips.Labels,
		"metadata.annotations":       ips.Annotations,
		"metadata.generation":        ips.Generation,

		"syncStatus.state": func() t.SyncState {
			if status == types.ResourceStatusDeleting {
				return t.SyncStateDeletingAtAgent
			}
			return t.SyncStateUpdatedAtAgent
		}(),
		"syncStatus.recordVersion": xips.RecordVersion,
		"syncStatus.lastSyncedAt":  opts.MessageTimestamp,
		"syncStatus.error":         nil,
	}

	uips, err := d.pullSecretsRepo.PatchById(ctx, xips.Id, patch)
	if err != nil {
		return err
	}

	d.resourceEventPublisher.PublishImagePullSecretEvent(uips, PublishUpdate)
	return errors.NewE(err)
}

func (d *domain) OnImagePullSecretDeleteMessage(ctx ResourceContext, ips entities.ImagePullSecret) error {
	xips, err := d.findImagePullSecret(ctx, ips.Name)
	if err != nil {
		return errors.NewE(err)
	}

	return d.pullSecretsRepo.DeleteById(ctx, xips.Id)
}

func (d *domain) OnImagePullSecretApplyError(ctx ResourceContext, errMsg string, name string, opts UpdateAndDeleteOpts) error {
	ips, err := d.findImagePullSecret(ctx, name)
	if err != nil {
		return err
	}

	patch := repos.Document{
		"syncStatus.state":        t.SyncStateErroredAtAgent,
		"syncStatus.lastSyncedAt": opts.MessageTimestamp,
		"syncStatus.error":        errMsg,
	}

	_, err = d.pullSecretsRepo.PatchById(ctx, ips.Id, patch)
	return errors.NewE(err)
}

func (d *domain) ResyncImagePullSecret(ctx ResourceContext, name string) error {
	if err := d.canMutateResourcesInEnvironment(ctx); err != nil {
		return errors.NewE(err)
	}

	xips, err := d.findImagePullSecret(ctx, name)
	if err != nil {
		return errors.NewE(err)
	}
	return d.resyncK8sResource(ctx, xips.ProjectName, xips.SyncStatus.Action, &xips.GeneratedK8sSecret, xips.RecordVersion)
}
