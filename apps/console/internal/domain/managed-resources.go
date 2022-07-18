package domain

import (
	"context"
	"fmt"
	"kloudlite.io/apps/console/internal/domain/entities"
	op_crds "kloudlite.io/apps/console/internal/domain/op-crds"
	"kloudlite.io/pkg/repos"
)

func (d *domain) GetManagedRes(ctx context.Context, managedResID repos.ID) (*entities.ManagedResource, error) {
	return d.managedResRepo.FindById(ctx, managedResID)
}
func (d *domain) GetManagedResources(ctx context.Context, projectID repos.ID) ([]*entities.ManagedResource, error) {
	return d.managedResRepo.Find(ctx, repos.Query{Filter: repos.Filter{
		"project_id": projectID,
	}})
}
func (d *domain) GetManagedResourcesOfService(ctx context.Context, installationId repos.ID) ([]*entities.ManagedResource, error) {
	fmt.Println("GetManagedResourcesOfService", installationId)
	return d.managedResRepo.Find(ctx, repos.Query{Filter: repos.Filter{
		"service_id": installationId,
	}})
}

func (d *domain) GetResourceOutputs(ctx context.Context, managedResID repos.ID) (map[string]string, error) {
	mRes, err := d.managedResRepo.FindById(ctx, managedResID)
	if err != nil {
		return nil, err
	}
	project, err := d.projectRepo.FindById(ctx, mRes.ProjectId)
	if err != nil {
		return nil, err
	}
	_, err = d.clusterRepo.FindOne(ctx, repos.Filter{
		"account_id": project.AccountId,
	})
	if err != nil {
		return nil, err
	}
	return nil, nil
}

func (d *domain) OnUpdateManagedRes(ctx context.Context, response *op_crds.StatusUpdate) error {
	one, err := d.managedResRepo.FindOne(ctx, repos.Filter{
		"id": response.Metadata.ResourceId,
	})
	if err != nil {
		return err
	}
	if response.IsReady {
		one.Status = entities.ManagedResourceStateLive
	} else {
		one.Status = entities.ManagedResourceStateSyncing
	}
	one.Conditions = response.ChildConditions
	_, err = d.managedResRepo.UpdateById(ctx, one.Id, one)
	err = d.notifier.Notify(one.Id)
	if err != nil {
		return err
	}
	return err
}

func (d *domain) InstallManagedRes(ctx context.Context, installationId repos.ID, name string, resourceType string, values map[string]string) (*entities.ManagedResource, error) {
	svc, err := d.managedSvcRepo.FindById(ctx, installationId)
	if err != nil {
		return nil, err
	}
	if svc == nil {
		return nil, fmt.Errorf("managed service not found")
	}
	prj, err := d.projectRepo.FindById(ctx, svc.ProjectId)
	if err != nil {
		return nil, err
	}
	if prj == nil {
		return nil, fmt.Errorf("project not found")
	}

	create, err := d.managedResRepo.Create(ctx, &entities.ManagedResource{
		ProjectId:    prj.Id,
		Namespace:    prj.Name,
		ServiceId:    svc.Id,
		ResourceType: entities.ManagedResourceType(resourceType),
		Name:         name,
		Values:       values,
	})
	if err != nil {
		return nil, err
	}

	template, err := d.GetManagedServiceTemplate(ctx, string(svc.ServiceType))
	var resTmpl entities.ManagedResourceTemplate
	for _, rt := range template.Resources {
		if rt.Name == resourceType {
			resTmpl = rt
			break
		}
	}

	err = d.workloadMessenger.SendAction("apply", string(create.Id), &op_crds.ManagedResource{
		APIVersion: op_crds.ManagedResourceAPIVersion,
		Kind:       op_crds.ManagedResourceKind,
		Metadata: op_crds.ManagedResourceMetadata{
			Name:      string(create.Id),
			Namespace: create.Namespace,
		},
		Spec: op_crds.ManagedResourceSpec{
			ManagedServiceName: string(svc.Id),
			ApiVersion:         resTmpl.ApiVersion,
			Kind:               resTmpl.Kind,
			Inputs:             create.Values,
		},
		Status: op_crds.Status{},
	})
	if err != nil {
		return nil, err
	}

	return create, nil
}
func (d *domain) UpdateManagedRes(ctx context.Context, managedResID repos.ID, values map[string]string) (bool, error) {
	id, err := d.managedResRepo.FindById(ctx, managedResID)
	if err != nil {
		return false, err
	}
	id.Values = values
	_, err = d.managedResRepo.UpdateById(ctx, managedResID, id)
	if err != nil {
		return false, err
	}
	return true, nil
}
func (d *domain) UnInstallManagedRes(ctx context.Context, appID repos.ID) (bool, error) {
	err := d.managedResRepo.DeleteById(ctx, appID)
	if err != nil {
		return false, err
	}
	return true, err
}
