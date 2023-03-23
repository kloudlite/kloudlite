package domain

import (
	"fmt"
	"time"

	"kloudlite.io/apps/console/internal/domain/entities"
	"kloudlite.io/pkg/repos"
	t "kloudlite.io/pkg/types"
)

func (d *domain) CreateManagedResource(ctx ConsoleContext, mres entities.MRes) (*entities.MRes, error) {
	mres.EnsureGVK()
	if err := d.k8sExtendedClient.ValidateStruct(ctx, &mres.ManagedResource); err != nil {
		return nil, err
	}

	mres.AccountName = ctx.accountName
	mres.ClusterName = ctx.clusterName
	mres.SyncStatus = t.GetSyncStatusForCreation()
	m, err := d.mresRepo.Create(ctx, &mres)
	if err != nil {
		if d.mresRepo.ErrAlreadyExists(err) {
			return nil, fmt.Errorf("mres with name %q already exists", mres.Name)
		}
		return nil, err
	}

	if err := d.applyK8sResource(ctx, &m.ManagedResource); err != nil {
		return m, err
	}

	return m, nil
}

func (d *domain) DeleteManagedResource(ctx ConsoleContext, namespace string, name string) error {
	m, err := d.findMRes(ctx, namespace, name)
	if err != nil {
		return err
	}

	m.SyncStatus = t.GetSyncStatusForDeletion(m.Generation)
	if _, err := d.mresRepo.UpdateById(ctx, m.Id, m); err != nil {
		return err
	}

	return d.deleteK8sResource(ctx, &m.ManagedResource)
}

func (d *domain) GetManagedResource(ctx ConsoleContext, namespace string, name string) (*entities.MRes, error) {
	return d.findMRes(ctx, namespace, name)
}

func (d *domain) ListManagedResources(ctx ConsoleContext, namespace string) ([]*entities.MRes, error) {
	return d.mresRepo.Find(ctx, repos.Query{Filter: repos.Filter{
		"accountName":        ctx.accountName,
		"clusterName":        ctx.clusterName,
		"metadata.namespace": namespace,
	}})
}

func (d *domain) UpdateManagedResource(ctx ConsoleContext, mres entities.MRes) (*entities.MRes, error) {
	mres.EnsureGVK()
	if err := d.k8sExtendedClient.ValidateStruct(ctx, &mres.ManagedResource); err != nil {
		return nil, err
	}

	m, err := d.findMRes(ctx, mres.Namespace, mres.Name)
	if err != nil {
		return nil, err
	}

	m.Spec = mres.Spec
	m.SyncStatus = t.GetSyncStatusForUpdation(m.Generation + 1)

	upMRes, err := d.mresRepo.UpdateById(ctx, m.Id, m)
	if err != nil {
		return nil, err
	}

	if err := d.applyK8sResource(ctx, &upMRes.ManagedResource); err != nil {
		return upMRes, err
	}

	return upMRes, nil
}

func (d *domain) findMRes(ctx ConsoleContext, namespace string, name string) (*entities.MRes, error) {
	mres, err := d.mresRepo.FindOne(ctx, repos.Filter{
		"accountName":        ctx.accountName,
		"clusterName":        ctx.clusterName,
		"metadata.namespace": namespace,
		"metadata.name":      name,
	})
	if err != nil {
		return nil, err
	}
	if mres == nil {
		return nil, fmt.Errorf("no managed resource with name=%q,namespace=%q found", name, namespace)
	}
	return mres, nil
}

func (d *domain) OnDeleteManagedResourceMessage(ctx ConsoleContext, mres entities.MRes) error {
	a, err := d.findMRes(ctx, mres.Namespace, mres.Name)
	if err != nil {
		return err
	}

	return d.mresRepo.DeleteById(ctx, a.Id)
}

func (d *domain) OnUpdateManagedResourceMessage(ctx ConsoleContext, mres entities.MRes) error {
	m, err := d.findMRes(ctx, mres.Namespace, mres.Name)
	if err != nil {
		return err
	}

	m.Status = mres.Status
	m.SyncStatus.LastSyncedAt = time.Now()
	m.SyncStatus.State = t.ParseSyncState(mres.Status.IsReady)

	_, err = d.mresRepo.UpdateById(ctx, m.Id, m)
	return err
}

func (d *domain) OnApplyManagedResourceError(ctx ConsoleContext, err error, namespace, name string) error {
	m, err2 := d.findMRes(ctx, namespace, name)
	if err2 != nil {
		return err2
	}

	m.SyncStatus.Error = err.Error()
	_, err = d.mresRepo.UpdateById(ctx, m.Id, m)
	return err
}
