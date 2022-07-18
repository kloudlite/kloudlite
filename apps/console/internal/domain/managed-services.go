package domain

import (
	"context"
	"fmt"
	"gopkg.in/yaml.v3"
	"kloudlite.io/apps/console/internal/domain/entities"
	op_crds "kloudlite.io/apps/console/internal/domain/op-crds"
	"kloudlite.io/pkg/errors"
	"kloudlite.io/pkg/repos"
	"os"
)

func (d *domain) GetManagedSvc(ctx context.Context, managedSvcID repos.ID) (*entities.ManagedService, error) {
	return d.managedSvcRepo.FindById(ctx, managedSvcID)
}
func (d *domain) GetManagedSvcs(ctx context.Context, projectID repos.ID) ([]*entities.ManagedService, error) {
	return d.managedSvcRepo.Find(ctx, repos.Query{Filter: repos.Filter{
		"project_id": projectID,
	}})
}
func (d *domain) GetManagedServiceTemplates(ctx context.Context) ([]*entities.ManagedServiceCategory, error) {
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
func (d *domain) GetManagedServiceTemplate(_ context.Context, name string) (*entities.ManagedServiceTemplate, error) {
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

func (d *domain) OnUpdateManagedSvc(ctx context.Context, response *op_crds.StatusUpdate) error {
	one, err := d.managedSvcRepo.FindOne(ctx, repos.Filter{
		"id": response.Metadata.ResourceId,
	})
	if err != nil {
		return err
	}
	if response.IsReady {
		one.Status = entities.ManagedServiceStateLive
	} else {
		one.Status = entities.ManagedServiceStateSyncing
	}
	one.Conditions = response.ChildConditions
	_, err = d.managedSvcRepo.UpdateById(ctx, one.Id, one)
	err = d.notifier.Notify(one.Id)
	if err != nil {
		return err
	}
	return err
}

func (d *domain) InstallManagedSvc(ctx context.Context, projectID repos.ID, templateID repos.ID, name string, values map[string]interface{}) (*entities.ManagedService, error) {
	prj, err := d.projectRepo.FindById(ctx, projectID)
	if err != nil {
		return nil, err
	}
	if prj == nil {
		return nil, fmt.Errorf("project not found")
	}
	create, err := d.managedSvcRepo.Create(ctx, &entities.ManagedService{
		Name:        name,
		Namespace:   prj.Name,
		ProjectId:   prj.Id,
		ServiceType: entities.ManagedServiceType(templateID),
		Values:      values,
		Status:      entities.ManagedServiceStateSyncing,
	})
	if err != nil {
		return nil, err
	}
	template, err := d.GetManagedServiceTemplate(ctx, string(templateID))
	err = d.workloadMessenger.SendAction("apply", string(create.Id), &op_crds.ManagedService{
		APIVersion: op_crds.ManagedServiceAPIVersion,
		Kind:       op_crds.ManagedServiceKind,
		Metadata: op_crds.ManagedServiceMetadata{
			Name:      string(create.Id),
			Namespace: create.Namespace,
		},
		Spec: op_crds.ManagedServiceSpec{
			ApiVersion: template.ApiVersion,
			Inputs: func() map[string]string {
				vs := make(map[string]string, 0)
				for k, v := range create.Values {
					vs[k] = v.(string)
				}
				return vs
			}(),
		},
	})
	if err != nil {
		return nil, err
	}
	return create, err
}
func (d *domain) UpdateManagedSvc(ctx context.Context, managedServiceId repos.ID, values map[string]interface{}) (bool, error) {
	managedSvc, err := d.managedSvcRepo.FindById(ctx, managedServiceId)
	if err != nil {
		return false, err
	}
	if managedSvc == nil {
		return false, fmt.Errorf("project not found")
	}
	managedSvc.Values = values
	managedSvc.Status = entities.ManagedServiceStateSyncing
	_, err = d.managedSvcRepo.UpdateById(ctx, managedServiceId, managedSvc)
	if err != nil {
		return false, err
	}
	return true, nil
}
func (d *domain) UnInstallManagedSvc(ctx context.Context, managedServiceId repos.ID) (bool, error) {
	err := d.managedSvcRepo.DeleteById(ctx, managedServiceId)
	if err != nil {
		return false, err
	}
	return true, nil
}
