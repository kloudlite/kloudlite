package domain

import (
	"encoding/base64"
	"encoding/json"

	iamT "github.com/kloudlite/api/apps/iam/types"

	"github.com/kloudlite/api/apps/console/internal/entities"
	fc "github.com/kloudlite/api/apps/console/internal/entities/field-constants"
	"github.com/kloudlite/api/common"
	"github.com/kloudlite/api/pkg/errors"
	"github.com/kloudlite/api/pkg/repos"
	t "github.com/kloudlite/api/pkg/types"
	"github.com/kloudlite/operator/operators/resource-watcher/types"
	"github.com/kloudlite/operator/pkg/constants"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (d *domain) ListImagePullSecrets(ctx ConsoleContext, search map[string]repos.MatchFilter, pagination repos.CursorPagination) (*repos.PaginatedRecord[*entities.ImagePullSecret], error) {
	if err := d.canPerformActionInAccount(ctx, iamT.ListImagePullSecrets); err != nil {
		return nil, errors.NewE(err)
	}

	return d.pullSecretsRepo.FindPaginated(ctx, d.pullSecretsRepo.MergeMatchFilters(entities.FilterListImagePullSecret(ctx.AccountName), search), pagination)
}

func (d *domain) findImagePullSecret(ctx ConsoleContext, name string) (*entities.ImagePullSecret, error) {
	ips, err := d.pullSecretsRepo.FindOne(ctx, entities.FilterUniqueImagePullSecret(ctx.AccountName, name))
	if err != nil {
		return nil, errors.NewE(err)
	}

	if ips == nil {
		return nil, errors.Newf("no image-pull-secret with name (%s) found", name)
	}
	return ips, nil
}

func (d *domain) GetImagePullSecret(ctx ConsoleContext, name string) (*entities.ImagePullSecret, error) {
	if err := d.canPerformActionInAccount(ctx, iamT.GetImagePullSecret); err != nil {
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

func (d *domain) CreateImagePullSecret(ctx ConsoleContext, ips entities.ImagePullSecret) (*entities.ImagePullSecret, error) {
	if err := d.canPerformActionInAccount(ctx, iamT.CreateImagePullSecret); err != nil {
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
	ips.Environments = []string{"*"}
	ips.SyncStatus = t.GenSyncStatus(t.SyncActionApply, ips.RecordVersion)

	pullSecret, err := generateImagePullSecret(ips)
	if err != nil {
		return nil, errors.NewEf(err, "failed to create a valid kubernetes ips")
	}
	ips.GeneratedK8sSecret = pullSecret

	nips, err := d.pullSecretsRepo.Create(ctx, &ips)
	if err != nil {
		if d.pullSecretsRepo.ErrAlreadyExists(err) {
			// TODO: better insights into error, when it is being caused by duplicated indexes
			return nil, errors.NewE(err)
		}
		return nil, errors.NewE(err)
	}

	d.resourceEventPublisher.PublishConsoleEvent(ctx, entities.ResourceTypeImagePullSecret, nips.Name, PublishAdd)

	if err := d.applyImagePullSecret(ctx, nips); err != nil {
		return nil, err
	}

	return nips, nil
}

func (d *domain) applyImagePullSecret(ctx ConsoleContext, ips *entities.ImagePullSecret) error {
	environments := ips.Environments

	allEnvironments := false
	for i := range ips.Environments {
		if ips.Environments[i] == "*" {
			allEnvironments = true
			break
		}
	}

	if allEnvironments {
		envs, err := d.listEnvironments(ctx)
		if err != nil {
			return err
		}
		for i := range envs {
			environments = append(environments, envs[i].Name)
		}
	}

	for i := range environments {
		if err := d.applyK8sResource(ctx, environments[i], &ips.GeneratedK8sSecret, ips.RecordVersion); err != nil {
			return err
		}
	}

	return nil
}

func (d *domain) deleteImagePullSecret(ctx ConsoleContext, ips *entities.ImagePullSecret) error {
	environments := ips.Environments

	allEnvironments := false
	for i := range ips.Environments {
		if ips.Environments[i] == "*" {
			allEnvironments = true
			break
		}
	}

	if allEnvironments {
		envs, err := d.listEnvironments(ctx)
		if err != nil {
			return err
		}
		for i := range envs {
			environments = append(environments, envs[i].Name)
		}
	}

	for i := range environments {
		if err := d.deleteK8sResource(ctx, environments[i], &corev1.Secret{
			TypeMeta:   v1.TypeMeta{APIVersion: "v1", Kind: "Secret"},
			ObjectMeta: v1.ObjectMeta{Name: ips.Name, Namespace: ips.Namespace},
		}); err != nil {
			if errors.Is(err, ErrNoClusterAttached) {
				return d.pullSecretsRepo.DeleteById(ctx, ips.Id)
			}
			return err
		}
	}

	return nil
}

func (d *domain) UpdateImagePullSecret(ctx ConsoleContext, ips entities.ImagePullSecret) (*entities.ImagePullSecret, error) {
	if err := d.canPerformActionInAccount(ctx, iamT.UpdateImagePullSecret); err != nil {
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

	patch := repos.Document{
		fc.ImagePullSecretFormat:             ips.Format,
		fc.ImagePullSecretDockerConfigJson:   ips.DockerConfigJson,
		fc.ImagePullSecretRegistryURL:        ips.RegistryURL,
		fc.ImagePullSecretRegistryUsername:   ips.RegistryUsername,
		fc.ImagePullSecretRegistryPassword:   ips.RegistryPassword,
		fc.ImagePullSecretGeneratedK8sSecret: pullSecret,
	}

	if ips.Environments != nil {
		patch[fc.ImagePullSecretEnvironments] = ips.Environments
	}

	patchForUpdate := common.PatchForUpdate(ctx, &ips, common.PatchOpts{XPatch: patch})
	upIps, err := d.pullSecretsRepo.Patch(ctx, entities.FilterUniqueImagePullSecret(ctx.AccountName, ips.Name), patchForUpdate)
	if err != nil {
		return nil, errors.NewE(err)
	}

	d.resourceEventPublisher.PublishConsoleEvent(ctx, entities.ResourceTypeImagePullSecret, upIps.Name, PublishUpdate)

	return upIps, errors.NewE(err)
}

func (d *domain) DeleteImagePullSecret(ctx ConsoleContext, name string) error {
	if err := d.canPerformActionInAccount(ctx, iamT.DeleteImagePullSecret); err != nil {
		return errors.NewE(err)
	}

	uips, err := d.pullSecretsRepo.Patch(ctx, entities.FilterUniqueImagePullSecret(ctx.AccountName, name), common.PatchForMarkDeletion())
	if err != nil {
		return errors.NewE(err)
	}

	d.resourceEventPublisher.PublishConsoleEvent(ctx, entities.ResourceTypeImagePullSecret, uips.Name, PublishUpdate)

	if err := d.deleteImagePullSecret(ctx, uips); err != nil {
		return err
	}

	//if err := d.deleteK8sResource(ctx, "", &corev1.Secret{
	//	TypeMeta:   v1.TypeMeta{APIVersion: "v1", Kind: "Secret"},
	//	ObjectMeta: v1.ObjectMeta{Name: uips.Name, Namespace: uips.Namespace},
	//}); err != nil {
	//	if errors.Is(err, ErrNoClusterAttached) {
	//		return d.pullSecretsRepo.DeleteById(ctx, uips.Id)
	//	}
	//	return err
	//}
	return nil
}

func (d *domain) OnImagePullSecretUpdateMessage(ctx ConsoleContext, ips entities.ImagePullSecret, status types.ResourceStatus, opts UpdateAndDeleteOpts) error {
	xips, err := d.findImagePullSecret(ctx, ips.Name)
	if err != nil {
		return errors.NewE(err)
	}

	if xips == nil {
		return errors.Newf("no image pull secret found")
	}

	recordVersion, err := d.MatchRecordVersion(ips.Annotations, xips.RecordVersion)
	if err != nil {
		return err
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

	d.resourceEventPublisher.PublishConsoleEvent(ctx, ips.GetResourceType(), uips.GetName(), PublishUpdate)
	return errors.NewE(err)
}

func (d *domain) OnImagePullSecretDeleteMessage(ctx ConsoleContext, ips entities.ImagePullSecret) error {
	err := d.pullSecretsRepo.DeleteOne(ctx, entities.FilterUniqueImagePullSecret(ctx.AccountName, ips.Name))
	if err != nil {
		return errors.NewE(err)
	}
	d.resourceEventPublisher.PublishConsoleEvent(ctx, ips.GetResourceType(), ips.Name, PublishDelete)
	return nil
}

func (d *domain) OnImagePullSecretApplyError(ctx ConsoleContext, errMsg string, name string, opts UpdateAndDeleteOpts) error {
	uips, err := d.pullSecretsRepo.Patch(
		ctx,
		entities.FilterUniqueImagePullSecret(ctx.AccountName, name),
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
	d.resourceEventPublisher.PublishConsoleEvent(ctx, entities.ResourceTypeImagePullSecret, uips.Name, PublishDelete)
	return nil
}

func (d *domain) ResyncImagePullSecret(ctx ConsoleContext, name string) error {
	return errors.Newf("not implemented")
	//	if err := d.canPerformActionInAccount(ctx, iamT.DeleteImagePullSecret); err != nil {
	//		return errors.NewE(err)
	//	}
	//
	//	xips, err := d.findImagePullSecret(ctx, name)
	//	if err != nil {
	//		return errors.NewE(err)
	//	}
	//
	//	return d.resyncK8sResource(ctx, xips.EnvironmentName, xips.SyncStatus.Action, &xips.GeneratedK8sSecret, xips.RecordVersion)
}
