package domain

import (
	"fmt"
	"time"

	"kloudlite.io/apps/console/internal/domain/entities"
	"kloudlite.io/pkg/repos"
	t "kloudlite.io/pkg/types"
)

// query

func (d *domain) ListManagedResources(ctx ConsoleContext, namespace string) ([]*entities.MRes, error) {
	if err := d.canReadResourcesInWorkspace(ctx, namespace); err != nil {
		return nil, err
	}
	return d.mresRepo.Find(ctx, repos.Query{Filter: repos.Filter{
		"accountName":        ctx.AccountName,
		"clusterName":        ctx.ClusterName,
		"metadata.namespace": namespace,
	}})
}

func (d *domain) findMRes(ctx ConsoleContext, namespace string, name string) (*entities.MRes, error) {
	mres, err := d.mresRepo.FindOne(ctx, repos.Filter{
		"accountName":        ctx.AccountName,
		"clusterName":        ctx.ClusterName,
		"metadata.namespace": namespace,
		"metadata.name":      name,
	})
	if err != nil {
		return nil, err
	}
	if mres == nil {
		return nil, fmt.Errorf(
			"no managed resource with name=%q,namespace=%q found",
			name,
			namespace,
		)
	}
	return mres, nil
}

func (d *domain) GetManagedResource(ctx ConsoleContext, namespace string, name string) (*entities.MRes, error) {
	if err := d.canReadResourcesInWorkspace(ctx, namespace); err != nil {
		return nil, err
	}

	return d.findMRes(ctx, namespace, name)
}

// mutations

func (d *domain) CreateManagedResource(ctx ConsoleContext, mres entities.MRes) (*entities.MRes, error) {
	if err := d.canMutateResourcesInWorkspace(ctx, mres.Namespace); err != nil {
		return nil, err
	}

	mres.EnsureGVK()
	if err := d.k8sExtendedClient.ValidateStruct(ctx, &mres.ManagedResource); err != nil {
		return nil, err
	}

	mres.AccountName = ctx.AccountName
	mres.ClusterName = ctx.ClusterName
	mres.Generation = 1
	mres.SyncStatus = t.GenSyncStatus(t.SyncActionApply, mres.Generation)

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

func (d *domain) UpdateManagedResource(ctx ConsoleContext, mres entities.MRes) (*entities.MRes, error) {
	if err := d.canReadResourcesInWorkspace(ctx, mres.Namespace); err != nil {
		return nil, err
	}

	mres.EnsureGVK()
	if err := d.k8sExtendedClient.ValidateStruct(ctx, &mres.ManagedResource); err != nil {
		return nil, err
	}

	m, err := d.findMRes(ctx, mres.Namespace, mres.Name)
	if err != nil {
		return nil, err
	}

	m.Spec = mres.Spec
	m.Generation += 1
	m.SyncStatus = t.GenSyncStatus(t.SyncActionApply, m.Generation)

	upMRes, err := d.mresRepo.UpdateById(ctx, m.Id, m)
	if err != nil {
		return nil, err
	}

	if err := d.applyK8sResource(ctx, &upMRes.ManagedResource); err != nil {
		return upMRes, err
	}

	return upMRes, nil
}

func (d *domain) DeleteManagedResource(ctx ConsoleContext, namespace string, name string) error {
	if err := d.canMutateResourcesInWorkspace(ctx, namespace); err != nil {
		return err
	}

	m, err := d.findMRes(ctx, namespace, name)
	if err != nil {
		return err
	}

	m.SyncStatus = t.GenSyncStatus(t.SyncActionDelete, m.Generation)
	if _, err := d.mresRepo.UpdateById(ctx, m.Id, m); err != nil {
		return err
	}

	return d.deleteK8sResource(ctx, &m.ManagedResource)
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
	m.SyncStatus.Error = nil
	m.SyncStatus.LastSyncedAt = time.Now()
	m.SyncStatus.Generation = mres.Generation
	m.SyncStatus.State = t.ParseSyncState(mres.Status.IsReady)

	_, err = d.mresRepo.UpdateById(ctx, m.Id, m)
	return err
}

func (d *domain) OnApplyManagedResourceError(ctx ConsoleContext, errMsg string, namespace string, name string) error {
	m, err2 := d.findMRes(ctx, namespace, name)
	if err2 != nil {
		return err2
	}

	m.SyncStatus.Error = &errMsg
	_, err := d.mresRepo.UpdateById(ctx, m.Id, m)
	return err
}

func (d *domain) ResyncManagedResource(ctx ConsoleContext, namespace, name string) error {
	if err := d.canMutateResourcesInWorkspace(ctx, namespace); err != nil {
		return err
	}

	m, err := d.findMRes(ctx, namespace, name)
	if err != nil {
		return err
	}
	return d.resyncK8sResource(ctx, m.SyncStatus.Action, &m.ManagedResource)
}
