package domain

import (
	"context"
	"fmt"

	"kloudlite.io/apps/console/internal/domain/entities"
	"kloudlite.io/pkg/repos"
)

func (d *domain) CreateManagedResource(ctx context.Context, mres entities.MRes) (*entities.MRes, error) {
	mres.EnsureGVK()
	if err := d.k8sExtendedClient.ValidateStruct(ctx, &mres.ManagedResource); err != nil {
		return nil, err
	}

	m, err := d.mresRepo.Create(ctx, &mres)
	if err != nil {
		return nil, err
	}

	if err := d.applyK8sResource(ctx, &m.ManagedResource); err != nil {
		return m, err
	}

	return m, nil
}

func (d *domain) DeleteManagedResource(ctx context.Context, namespace string, name string) error {
	m, err := d.findMRes(ctx, namespace, name)
	if err != nil {
		return err
	}
	return d.k8sYamlClient.DeleteResource(ctx, &m.ManagedResource)
}

func (d *domain) GetManagedResource(ctx context.Context, namespace string, name string) (*entities.MRes, error) {
	return d.findMRes(ctx, namespace, name)
}

func (d *domain) GetManagedResources(ctx context.Context, namespace string) ([]*entities.MRes, error) {
	return d.mresRepo.Find(ctx, repos.Query{Filter: repos.Filter{"metadata.namespace": namespace}})
}

func (d *domain) UpdateManagedResource(ctx context.Context, mres entities.MRes) (*entities.MRes, error) {
	mres.EnsureGVK()
	if err := d.k8sExtendedClient.ValidateStruct(ctx, &mres.ManagedResource); err != nil {
		return nil, err
	}

	s, err := d.findMRes(ctx, mres.Namespace, mres.Name)
	if err != nil {
		return nil, err
	}

	status := s.Status
	s.ManagedResource = mres.ManagedResource
	s.Status = status

	upMRes, err := d.mresRepo.UpdateById(ctx, s.Id, s)
	if err != nil {
		return nil, err
	}

	if err := d.applyK8sResource(ctx, &upMRes.ManagedResource); err != nil {
		return upMRes, err
	}

	return upMRes, nil
}

func (d *domain) findMRes(ctx context.Context, namespace string, name string) (*entities.MRes, error) {
	mres, err := d.mresRepo.FindOne(ctx, repos.Filter{"metadata.namespace": namespace, "metadata.name": name})
	if err != nil {
		return nil, err
	}
	if mres == nil {
		return nil, fmt.Errorf("no managed resource with name=%s,namespace=%s found", name, namespace)
	}
	return mres, nil
}
