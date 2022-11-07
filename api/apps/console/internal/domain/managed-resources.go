package domain

import (
	"context"
	"fmt"
	"strings"

	"kloudlite.io/apps/console/internal/domain/entities"
	opCrds "kloudlite.io/apps/console/internal/domain/op-crds"
	"kloudlite.io/pkg/kubeapi"
	"kloudlite.io/pkg/repos"
)

func (d *domain) GetManagedRes(ctx context.Context, managedResID repos.ID) (*entities.ManagedResource, error) {
	if strings.HasPrefix(string(managedResID), "mgsvc-") {
		msvc, err := d.managedSvcRepo.FindById(ctx, managedResID)
		if err = mongoError(err, "resource not found"); err != nil {
			return nil, err
		}
		err = d.checkProjectAccess(ctx, msvc.ProjectId, ReadProject)
		if err != nil {
			return nil, err
		}

		return &entities.ManagedResource{
			BaseEntity: repos.BaseEntity{
				Id:           msvc.Id,
				CreationTime: msvc.CreationTime,
				UpdateTime:   msvc.UpdateTime,
			},
			ClusterId: msvc.ClusterId,
			ProjectId: msvc.ProjectId,
			Name:      msvc.Name,
			Namespace: msvc.Namespace,
			ServiceId: msvc.Id,
			// Values:     msvc.Values,
			// Status:     msvc.Status,
			Conditions: msvc.Conditions,
		}, nil
	}
	mr, err := d.managedResRepo.FindById(ctx, managedResID)
	if err = mongoError(err, "resource not found"); err != nil {
		return nil, err
	}

	err = d.checkProjectAccess(ctx, mr.ProjectId, ReadProject)
	if err != nil {
		return nil, err
	}

	return mr, nil
}

func (d *domain) GetManagedResources(ctx context.Context, projectID repos.ID) ([]*entities.ManagedResource, error) {
	err := d.checkProjectAccess(ctx, projectID, ReadProject)
	if err != nil {
		return nil, err
	}

	return d.managedResRepo.Find(
		ctx, repos.Query{Filter: repos.Filter{
			"project_id": projectID,
		}},
	)
}

func (d *domain) GetManagedResourcesOfService(ctx context.Context, installationId repos.ID) ([]*entities.ManagedResource, error) {
	mres, err := d.managedResRepo.Find(
		ctx, repos.Query{Filter: repos.Filter{
			"service_id": installationId,
		}},
	)

	if err != nil {
		return nil, err
	}

	if len(mres) > 0 {
		err = d.checkProjectAccess(ctx, mres[0].ProjectId, ReadProject)
		if err != nil {
			return nil, err
		}
	}

	return mres, nil
}

