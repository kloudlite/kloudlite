package domain

import (
	"fmt"
	iamT "kloudlite.io/apps/iam/types"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"kloudlite.io/apps/infra/internal/entities"
	fn "kloudlite.io/pkg/functions"
	"kloudlite.io/pkg/repos"
	"kloudlite.io/pkg/types"
)

func (d *domain) CreateProviderSecret(ctx InfraContext, secret entities.CloudProviderSecret) (*entities.CloudProviderSecret, error) {
	if err := d.canPerformActionInAccount(ctx, iamT.CreateCloudProviderSecret); err != nil {
		return nil, err
	}
	secret.EnsureGVK()

	secret.AccountName = ctx.AccountName
	secret.Namespace = d.getAccountNamespace(ctx.AccountName)

	if err := d.k8sExtendedClient.ValidateStruct(ctx, &secret.Secret); err != nil {
		return nil, err
	}

	secret.IncrementRecordVersion()
	cSecret := corev1.Secret{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "Secret",
		},
		ObjectMeta: secret.ObjectMeta,
		Data:       secret.Data,
		StringData: secret.StringData,
		Type:       secret.Type,
	}

	if err := d.ensureNamespaceForAccount(ctx, ctx.AccountName); err != nil {
		return nil, err
	}

	if err := d.applyK8sResource(ctx, &cSecret, secret.RecordVersion); err != nil {
		return nil, err
	}

	secret.Status.IsReady = true
	nSecret, err := d.secretRepo.Create(ctx, &secret)
	if err != nil {
		return nil, err
	}

	return nSecret, nil
}

func (d *domain) UpdateProviderSecret(ctx InfraContext, secret entities.CloudProviderSecret) (*entities.CloudProviderSecret, error) {
	if err := d.canPerformActionInAccount(ctx, iamT.UpdateCloudProviderSecret); err != nil {
		return nil, err
	}
	secret.EnsureGVK()
	secret.Namespace = d.env.ProviderSecretNamespace

	if err := d.k8sExtendedClient.ValidateStruct(ctx, &secret.Secret); err != nil {
		return nil, err
	}

	scrt, err := d.findProviderSecret(ctx, secret.Name)
	if err != nil {
		return nil, err
	}

	scrt.IncrementRecordVersion()
	scrt.Labels = secret.Labels
	scrt.Annotations = secret.Annotations
	scrt.Secret.Data = secret.Secret.Data
	scrt.Secret.StringData = secret.Secret.StringData

	uScrt, err := d.secretRepo.UpdateById(ctx, scrt.Id, scrt)
	if err != nil {
		return nil, err
	}

	cSecret := corev1.Secret{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "Secret",
		},
		ObjectMeta: uScrt.ObjectMeta,
		Data:       uScrt.Data,
		StringData: uScrt.StringData,
		Type:       uScrt.Type,
	}

	if err := d.applyK8sResource(ctx, &cSecret, uScrt.RecordVersion); err != nil {
		return nil, err
	}

	return uScrt, nil
}

func (d *domain) DeleteProviderSecret(ctx InfraContext, secretName string) error {
	if err := d.canPerformActionInAccount(ctx, iamT.DeleteCloudProviderSecret); err != nil {
		return err
	}
	cps, err := d.findProviderSecret(ctx, secretName)
	if err != nil {
		return err
	}

	if cps.IsMarkedForDeletion() {
		return fmt.Errorf("cloud provider secret %q is already marked for deletion", secretName)
	}

	cps.MarkedForDeletion = fn.New(true)
	if _, err := d.secretRepo.UpdateById(ctx, cps.Id, cps); err != nil {
		return err
	}
	return nil
}

func (d *domain) ListProviderSecrets(ctx InfraContext, search *repos.SearchFilter, pagination types.CursorPagination) (*repos.PaginatedRecord[*entities.CloudProviderSecret], error) {
	if err := d.canPerformActionInAccount(ctx, iamT.ListCloudProviderSecrets); err != nil {
		return nil, err
	}
	filter := repos.Filter{
		"accountName":        ctx.AccountName,
		"metadata.namespace": d.getAccountNamespace(ctx.AccountName),
	}
	return d.secretRepo.FindPaginated(ctx, d.secretRepo.MergeSearchFilter(filter, search), pagination)
}

func (d *domain) GetProviderSecret(ctx InfraContext, name string) (*entities.CloudProviderSecret, error) {
	if err := d.canPerformActionInAccount(ctx, iamT.GetCloudProviderSecret); err != nil {
		return nil, err
	}
	return d.findProviderSecret(ctx, name)
}

func (d *domain) findProviderSecret(ctx InfraContext, name string) (*entities.CloudProviderSecret, error) {
	scrt, err := d.secretRepo.FindOne(ctx, repos.Filter{
		"accountName":        ctx.AccountName,
		"metadata.namespace": d.getAccountNamespace(ctx.AccountName),
		"metadata.name":      name,
	})
	if err != nil {
		return nil, err
	}

	if scrt == nil {
		return nil, fmt.Errorf("provider secret with name %q not found", name)
	}

	return scrt, nil
}
