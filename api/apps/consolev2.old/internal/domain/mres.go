package domain

import (
	"context"
	"kloudlite.io/apps/consolev2.old/internal/domain/entities"
	"kloudlite.io/pkg/repos"
)

func (d *domain) GetManagedRes(ctx context.Context, namespace string, name string) (*entities.ManagedResource, error) {
	return d.managedResRepo.FindOne(ctx, repos.Filter{"metadata.namespace": namespace, "metadata.name": name})
}

func (d *domain) GetManagedResOutput(ctx context.Context, namespace string, name string) (map[string]any, error) {
	//TODO implement me
	panic("implement me")
}

func (d *domain) GetManagedResources(ctx context.Context, namespace string) ([]*entities.ManagedResource, error) {
	return d.managedResRepo.Find(ctx, repos.Query{Filter: repos.Filter{"metadata.namespace": namespace}})
}

func (d *domain) GetManagedResourcesOfService(ctx context.Context, msvcNamespace string, msvcName string) ([]*entities.ManagedResource, error) {
	return d.managedResRepo.Find(ctx, repos.Query{Filter: repos.Filter{"metadata.namespace": msvcNamespace, "spec.msvcRef.name": msvcName}})
}

func (d *domain) upsertMres(ctx context.Context, mres entities.ManagedResource) (*entities.ManagedResource, error) {
	nMres, err := d.managedResRepo.Upsert(ctx, repos.Filter{"metadata.namespace": mres.Namespace, "metadata.name": mres.Name}, &mres)
	if err != nil {
		return nil, err
	}
	clusterId, err := d.getClusterForProject(ctx, mres.Spec.ProjectName)
	if err != nil {
		return nil, err
	}

	if err := d.workloadMessenger.SendAction(ActionApply, d.getDispatchKafkaTopic(clusterId), string(nMres.Id), nMres.ManagedResource); err != nil {
		return nil, err
	}
	return nMres, nil
}

func (d *domain) CreateManagedRes(ctx context.Context, mres entities.ManagedResource) (*entities.ManagedResource, error) {
	return d.upsertMres(ctx, mres)
}

func (d *domain) UpdateManagedRes(ctx context.Context, mres entities.ManagedResource) (*entities.ManagedResource, error) {
	uMres, err := d.upsertMres(ctx, mres)
	if err != nil {
		return nil, err
	}
	return uMres, nil
}

func (d *domain) DeleteManagedRes(ctx context.Context, namespace string, name string) (bool, error) {
	if err := d.managedResRepo.DeleteOne(ctx, repos.Filter{"metadata.namespace": namespace, "metadata.name": name}); err != nil {
		return false, err
	}
	return true, nil
}
