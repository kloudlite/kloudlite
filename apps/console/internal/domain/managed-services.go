package domain

import (
	"context"
	"encoding/json"
	"fmt"
	"google.golang.org/protobuf/types/known/anypb"
	"gopkg.in/yaml.v3"
	"kloudlite.io/apps/console/internal/domain/entities"
	op_crds "kloudlite.io/apps/console/internal/domain/op-crds"
	"kloudlite.io/grpc-interfaces/kloudlite.io/rpc/jseval"
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
	eval, err := d.jsEvalClient.Eval(ctx, &jseval.EvalIn{
		Init:    template.InputMiddleware,
		FunName: "inputMiddleware",
		Inputs: func() *anypb.Any {
			marshal, _ := json.Marshal(values)
			return &anypb.Any{
				TypeUrl: "",
				Value:   marshal,
			}
		}(),
	})
	transformedInputs := map[string]any{}
	err = json.Unmarshal(eval.Output.Value, &transformedInputs)
	fmt.Println(transformedInputs, err)
	err = d.workloadMessenger.SendAction("apply", string(create.Id), &op_crds.ManagedService{
		APIVersion: op_crds.ManagedServiceAPIVersion,
		Kind:       op_crds.ManagedServiceKind,
		Metadata: op_crds.ManagedServiceMetadata{
			Name:      string(create.Id),
			Namespace: create.Namespace,
			Annotations: func() map[string]string {
				if transformedInputs["annotations"] == nil {
					return nil
				}
				a := transformedInputs["annotations"].(map[string]string)
				return a
			}(),
		},
		Spec: op_crds.ManagedServiceSpec{
			NodeSelector: map[string]string{
				"kloudlite.io/region": prj.Region,
			},
			Inputs: func() map[string]string {
				vs := make(map[string]string, 0)
				for k, v := range transformedInputs["inputs"].(map[string]interface{}) {
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
	proj, err := d.projectRepo.FindById(ctx, managedSvc.ProjectId)
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
	template, err := d.GetManagedServiceTemplate(ctx, string(managedSvc.ServiceType))
	eval, err := d.jsEvalClient.Eval(ctx, &jseval.EvalIn{
		Init:    template.InputMiddleware,
		FunName: "inputMiddleware",
		Inputs: func() *anypb.Any {
			marshal, _ := json.Marshal(values)
			return &anypb.Any{
				TypeUrl: "",
				Value:   marshal,
			}
		}(),
	})
	transformedInputs := map[string]any{}
	err = json.Unmarshal(eval.Output.Value, &transformedInputs)
	err = d.workloadMessenger.SendAction("apply", string(managedSvc.Id), &op_crds.ManagedService{
		APIVersion: op_crds.ManagedServiceAPIVersion,
		Kind:       op_crds.ManagedServiceKind,
		Metadata: op_crds.ManagedServiceMetadata{
			Name:      string(managedSvc.Id),
			Namespace: managedSvc.Namespace,
			Annotations: func() map[string]string {
				if transformedInputs["annotations"] == nil {
					return nil
				}
				a := transformedInputs["annotations"].(map[string]string)
				return a
			}(),
		},
		Spec: op_crds.ManagedServiceSpec{
			CloudProvider: "do", // TODO:
			MsvcType: op_crds.MsvcType{
				APIVersion: template.ApiVersion,
				Kind:       "Service",
			},
			NodeSelector: map[string]string{
				"kloudlite.io/region": proj.Region,
			},
			Inputs: func() map[string]string {
				vs := make(map[string]string, 0)
				for k, v := range transformedInputs["inputs"].(map[string]interface{}) {
					vs[k] = v.(string)
				}
				return vs
			}(),
		},
	})
	return true, nil
}

func (d *domain) UnInstallManagedSvc(ctx context.Context, managedServiceId repos.ID) (bool, error) {
	managedSvc, err := d.managedSvcRepo.FindById(ctx, managedServiceId)
	if err != nil {
		return false, err
	}
	err = d.managedSvcRepo.DeleteById(ctx, managedServiceId)
	if err != nil {
		return false, err
	}
	err = d.workloadMessenger.SendAction("delete", string(managedServiceId), &op_crds.ManagedService{
		APIVersion: op_crds.ManagedServiceAPIVersion,
		Kind:       op_crds.ManagedServiceKind,
		Metadata: op_crds.ManagedServiceMetadata{
			Name:      string(managedServiceId),
			Namespace: managedSvc.Namespace,
		},
	})
	return true, nil
}
