package domain

import (
	"context"
	"kloudlite.io/apps/consolev2/internal/domain/entities"
	"kloudlite.io/pkg/errors"
	"kloudlite.io/pkg/repos"
	"os"
	"sigs.k8s.io/yaml"
)

func (d *domain) GetManagedSvc(ctx context.Context, namespace string, name string) (*entities.ManagedService, error) {
	return d.managedSvcRepo.FindOne(ctx, repos.Filter{"metadata.namespace": namespace, "metadata.name": name})
}

func (d *domain) GetManagedServiceTemplates(ctx context.Context) ([]*entities.ManagedServiceCategory, error) {
	if _, err := GetUser(ctx); err != nil {
		return nil, err
	}

	templates := make([]*entities.ManagedServiceCategory, 0)
	data, err := os.ReadFile(d.managedTemplatesPath)
	if err != nil {
		return nil, err
	}
	err = yaml.Unmarshal(data, &templates)
	if err != nil {
		return nil, err
	}
	return templates, nil
}

func (d *domain) GetManagedServiceTemplate(ctx context.Context, name string) (*entities.ManagedServiceTemplate, error) {
	if _, err := GetUser(ctx); err != nil {
		return nil, err
	}

	templates := make([]*entities.ManagedServiceCategory, 0)
	data, err := os.ReadFile(d.managedTemplatesPath)
	if err != nil {
		return nil, err
	}
	err = yaml.Unmarshal(data, &templates)
	if err != nil {
		return nil, err
	}
	for _, t := range templates {
		for _, s := range t.List {
			if s.Name == name {
				return s, nil
			}
		}
	}
	return nil, errors.New("not found")
}

func (d *domain) upsertMsvc(ctx context.Context, msvc entities.ManagedService) (*entities.ManagedService, error) {
	nMsvc, err := d.managedSvcRepo.Upsert(ctx, repos.Filter{"metadata.namespace": msvc.Namespace, "metadata.name": msvc.Name}, &msvc)
	if err != nil {
		return nil, err
	}
	clusterId, err := d.getClusterForProject(ctx, msvc.Spec.ProjectName)
	if err != nil {
		return nil, err
	}

	if err := d.workloadMessenger.SendAction(ActionApply, d.getDispatchKafkaTopic(clusterId), string(nMsvc.Id), nMsvc.ManagedService); err != nil {
		return nil, err
	}
	return nMsvc, nil
}

func (d *domain) InstallManagedSvc(ctx context.Context, msvc entities.ManagedService) (*entities.ManagedService, error) {
	return d.upsertMsvc(ctx, msvc)
}

func (d *domain) UnInstallManagedSvc(ctx context.Context, namespace, name string) (bool, error) {
	if err := d.managedSvcRepo.DeleteOne(ctx, repos.Filter{"metadata.namespace": namespace, "metadata.name": name}); err != nil {
		return false, err
	}
	return true, nil
}

func (d *domain) UpdateManagedSvc(ctx context.Context, msvc entities.ManagedService) (*entities.ManagedService, error) {
	uMsvc, err := d.upsertMsvc(ctx, msvc)
	if err != nil {
		return nil, err
	}
	return uMsvc, nil
}

func (d *domain) GetManagedSvcOutput(ctx context.Context, namespace, name string) (map[string]any, error) {
	//TODO implement me
	panic("implement me")
}

func (d *domain) GetManagedSvcs(ctx context.Context, namespace string) ([]*entities.ManagedService, error) {
	return d.managedSvcRepo.Find(ctx, repos.Query{Filter: repos.Filter{"metadata.namespace": namespace}})
}
