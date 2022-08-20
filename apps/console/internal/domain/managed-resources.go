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

func (d *domain) OnUpdateManagedRes(ctx context.Context, response *op_crds.StatusUpdate) error {
	one, err := d.managedResRepo.FindOne(ctx, repos.Filter{
		"id": response.Metadata.ResourceId,
	})
	if err != nil {
		return err
	}
	if one == nil {
		return fmt.Errorf("managed resource not found")
	}
	newStatus := one.Status
	if response.IsReady {
		newStatus = entities.ManagedResourceStateLive
	}
	shouldUpdate := newStatus != one.Status
	one.Conditions = response.ChildConditions
	one.Status = newStatus
	_, err = d.managedResRepo.UpdateById(ctx, one.Id, one)
	if shouldUpdate {
		err = d.notifier.Notify(one.Id)
		if err != nil {
			return err
		}
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
			MsvcRef: op_crds.MsvcRef{
				APIVersion: resTmpl.ApiVersion,
				Kind:       "Service",
				Name:       string(svc.Id),
			},
			MresKind: op_crds.MresKind{
				Kind: resTmpl.Kind,
			},
			Inputs: create.Values,
		},
	})
	if err != nil {
		return nil, err
	}

	return create, nil
}
func (d *domain) UpdateManagedRes(ctx context.Context, managedResID repos.ID, values map[string]string) (bool, error) {
	mres, err := d.managedResRepo.FindById(ctx, managedResID)
	if err != nil {
		return false, err
	}
	mres.Values = values
	_, err = d.managedResRepo.UpdateById(ctx, managedResID, mres)
	if err != nil {
		return false, err
	}
	err = d.workloadMessenger.SendAction("apply", string(mres.Id), &op_crds.ManagedResource{
		APIVersion: op_crds.ManagedResourceAPIVersion,
		Kind:       op_crds.ManagedResourceKind,
		Metadata: op_crds.ManagedResourceMetadata{
			Name:      string(mres.Id),
			Namespace: mres.Namespace,
		},
		Spec: op_crds.ManagedResourceSpec{
			MsvcRef: op_crds.MsvcRef{
				APIVersion: op_crds.ManagedResourceAPIVersion,
				Kind:       "Service",
				Name:       string(mres.ServiceId),
			},
			MresKind: op_crds.MresKind{
				Kind: string(mres.ResourceType),
			},
			Inputs: mres.Values,
		},
	})
	if err != nil {
		return false, err
	}
	return true, nil
}
func (d *domain) UnInstallManagedRes(ctx context.Context, appID repos.ID) (bool, error) {
	id, err := d.managedResRepo.FindById(ctx, appID)
	if err != nil {
		return false, err
	}
	err = d.managedResRepo.DeleteById(ctx, appID)
	if err != nil {
		return false, err
	}
	err = d.workloadMessenger.SendAction("apply", string(appID), &op_crds.ManagedResource{
		APIVersion: op_crds.ManagedResourceAPIVersion,
		Kind:       op_crds.ManagedResourceKind,
		Metadata: op_crds.ManagedResourceMetadata{
			Name:      string(appID),
			Namespace: id.Namespace,
		},
	})
	if err != nil {
		return false, err
	}
	return true, err
}

func (d *domain) getManagedResOutput(ctx context.Context, managedResID repos.ID) (map[string]any, error) {
	mres, err := d.managedResRepo.FindById(ctx, managedResID)
	if err != nil {
		return nil, err
	}
	secret, err := d.kubeCli.GetSecret(ctx, mres.Namespace, fmt.Sprint("mres-", mres.Id))
	if err != nil {
		return nil, err
	}
	parsedSec := make(map[string]any)
	for k, v := range secret.Data {
		parsedSec[k] = string(v)
	}
	return parsedSec, nil
}

func (d *domain) GetManagedResOutput(ctx context.Context, managedResID repos.ID) (map[string]any, error) {
	return d.getManagedResOutput(ctx, managedResID)
}

func (d *domain) OnDeleteManagedResource(ctx context.Context, response *op_crds.StatusUpdate) error {
	return d.managedResRepo.DeleteById(ctx, repos.ID(response.Metadata.ResourceId))
}