func (d *domain) OnUpdateManagedRes(ctx context.Context, response *opCrds.StatusUpdate) error {
	one, err := d.managedResRepo.FindById(ctx, repos.ID(response.Metadata.ResourceId))
	if err = mongoError(err, "managed resource not found"); err != nil {
		// Ignore unknown resource
		return nil
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
	if err = mongoError(err, "service not found"); err != nil {
		return nil, err
	}

	err = d.checkProjectAccess(ctx, svc.ProjectId, UpdateProject)
	if err != nil {
		return nil, err
	}

	prj, err := d.projectRepo.FindById(ctx, svc.ProjectId)
	if err = mongoError(err, "project not found"); err != nil {
		return nil, err
	}

	create, err := d.managedResRepo.Create(
		ctx, &entities.ManagedResource{
			ProjectId:    prj.Id,
			Namespace:    prj.Name,
			ServiceId:    svc.Id,
			ResourceType: entities.ManagedResourceType(resourceType),
			Name:         name,
			Values:       values,
		},
	)
	if err != nil {
		return nil, err
	}

	template, err := d.GetManagedServiceTemplate(ctx, string(svc.ServiceType))
	if err != nil {
		return nil, err
	}
	var resTmpl entities.ManagedResourceTemplate
	for _, rt := range template.Resources {
		if rt.Name == resourceType {
			resTmpl = rt
			break
		}
	}

	clusterId, err := d.getClusterForAccount(ctx, prj.AccountId)
	if err != nil {
		return nil, err
	}

	err = d.workloadMessenger.SendAction(
		"apply", d.getDispatchKafkaTopic(clusterId), string(create.Id), &opCrds.ManagedResource{
			APIVersion: opCrds.ManagedResourceAPIVersion,
			Kind:       opCrds.ManagedResourceKind,
			Metadata: opCrds.ManagedResourceMetadata{
				Name:      string(create.Id),
				Namespace: create.Namespace,
			},
			Spec: opCrds.ManagedResourceSpec{
				MsvcRef: opCrds.MsvcRef{
					APIVersion: template.ApiVersion,
					Kind:       template.Kind,
					Name:       string(svc.Id),
				},
				MresKind: opCrds.MresKind{
					Kind: resTmpl.Kind,
				},
				Inputs: func() map[string]string {
					create.Values["resourceName"] = svc.Name
					return create.Values
				}(),
			},
		},
	)
	if err != nil {
		return nil, err
	}

	return create, nil
}
func (d *domain) UpdateManagedRes(ctx context.Context, managedResID repos.ID, values map[string]string) (bool, error) {
	mres, err := d.managedResRepo.FindById(ctx, managedResID)
	if err = mongoError(err, "managed resource not found"); err != nil {
		return false, err
	}

	msvc, err := d.managedSvcRepo.FindById(ctx, mres.ServiceId)
	if err != nil {
		return false, err
	}
	template, err := d.GetManagedServiceTemplate(ctx, string(msvc.ServiceType))
	if err != nil {
		return false, err
	}

	err = d.checkProjectAccess(ctx, mres.ProjectId, UpdateProject)
	if err != nil {
		return false, err
	}

	mres.Values = values
	_, err = d.managedResRepo.UpdateById(ctx, managedResID, mres)
	if err != nil {
		return false, err
	}

	clusterId, err := d.getClusterIdForProject(ctx, msvc.ProjectId)
	if err != nil {
		return false, err
	}

	err = d.workloadMessenger.SendAction(
		"apply", d.getDispatchKafkaTopic(clusterId), string(mres.Id), &opCrds.ManagedResource{
			APIVersion: opCrds.ManagedResourceAPIVersion,
			Kind:       opCrds.ManagedResourceKind,
			Metadata: opCrds.ManagedResourceMetadata{
				Name:      string(mres.Id),
				Namespace: mres.Namespace,
			},
			Spec: opCrds.ManagedResourceSpec{
				MsvcRef: opCrds.MsvcRef{
					APIVersion: template.ApiVersion,
					Kind:       template.Kind,
					Name:       string(mres.ServiceId),
				},
				MresKind: opCrds.MresKind{
					Kind: string(mres.ResourceType),
				},
				Inputs: func() map[string]string {
					mres.Values["resourceName"] = mres.Name
					return mres.Values
				}(),
			},
		},
	)
	if err != nil {
		return false, err
	}
	return true, nil
}

func (d *domain) UnInstallManagedRes(ctx context.Context, mresId repos.ID) (bool, error) {
	mres, err := d.managedResRepo.FindById(ctx, mresId)

	if err = mongoError(err, "managed resource not found"); err != nil {
		return false, err
	}

	err = d.checkProjectAccess(ctx, mres.ProjectId, UpdateProject)
	if err != nil {
		return false, err
	}

	if err != nil {
		return false, err
	}
	err = d.managedResRepo.DeleteById(ctx, mresId)
	if err != nil {
		return false, err
	}

	clusterId, err := d.getClusterIdForProject(ctx, mres.ProjectId)
	if err != nil {
		return false, err
	}

	err = d.workloadMessenger.SendAction(
		"delete", d.getDispatchKafkaTopic(clusterId), string(mresId), &opCrds.ManagedResource{
			APIVersion: opCrds.ManagedResourceAPIVersion,
			Kind:       opCrds.ManagedResourceKind,
			Metadata: opCrds.ManagedResourceMetadata{
				Name:      string(mresId),
				Namespace: mres.Namespace,
			},
		},
	)
	if err != nil {
		return false, err
	}
	return true, nil
}

func (d *domain) getManagedResOutput(ctx context.Context, managedResID repos.ID) (map[string]any, error) {
	mres, err := d.managedResRepo.FindById(ctx, managedResID)

	if err = mongoError(err, "managed resource not found"); err != nil {
		return nil, err
	}

	err = d.checkProjectAccess(ctx, mres.ProjectId, UpdateProject)
	if err != nil {
		return nil, err
	}

	kubecli := kubeapi.NewClientWithConfigPath(fmt.Sprintf("%s/kl-01", d.clusterConfigsPath))

	secret, err := kubecli.GetSecret(ctx, mres.Namespace, fmt.Sprint("mres-", mres.Id))
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

	mres, err := d.managedResRepo.FindById(ctx, managedResID)
	if err = mongoError(err, "managed resource not found"); err != nil {
		return nil, err
	}

	err = d.checkProjectAccess(ctx, mres.ProjectId, ReadProject)
	if err != nil {
		return nil, err
	}

	return d.getManagedResOutput(ctx, managedResID)
}

func (d *domain) OnDeleteManagedResource(ctx context.Context, response *opCrds.StatusUpdate) error {
	return d.managedResRepo.DeleteById(ctx, repos.ID(response.Metadata.ResourceId))
}
