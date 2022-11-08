package domain

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"google.golang.org/protobuf/types/known/anypb"
	"gopkg.in/yaml.v3"
	"kloudlite.io/apps/console/internal/domain/entities"
	opCrds "kloudlite.io/apps/console/internal/domain/op-crds"
	"kloudlite.io/grpc-interfaces/kloudlite.io/rpc/jseval"
	"kloudlite.io/pkg/errors"
	"kloudlite.io/pkg/kubeapi"
	"kloudlite.io/pkg/repos"
)

func (d *domain) GetManagedSvc(ctx context.Context, managedSvcID repos.ID) (*entities.ManagedService, error) {
	msvc, err := d.managedSvcRepo.FindById(ctx, managedSvcID)
	if err = mongoError(err, "managed service not found"); err != nil {
		return nil, err
	}

	err = d.checkProjectAccess(ctx, msvc.ProjectId, ReadProject)
	if err != nil {
		return nil, err
	}

	return msvc, nil
}

func (d *domain) GetManagedSvcs(ctx context.Context, projectID repos.ID) ([]*entities.ManagedService, error) {
	msvcs, err := d.managedSvcRepo.Find(
		ctx, repos.Query{Filter: repos.Filter{
			"project_id": projectID,
		}},
	)
	if err != nil {
		return nil, err
	}

	if len(msvcs) >= 1 {
		err = d.checkProjectAccess(ctx, msvcs[0].ProjectId, ReadProject)
		if err != nil {
			return nil, err
		}
	}

	return msvcs, nil

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

func (d *domain) OnUpdateManagedSvc(ctx context.Context, response *opCrds.StatusUpdate) error {
	one, err := d.managedSvcRepo.FindById(ctx, repos.ID(response.Metadata.ResourceId))
	if err = mongoError(err, "managed service not found"); err != nil {
		return err
	}

	newStatus := one.Status
	if response.IsReady {
		newStatus = entities.ManagedServiceStateLive
	}
	shouldUpdate := newStatus != one.Status
	one.Conditions = response.ChildConditions
	one.Status = newStatus
	_, err = d.managedSvcRepo.UpdateById(ctx, one.Id, one)
	if shouldUpdate {
		err = d.notifier.Notify(one.Id)
		if err != nil {
			return err
		}
	}
	return err
}

func (d *domain) InstallManagedSvc(ctx context.Context, projectID repos.ID, templateID repos.ID, name string, values map[string]interface{}) (*entities.ManagedService, error) {
	prj, err := d.projectRepo.FindById(ctx, projectID)
	if err = mongoError(err, "project not found"); err != nil {
		return nil, err
	}

	err = d.checkProjectAccess(ctx, projectID, UpdateProject)
	if err != nil {
		return nil, err
	}

	_, region, err := d.getProjectRegionDetails(ctx, prj)
	if err != nil {
		return nil, err
	}

	create, err := d.managedSvcRepo.Create(
		ctx, &entities.ManagedService{
			Name:        name,
			Namespace:   prj.Name,
			ProjectId:   prj.Id,
			ServiceType: entities.ManagedServiceType(templateID),
			Values:      values,
			Status:      entities.ManagedServiceStateSyncing,
		},
	)
	if err != nil {
		return nil, err
	}

	template, err := d.GetManagedServiceTemplate(ctx, string(templateID))
	if err != nil {
		return nil, err
	}

	eval, err := d.jsEvalClient.Eval(
		ctx, &jseval.EvalIn{
			Init:    template.InputMiddleware,
			FunName: "inputMiddleware",
			Inputs: func() *anypb.Any {
				marshal, _ := json.Marshal(values)
				return &anypb.Any{
					TypeUrl: "",
					Value:   marshal,
				}
			}(),
		},
	)
	if err != nil {
		return nil, err
	}

	var transformedInputs struct {
		Annotation map[string]string `json:"annotation,omitempty"`
		Inputs     map[string]any    `json:"inputs"`
		Error      error             `json:"error,omitempty"`
	}

	err = json.Unmarshal(eval.Output.Value, &transformedInputs)
	if err != nil {
		return nil, err
	}

	clusterId, err := d.getClusterForAccount(ctx, prj.AccountId)
	if err != nil {
		return nil, err
	}

	err = d.workloadMessenger.SendAction(
		"apply", d.getDispatchKafkaTopic(clusterId), string(create.Id), &opCrds.ManagedService{
			APIVersion: opCrds.ManagedServiceAPIVersion,
			Kind:       opCrds.ManagedServiceKind,
			Metadata: opCrds.ManagedServiceMetadata{
				Name:      string(create.Id),
				Namespace: create.Namespace,
				Annotations: func() map[string]string {
					if transformedInputs.Annotation == nil {
						return nil
					}
					a := transformedInputs.Annotation
					a["kloudlite.io/account-ref"] = string(prj.AccountId)
					a["kloudlite.io/project-ref"] = string(prj.Id)
					a["kloudlite.io/resource-ref"] = string(create.Id)
					a["kloudlite.io/updated-at"] = time.Now().String()
					return a
				}(),
			},
			Spec: opCrds.ManagedServiceSpec{
				// CloudProvider: op_crds.CloudProvider{
				// 	Cloud:  cloudProvider,
				// 	Region: region,
				// },
				Region: region,
				// NodeSelector: map[string]string{
				// 	"kloudlite.io/region": region,
				// },
				MsvcKind: opCrds.MsvcKind{
					APIVersion: template.ApiVersion,
					Kind:       template.Kind,
				},
				Inputs: transformedInputs.Inputs,
			},
		},
	)
	if err != nil {
		return nil, err
	}
	return create, err
}

func (d *domain) UpdateManagedSvc(ctx context.Context, managedServiceId repos.ID, values map[string]interface{}) (bool, error) {
	managedSvc, err := d.managedSvcRepo.FindById(ctx, managedServiceId)
	if err = mongoError(err, "managed service not found"); err != nil {
		return false, err
	}

	err = d.checkProjectAccess(ctx, managedSvc.ProjectId, UpdateProject)
	if err != nil {
		return false, err
	}

	proj, err := d.projectRepo.FindById(ctx, managedSvc.ProjectId)
	if err = mongoError(err, "managed resource not found"); err != nil {
		return false, err
	}

	_, region, err := d.getProjectRegionDetails(ctx, proj)
	if err != nil {
		return false, err
	}

	managedSvc.Values = values
	managedSvc.Status = entities.ManagedServiceStateSyncing
	_, err = d.managedSvcRepo.UpdateById(ctx, managedServiceId, managedSvc)
	if err != nil {
		return false, err
	}
	template, err := d.GetManagedServiceTemplate(ctx, string(managedSvc.ServiceType))
	if err != nil {
		return false, err
	}

	eval, err := d.jsEvalClient.Eval(
		ctx, &jseval.EvalIn{
			Init:    template.InputMiddleware,
			FunName: "inputMiddleware",
			Inputs: func() *anypb.Any {
				marshal, _ := json.Marshal(values)
				return &anypb.Any{
					TypeUrl: "",
					Value:   marshal,
				}
			}(),
		},
	)
	if err != nil {
		return false, err
	}

	var transformedInputs struct {
		Inputs     map[string]any    `json:"inputs"`
		Annotation map[string]string `json:"annotation,omitempty"`
	}
	err = json.Unmarshal(eval.Output.Value, &transformedInputs)
	if err != nil {
		return false, err
	}

	clusterId, err := d.getClusterForAccount(ctx, proj.AccountId)
	if err != nil {
		return false, err
	}

	err = d.workloadMessenger.SendAction(
		"apply", d.getDispatchKafkaTopic(clusterId), string(managedSvc.Id), &opCrds.ManagedService{
			APIVersion: opCrds.ManagedServiceAPIVersion,
			Kind:       opCrds.ManagedServiceKind,
			Metadata: opCrds.ManagedServiceMetadata{
				Name:      string(managedSvc.Id),
				Namespace: managedSvc.Namespace,
				Annotations: func() map[string]string {
					if transformedInputs.Annotation == nil {
						return nil
					}
					a := transformedInputs.Annotation
					a["kloudlite.io/account-ref"] = string(proj.AccountId)
					a["kloudlite.io/project-ref"] = string(proj.Id)
					a["kloudlite.io/resource-ref"] = string(managedSvc.Id)
					a["kloudlite.io/updated-at"] = time.Now().String()
					return a
				}(),
			},
			Spec: opCrds.ManagedServiceSpec{
				MsvcKind: opCrds.MsvcKind{
					APIVersion: template.ApiVersion,
					Kind:       template.Kind,
				},
				Region: region,
				// CloudProvider: op_crds.CloudProvider{
				// 	Cloud:  cloudProvider,
				// 	Region: region,
				// },
				// NodeSelector: map[string]string{
				// 	"kloudlite.io/region": region,
				// },
				Inputs: transformedInputs.Inputs,
			},
		},
	)
	if err != nil {
		return false, err
	}

	return true, nil
}

func (d *domain) UnInstallManagedSvc(ctx context.Context, managedServiceId repos.ID) (bool, error) {
	managedSvc, err := d.managedSvcRepo.FindById(ctx, managedServiceId)
	if err = mongoError(err, "managed resource not found"); err != nil {
		return false, err
	}

	err = d.checkProjectAccess(ctx, managedSvc.ProjectId, UpdateProject)
	if err != nil {
		return false, err
	}

	managedSvc.Status = entities.ManagedServiceStateDeleting
	_, err = d.managedSvcRepo.UpdateById(ctx, managedServiceId, managedSvc)
	if err != nil {
		return false, err
	}

	clusterId, err := d.getClusterIdForProject(ctx, managedSvc.ProjectId)
	if err != nil {
		return false, err
	}

	err = d.workloadMessenger.SendAction(
		"delete", d.getDispatchKafkaTopic(clusterId), string(managedServiceId), &opCrds.ManagedService{
			APIVersion: opCrds.ManagedServiceAPIVersion,
			Kind:       opCrds.ManagedServiceKind,
			Metadata: opCrds.ManagedServiceMetadata{
				Name:      string(managedServiceId),
				Namespace: managedSvc.Namespace,
			},
		},
	)
	if err != nil {
		return false, err
	}

	return true, nil
}

func (d *domain) GetManagedSvcOutput(ctx context.Context, managedSvcID repos.ID) (map[string]any, error) {
	msvc, err := d.managedSvcRepo.FindById(ctx, managedSvcID)
	if err = mongoError(err, "managed service not found"); err != nil {
		return nil, err
	}

	err = d.checkProjectAccess(ctx, msvc.ProjectId, ReadProject)
	if err != nil {
		return nil, err
	}

	project, err := d.projectRepo.FindById(ctx, msvc.ProjectId)
	if err != nil {
		return nil, err
	}

	cluster, err := d.getClusterForAccount(ctx, project.AccountId)
	if err != nil {
		return nil, err
	}

	kubecli := kubeapi.NewClientWithConfigPath(fmt.Sprintf("%s/%s", d.clusterConfigsPath, getClusterKubeConfig(cluster)))
	secret, err := kubecli.GetSecret(ctx, msvc.Namespace, fmt.Sprint("msvc-", msvc.Id))
	if err != nil {
		return nil, err
	}
	parsedSec := make(map[string]any)
	for k, v := range secret.Data {
		parsedSec[k] = string(v)
	}
	return parsedSec, nil
}

func (d *domain) OnDeleteManagedService(ctx context.Context, response *opCrds.StatusUpdate) error {
	return d.managedSvcRepo.DeleteById(ctx, repos.ID(response.Metadata.ResourceId))
}
