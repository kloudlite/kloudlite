package domain

import (
	"fmt"
	"time"

	"kloudlite.io/apps/console/internal/domain/entities"
	"kloudlite.io/pkg/repos"
	t "kloudlite.io/pkg/types"
)

func (d *domain) CreateSecret(ctx ConsoleContext, secret entities.Secret) (*entities.Secret, error) {
	secret.EnsureGVK()
	if err := d.k8sExtendedClient.ValidateStruct(ctx, &secret.Secret); err != nil {
		return nil, err
	}

	secret.AccountName = ctx.accountName
	secret.ClusterName = ctx.clusterName
	secret.SyncStatus = t.GetSyncStatusForCreation()
	s, err := d.secretRepo.Create(ctx, &secret)
	if err != nil {
		if d.secretRepo.ErrAlreadyExists(err) {
			return nil, fmt.Errorf("secret with name '%s' already exists", secret.Name)
		}
		return nil, err
	}

	if err := d.applyK8sResource(ctx, &s.Secret); err != nil {
		return s, err
	}

	return s, nil
}

func (d *domain) DeleteSecret(ctx ConsoleContext, namespace string, name string) error {
	s, err := d.findSecret(ctx, namespace, name)
	if err != nil {
		return err
	}
	s.SyncStatus = t.GetSyncStatusForDeletion(s.Generation)
	if _, err := d.secretRepo.UpdateById(ctx, s.Id, s); err != nil {
		return err
	}

	return d.deleteK8sResource(ctx, &s.Secret)
}

func (d *domain) GetSecret(ctx ConsoleContext, namespace string, name string) (*entities.Secret, error) {
	return d.findSecret(ctx, namespace, name)
}

func (d *domain) ListSecrets(ctx ConsoleContext, namespace string) ([]*entities.Secret, error) {
	return d.secretRepo.Find(ctx, repos.Query{Filter: repos.Filter{
		"accountName":        ctx.accountName,
		"clusterName":        ctx.clusterName,
		"metadata.namespace": namespace,
	}})
}

func (d *domain) UpdateSecret(ctx ConsoleContext, secret entities.Secret) (*entities.Secret, error) {
	secret.EnsureGVK()
	if err := d.k8sExtendedClient.ValidateStruct(ctx, &secret.Secret); err != nil {
		return nil, err
	}

	s, err := d.findSecret(ctx, secret.Namespace, secret.Name)
	if err != nil {
		return nil, err
	}

	s.Data = secret.Data
	s.StringData = secret.StringData
	s.SyncStatus = t.GetSyncStatusForUpdation(s.Generation + 1)

	upSecret, err := d.secretRepo.UpdateById(ctx, s.Id, s)
	if err != nil {
		return nil, err
	}

	if err := d.applyK8sResource(ctx, &upSecret.Secret); err != nil {
		return upSecret, err
	}

	return upSecret, nil
}

func (d *domain) findSecret(ctx ConsoleContext, namespace string, name string) (*entities.Secret, error) {
	scrt, err := d.secretRepo.FindOne(ctx, repos.Filter{
		"accountName":        ctx.accountName,
		"clusterName":        ctx.clusterName,
		"metadata.namespace": namespace,
		"metadata.name":      name,
	})
	if err != nil {
		return nil, err
	}
	if scrt == nil {
		return nil, fmt.Errorf("no secret with name=%s,namespace=%s found", name, namespace)
	}
	return scrt, nil
}

func (d *domain) OnDeleteSecretMessage(ctx ConsoleContext, secret entities.Secret) error {
	s, err := d.findSecret(ctx, secret.Namespace, secret.Name)
	if err != nil {
		return err
	}

	return d.secretRepo.DeleteById(ctx, s.Id)
}

func (d *domain) OnUpdateSecretMessage(ctx ConsoleContext, secret entities.Secret) error {
	s, err := d.findSecret(ctx, secret.Namespace, secret.Name)
	if err != nil {
		return err
	}

	s.Status = secret.Status
	s.SyncStatus.LastSyncedAt = time.Now()
	s.SyncStatus.State = t.ParseSyncState(secret.Status.IsReady)

	_, err = d.secretRepo.UpdateById(ctx, s.Id, s)
	return err
}

func (d *domain) OnApplySecretError(ctx ConsoleContext, err error, namespace, name string) error {
	s, err2 := d.findSecret(ctx, namespace, name)
	if err2 != nil {
		return err2
	}

	s.SyncStatus.Error = err.Error()
	_, err = d.secretRepo.UpdateById(ctx, s.Id, s)
	return err
}
