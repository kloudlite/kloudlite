package domain

import (
	"context"
	"fmt"

	"kloudlite.io/apps/console/internal/domain/entities"
	"kloudlite.io/pkg/repos"
)

func (d *domain) CreateManagedService(ctx context.Context, msvc entities.MSvc) (*entities.MSvc, error) {
	msvc.EnsureGVK()
	if err := d.k8sExtendedClient.ValidateStruct(ctx, &msvc.ManagedService); err != nil {
		return nil, err
	}

	m, err := d.msvcRepo.Create(ctx, &msvc)
	if err != nil {
		return nil, err
	}

	if err := d.applyK8sResource(ctx, &m.ManagedService); err != nil {
		return m, err
	}

	return m, nil
}

func (d *domain) DeleteManagedService(ctx context.Context, namespace string, name string) error {
	m, err := d.findMSvc(ctx, namespace, name)
	if err != nil {
		return err
	}
	return d.k8sYamlClient.DeleteResource(ctx, &m.ManagedService)
}

func (d *domain) GetManagedService(ctx context.Context, namespace string, name string) (*entities.MSvc, error) {
	return d.findMSvc(ctx, namespace, name)
}

func (d *domain) GetManagedServices(ctx context.Context, namespace string) ([]*entities.MSvc, error) {
	return d.msvcRepo.Find(ctx, repos.Query{Filter: repos.Filter{"metadata.namespace": namespace}})
}

func (d *domain) UpdateManagedService(ctx context.Context, msvc entities.MSvc) (*entities.MSvc, error) {
	msvc.EnsureGVK()
	if err := d.k8sExtendedClient.ValidateStruct(ctx, &msvc.ManagedService); err != nil {
		return nil, err
	}

	s, err := d.findMSvc(ctx, msvc.Namespace, msvc.Name)
	if err != nil {
		return nil, err
	}

	status := s.Status
	s.ManagedService = msvc.ManagedService
	s.Status = status

	upMSvc, err := d.msvcRepo.UpdateById(ctx, s.Id, s)
	if err != nil {
		return nil, err
	}

	if err := d.applyK8sResource(ctx, &upMSvc.ManagedService); err != nil {
		return upMSvc, err
	}

	return upMSvc, nil
}

func (d *domain) findMSvc(ctx context.Context, namespace string, name string) (*entities.MSvc, error) {
	mres, err := d.msvcRepo.FindOne(ctx, repos.Filter{"metadata.namespace": namespace, "metadata.name": name})
	if err != nil {
		return nil, err
	}
	if mres == nil {
		return nil, fmt.Errorf("no secret with name=%s,namespace=%s found", name, namespace)
	}
	return mres, nil
}
