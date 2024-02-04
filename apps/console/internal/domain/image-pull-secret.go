package domain

import (
	"encoding/base64"
	"encoding/json"
	"github.com/kloudlite/api/apps/console/internal/entities"
	fc "github.com/kloudlite/api/apps/console/internal/entities/field-constants"
	"github.com/kloudlite/api/common"
	"github.com/kloudlite/api/common/fields"
	"github.com/kloudlite/api/pkg/errors"
	"github.com/kloudlite/api/pkg/repos"
	t "github.com/kloudlite/api/pkg/types"
	"github.com/kloudlite/operator/operators/resource-watcher/types"
	"github.com/kloudlite/operator/pkg/constants"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (d *domain) ListImagePullSecrets(ctx ResourceContext, search map[string]repos.MatchFilter, pagination repos.CursorPagination) (*repos.PaginatedRecord[*entities.ImagePullSecret], error) {
	if err := d.canReadResourcesInEnvironment(ctx); err != nil {
		return nil, errors.NewE(err)
	}

	filters := ctx.DBFilters()

	return d.pullSecretsRepo.FindPaginated(ctx, d.pullSecretsRepo.MergeMatchFilters(filters, search), pagination)
}

func (d *domain) findImagePullSecret(ctx ResourceContext, name string) (*entities.ImagePullSecret, error) {

	ips, err := d.pullSecretsRepo.FindOne(
		ctx,
		ctx.DBFilters().Add(fields.MetadataName, name),
	)
	if err != nil {
		return nil, errors.NewE(err)
	}

	if ips == nil {
		return nil, errors.Newf("no image-pull-secret with name (%s) found", name)
	}
	return ips, nil
}

func (d *domain) GetImagePullSecret(ctx ResourceContext, name string) (*entities.ImagePullSecret, error) {
	if err := d.canReadResourcesInEnvironment(ctx); err != nil {
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
		Type: corev1.SecretTypeDockerConfigJson,
	}

	return secret, nil
}

func (d *domain) CreateImagePullSecret(ctx ResourceContext, ips entities.ImagePullSecret) (*entities.ImagePullSecret, error) {
	if err := d.canMutateResourcesInEnvironment(ctx); err != nil {
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

	d.resourceEventPublisher.PublishResourceEvent(ctx, entities.ResourceTypeImagePullSecret, nips.Name, PublishAdd)

	if err := d.applyK8sResource(ctx, nips.ProjectName, &pullSecret, nips.RecordVersion); err != nil {
		return nil, errors.NewE(err)
	}

	return nips, nil
}

func (d *domain) UpdateImagePullSecret(ctx ResourceContext, ips entities.ImagePullSecret) (*entities.ImagePullSecret, error) {
	if err := d.canMutateResourcesInEnvironment(ctx); err != nil {
		return nil, errors.NewE(err)
	}

	if err := ips.Validate(); err != nil {
		return nil, errors.NewE(err)
	}

	xips, err := d.findImagePullSecret(ctx, ips.Name)
	if err != nil {
		return nil, errors.NewE(err)
	}

	if xips == nil {
		return nil, errors.Newf("no image pull secret found")
	}

	pullSecret, err := generateImagePullSecret(*xips)
	if err != nil {
		return nil, errors.NewE(err)
	}

	patchForUpdate := common.PatchForUpdate(
		ctx,
		&ips,
		common.PatchOpts{
			XPatch: repos.Document{
				fc.ImagePullSecretFormat:             ips.Format,
				fc.ImagePullSecretDockerConfigJson:   ips.DockerConfigJson,
				fc.ImagePullSecretRegistryURL:        ips.RegistryURL,
				fc.ImagePullSecretRegistryUsername:   ips.RegistryUsername,
				fc.ImagePullSecretRegistryPassword:   ips.RegistryPassword,
				fc.ImagePullSecretGeneratedK8sSecret: pullSecret,
			},
		},
	)
	upIps, err := d.pullSecretsRepo.Patch(
		ctx,
		ctx.DBFilters().Add(fields.MetadataName, ips.Name),
		patchForUpdate,
	)
	if err != nil {
		return nil, errors.NewE(err)
	}
	d.resourceEventPublisher.PublishResourceEvent(ctx, entities.ResourceTypeImagePullSecret, upIps.Name, PublishUpdate)

	if err := d.applyK8sResource(ctx, upIps.ProjectName, &pullSecret, upIps.RecordVersion); err != nil {
		return nil, errors.NewE(err)
	}

	return upIps, errors.NewE(err)
}

func (d *domain) DeleteImagePullSecret(ctx ResourceContext, name string) error {
	if err := d.canMutateResourcesInEnvironment(ctx); err != nil {
		return errors.NewE(err)
	}

	uips, err := d.pullSecretsRepo.Patch(
		ctx,
		ctx.DBFilters().Add(fields.MetadataName, name),
		common.PatchForMarkDeletion(),
	)
	if err != nil {
		return errors.NewE(err)
	}

	d.resourceEventPublisher.PublishResourceEvent(ctx, entities.ResourceTypeApp, uips.Name, PublishUpdate)

	if err := d.deleteK8sResource(ctx, uips.ProjectName, &corev1.Secret{
		TypeMeta:   v1.TypeMeta{APIVersion: "v1", Kind: "Secret"},
		ObjectMeta: v1.ObjectMeta{Name: uips.Name, Namespace: uips.Namespace},
	}); err != nil {
		if errors.Is(err, ErrNoClusterAttached) {
			return d.pullSecretsRepo.DeleteById(ctx, uips.Id)
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

	if xips == nil {
		return errors.Newf("no image pull secret found")
	}

	recordVersion, err := d.MatchRecordVersion(ips.Annotations, xips.RecordVersion)
	if err != nil {
		return d.resyncK8sResource(ctx, xips.ProjectName, xips.SyncStatus.Action, &xips.GeneratedK8sSecret, xips.RecordVersion)
	}

	uips, err := d.pullSecretsRepo.PatchById(
		ctx,
		xips.Id,
		common.PatchForSyncFromAgent(&ips, recordVersion, status, common.PatchOpts{
			MessageTimestamp: opts.MessageTimestamp,
		}))

	if err != nil {
		return err
	}

	d.resourceEventPublisher.PublishResourceEvent(ctx, uips.GetResourceType(), uips.GetName(), PublishUpdate)
	return errors.NewE(err)
}

func (d *domain) OnImagePullSecretDeleteMessage(ctx ResourceContext, ips entities.ImagePullSecret) error {
	err := d.pullSecretsRepo.DeleteOne(
		ctx,
		ctx.DBFilters().Add(fields.MetadataName, ips.Name),
	)
	if err != nil {
		return errors.NewE(err)
	}
	d.resourceEventPublisher.PublishResourceEvent(ctx, entities.ResourceTypeImagePullSecret, ips.Name, PublishDelete)
	return nil
}

func (d *domain) OnImagePullSecretApplyError(ctx ResourceContext, errMsg string, name string, opts UpdateAndDeleteOpts) error {
	uips, err := d.pullSecretsRepo.Patch(
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
		return err
	}
	d.resourceEventPublisher.PublishResourceEvent(ctx, entities.ResourceTypeImagePullSecret, uips.Name, PublishDelete)
	return nil
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
