package domain

import (
	"fmt"
	"time"

	"kloudlite.io/apps/console/internal/domain/entities"
	"kloudlite.io/pkg/repos"
	t "kloudlite.io/pkg/types"
)

// query

func (d *domain) ListSecrets(ctx ConsoleContext, namespace string) ([]*entities.Secret, error) {
	if err := d.canReadResourcesInWorkspace(ctx, namespace); err != nil {
		return nil, err
	}

	return d.secretRepo.Find(ctx, repos.Query{Filter: repos.Filter{
		"accountName":        ctx.AccountName,
		"clusterName":        ctx.ClusterName,
		"metadata.namespace": namespace,
	}})
}

func (d *domain) findSecret(ctx ConsoleContext, namespace string, name string) (*entities.Secret, error) {
	scrt, err := d.secretRepo.FindOne(ctx, repos.Filter{
		"accountName":        ctx.AccountName,
		"clusterName":        ctx.ClusterName,
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

func (d *domain) GetSecret(ctx ConsoleContext, namespace string, name string) (*entities.Secret, error) {
	if err := d.canReadResourcesInWorkspace(ctx, namespace); err != nil {
		return nil, err
	}
	return d.findSecret(ctx, namespace, name)
}

// mutations

func (d *domain) CreateSecret(ctx ConsoleContext, secret entities.Secret) (*entities.Secret, error) {
	if err := d.canMutateResourcesInWorkspace(ctx, secret.Namespace); err != nil {
		return nil, err
	}

	secret.EnsureGVK()
	if err := d.k8sExtendedClient.ValidateStruct(ctx, &secret.Secret); err != nil {
		return nil, err
	}

	secret.AccountName = ctx.AccountName
	secret.ClusterName = ctx.ClusterName
	secret.Generation = 1
	secret.SyncStatus = t.GenSyncStatus(t.SyncActionApply, secret.Generation)

	s, err := d.secretRepo.Create(ctx, &secret)
	if err != nil {
		if d.secretRepo.ErrAlreadyExists(err) {
			return nil, fmt.Errorf("secret with name %q, already exists", secret.Name)
		}
		return nil, err
	}

	if err := d.applyK8sResource(ctx, &s.Secret); err != nil {
		return s, err
	}

	return s, nil
}

func (d *domain) UpdateSecret(ctx ConsoleContext, secret entities.Secret) (*entities.Secret, error) {
	if err := d.canMutateResourcesInWorkspace(ctx, secret.Namespace); err != nil {
		return nil, err
	}

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
	s.Generation += 1
	s.SyncStatus = t.GenSyncStatus(t.SyncActionApply, s.Generation)

	upSecret, err := d.secretRepo.UpdateById(ctx, s.Id, s)
	if err != nil {
		return nil, err
	}

	if err := d.applyK8sResource(ctx, &upSecret.Secret); err != nil {
		return upSecret, err
	}

	return upSecret, nil
}

func (d *domain) DeleteSecret(ctx ConsoleContext, namespace string, name string) error {
	if err := d.canMutateResourcesInWorkspace(ctx, namespace); err != nil {
		return err
	}

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
	s.SyncStatus.Error = nil
	s.SyncStatus.LastSyncedAt = time.Now()
	s.SyncStatus.Generation = secret.Generation
	s.SyncStatus.State = t.ParseSyncState(secret.Status.IsReady)

	_, err = d.secretRepo.UpdateById(ctx, s.Id, s)
	return err
}

func (d *domain) OnApplySecretError(ctx ConsoleContext, errMsg, namespace, name string) error {
	s, err2 := d.findSecret(ctx, namespace, name)
	if err2 != nil {
		return err2
	}

	s.SyncStatus.Error = &errMsg
	_, err := d.secretRepo.UpdateById(ctx, s.Id, s)
	return err
}

func (d *domain) ResyncSecret(ctx ConsoleContext, namespace, name string) error {
	if err := d.canMutateResourcesInWorkspace(ctx, namespace); err != nil {
		return err
	}

	s, err := d.findSecret(ctx, namespace, name)
	if err != nil {
		return err
	}

	return d.resyncK8sResource(ctx, s.SyncStatus.Action, &s.Secret)
}
