package domain

import (
	"fmt"
	"time"

	"kloudlite.io/apps/console/internal/domain/entities"
	"kloudlite.io/pkg/repos"
	t "kloudlite.io/pkg/types"
)

func (d *domain) CreateManagedService(ctx ConsoleContext, msvc entities.MSvc) (*entities.MSvc, error) {
	msvc.EnsureGVK()
	if err := d.k8sExtendedClient.ValidateStruct(ctx, &msvc.ManagedService); err != nil {
		return nil, err
	}

	msvc.AccountName = ctx.accountName
	msvc.ClusterName = ctx.clusterName
	msvc.SyncStatus = t.GetSyncStatusForCreation()
	m, err := d.msvcRepo.Create(ctx, &msvc)
	if err != nil {
		if d.msvcRepo.ErrAlreadyExists(err) {
			return nil, fmt.Errorf("msvc with name %q already exists", msvc.Name)
		}
		return nil, err
	}

	if err := d.applyK8sResource(ctx, &m.ManagedService); err != nil {
		return m, err
	}

	return m, nil
}

func (d *domain) DeleteManagedService(ctx ConsoleContext, namespace string, name string) error {
	m, err := d.findMSvc(ctx, namespace, name)
	if err != nil {
		return err
	}

	m.SyncStatus = t.GetSyncStatusForDeletion(m.Generation)
	if _, err := d.msvcRepo.UpdateById(ctx, m.Id, m); err != nil {
		return err
	}

	return d.deleteK8sResource(ctx, &m.ManagedService)
}

func (d *domain) GetManagedService(ctx ConsoleContext, namespace string, name string) (*entities.MSvc, error) {
	return d.findMSvc(ctx, namespace, name)
}

func (d *domain) ListManagedServices(ctx ConsoleContext, namespace string) ([]*entities.MSvc, error) {
	return d.msvcRepo.Find(ctx, repos.Query{Filter: repos.Filter{
		"accountName":        ctx.accountName,
		"clusterName":        ctx.clusterName,
		"metadata.namespace": namespace,
	}})
}

func (d *domain) UpdateManagedService(ctx ConsoleContext, msvc entities.MSvc) (*entities.MSvc, error) {
	msvc.EnsureGVK()
	if err := d.k8sExtendedClient.ValidateStruct(ctx, &msvc.ManagedService); err != nil {
		return nil, err
	}

	m, err := d.findMSvc(ctx, msvc.Namespace, msvc.Name)
	if err != nil {
		return nil, err
	}

	m.Spec = msvc.Spec
	m.SyncStatus = t.GetSyncStatusForUpdation(m.Generation + 1)

	upMSvc, err := d.msvcRepo.UpdateById(ctx, m.Id, m)
	if err != nil {
		return nil, err
	}

	if err := d.applyK8sResource(ctx, &upMSvc.ManagedService); err != nil {
		return upMSvc, err
	}

	return upMSvc, nil
}

func (d *domain) findMSvc(ctx ConsoleContext, namespace string, name string) (*entities.MSvc, error) {
	mres, err := d.msvcRepo.FindOne(ctx, repos.Filter{
		"accountName":        ctx.accountName,
		"clusterName":        ctx.clusterName,
		"metadata.namespace": namespace,
		"metadata.name":      name,
	})
	if err != nil {
		return nil, err
	}
	if mres == nil {
		return nil, fmt.Errorf("no secret with name=%q,namespace=%q found", name, namespace)
	}
	return mres, nil
}

func (d *domain) OnDeleteManagedServiceMessage(ctx ConsoleContext, msvc entities.MSvc) error {
	m, err := d.findMSvc(ctx, msvc.Namespace, msvc.Name)
	if err != nil {
		return err
	}

	return d.msvcRepo.DeleteById(ctx, m.Id)
}

func (d *domain) OnUpdateManagedServiceMessage(ctx ConsoleContext, msvc entities.MSvc) error {
	m, err := d.findMSvc(ctx, msvc.Namespace, msvc.Name)
	if err != nil {
		return err
	}

	m.Status = msvc.Status
	m.SyncStatus.LastSyncedAt = time.Now()
	m.SyncStatus.State = t.ParseSyncState(msvc.Status.IsReady)

	_, err = d.msvcRepo.UpdateById(ctx, m.Id, m)
	return err
}

func (d *domain) OnApplyManagedServiceError(ctx ConsoleContext, err error, namespace, name string) error {
	m, err2 := d.findMSvc(ctx, namespace, name)
	if err2 != nil {
		return err2
	}

	m.SyncStatus.Error = err.Error()
	_, err2 = d.msvcRepo.UpdateById(ctx, m.Id, m)
	return err2
}
