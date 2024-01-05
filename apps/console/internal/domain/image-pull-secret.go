package domain

import (
	"encoding/base64"
	"encoding/json"
	"github.com/kloudlite/api/apps/console/internal/entities"
	"github.com/kloudlite/api/common"
	"github.com/kloudlite/api/pkg/errors"
	"github.com/kloudlite/api/pkg/repos"
	t "github.com/kloudlite/api/pkg/types"
	"github.com/kloudlite/operator/operators/resource-watcher/types"
	"github.com/kloudlite/operator/pkg/constants"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1"
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

	if _, err := d.upsertResourceMapping(ctx, &ips); err != nil {
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

	xips.IncrementRecordVersion()

	xips.LastUpdatedBy = common.CreatedOrUpdatedBy{
		UserId:    ctx.UserId,
		UserName:  ctx.UserName,
		UserEmail: ctx.UserEmail,
	}
	xips.DisplayName = ips.DisplayName

	xips.Annotations = ips.Annotations
	xips.Labels = ips.Labels

	xips.SyncStatus = t.GenSyncStatus(t.SyncActionApply, xips.RecordVersion)

	xips.Format = ips.Format
	xips.DockerConfigJson = ips.DockerConfigJson

	xips.RegistryURL = ips.RegistryURL
	xips.RegistryUsername = ips.RegistryUsername
	xips.RegistryPassword = ips.RegistryPassword

	pullSecret, err := generateImagePullSecret(*xips)
	if err != nil {
		return nil, errors.NewE(err)
	}

	xips.GeneratedK8sSecret = pullSecret

	upIps, err := d.pullSecretsRepo.UpdateById(ctx, xips.Id, xips)
	if err != nil {
		return nil, errors.NewE(err)
	}

	if err := d.applyK8sResource(ctx, xips.ProjectName, &pullSecret, xips.RecordVersion); err != nil {
		return nil, errors.NewE(err)
	}

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

	ips.SyncStatus = t.GenSyncStatus(t.SyncActionDelete, ips.RecordVersion)

	if _, err := d.pullSecretsRepo.UpdateById(ctx, ips.Id, ips); err != nil {
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

	xips.CreationTimestamp = ips.CreationTimestamp
	xips.Labels = ips.Labels
	xips.Annotations = ips.Annotations
	xips.Generation = ips.Generation

	xips.SyncStatus.State = func() t.SyncState {
		if status == types.ResourceStatusDeleting {
			return t.SyncStateDeletingAtAgent
		}
		return t.SyncStateUpdatedAtAgent
	}()
	xips.SyncStatus.RecordVersion = xips.RecordVersion
	xips.SyncStatus.Error = nil
	xips.SyncStatus.LastSyncedAt = opts.MessageTimestamp

	_, err = d.pullSecretsRepo.UpdateById(ctx, xips.Id, xips)
	return errors.NewE(err)
}

func (d *domain) OnImagePullSecretDeleteMessage(ctx ResourceContext, ips entities.ImagePullSecret) error {
	xips, err := d.findImagePullSecret(ctx, ips.Name)
	if err != nil {
		return errors.NewE(err)
	}

	if err := d.MatchRecordVersion(ips.Annotations, xips.RecordVersion); err != nil {
		return d.resyncK8sResource(ctx, xips.ProjectName, xips.SyncStatus.Action, &xips.GeneratedK8sSecret, xips.RecordVersion)
	}

	return d.pullSecretsRepo.DeleteById(ctx, xips.Id)
}

func (d *domain) OnImagePullSecretApplyError(ctx ResourceContext, errMsg string, name string, opts UpdateAndDeleteOpts) error {
	ips, err := d.findImagePullSecret(ctx, name)
	if err != nil {
		return err
	}

	ips.SyncStatus.State = t.SyncStateErroredAtAgent
	ips.SyncStatus.LastSyncedAt = opts.MessageTimestamp
	ips.SyncStatus.Error = &errMsg
	_, err = d.pullSecretsRepo.UpdateById(ctx, ips.Id, ips)
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
