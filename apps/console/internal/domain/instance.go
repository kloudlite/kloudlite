package domain

import (
	"context"
	"fmt"
	"strings"

	"kloudlite.io/apps/console/internal/app/graph/model"
	"kloudlite.io/apps/console/internal/domain/entities"
	op_crds "kloudlite.io/apps/console/internal/domain/op-crds"
	"kloudlite.io/common"
	"kloudlite.io/pkg/repos"

	createjsonpatch "github.com/snorwin/jsonpatch"
)

const (
	ENV_INSTANCE string = "self"
)

func (d *domain) GetResInstances(ctx context.Context, envId repos.ID, resType string) ([]*entities.ResInstance, error) {
	return d.instanceRepo.Find(ctx, repos.Query{
		Filter: repos.Filter{
			"environment_id": envId,
			"resource_type":  resType,
			"is_deleted":     false,
		},
	})
}

func (d *domain) ValidateResourecType(ctx context.Context, resType string) error {
	switch common.ResourceType(resType) {
	case common.ResourceApp,
		common.ResourceRouter,
		common.ResourceConfig,
		common.ResourceSecret,
		common.ResourceManagedResource,
		common.ResourceManagedService:
		return nil

	default:

		return fmt.Errorf(
			"resource type is not valid, use one of [%s, %s, %s, %s, %s, %s] resource type",
			common.ResourceApp,
			common.ResourceRouter,
			common.ResourceConfig,
			common.ResourceSecret,
			common.ResourceManagedResource,
			common.ResourceManagedService,
		)
	}
}

func (d *domain) GetResInstance(ctx context.Context, envID repos.ID, resID string) (*entities.ResInstance, error) {
	inst, err := d.instanceRepo.FindOne(ctx,
		repos.Filter{
			"environment_id": envID,
			"resource_id":    resID,
			"is_deleted":     false,
		})
	if err != nil {
		return nil, err
	}
	if inst == nil {
		return nil, fmt.Errorf("no resource found with given id")
	}
	return inst, nil
}

func (d *domain) GetResInstanceById(ctx context.Context, instanceId repos.ID) (*entities.ResInstance, error) {
	return d.instanceRepo.FindById(ctx, repos.ID(instanceId))
}

func (d *domain) DeleteInstance(ctx context.Context, instanceId repos.ID) error {
	instance, err := d.instanceRepo.FindById(ctx, instanceId)
	if err != nil {
		return err
	}

	env, err := d.environmentRepo.FindById(ctx, instance.EnvironmentId)
	if err != nil {
		return err
	}

	project, err := d.projectRepo.FindById(ctx, instance.BlueprintId)
	if err != nil {
		return err
	}

	clusterId, err := d.getClusterForAccount(ctx, project.AccountId)
	if err != nil {
		return err
	}

	if instance.IsSelf {
		instance.IsDeleted = true

	} else {
		instance.Enabled = false
	}

	// if _, err = d.instanceRepo.UpdateById(ctx, instanceId, instance); err != nil {
	// 	return err
	// }

	var apiVersion, kind, name string
	name = string(instance.ResourceId)
	isSelf := strings.HasPrefix(string(instance.ResourceId), ENV_INSTANCE)

	switch instance.ResourceType {

	case common.ResourceConfig:
		apiVersion = op_crds.ConfigAPIVersion
		kind = op_crds.ConfigKind

		if !isSelf {
			if c, err := d.configRepo.FindById(ctx, instance.ResourceId); err != nil {
				return err
			} else {
				name = string(c.Id)
			}
		}

	}

	if isSelf {

		if err = d.workloadMessenger.SendAction("delete", d.getDispatchKafkaTopic(clusterId), string(instanceId), &op_crds.Resource{
			APIVersion: apiVersion,
			Kind:       kind,
			Metadata: op_crds.ResourceMetadata{
				Name:      name,
				Namespace: fmt.Sprintf("%s-%s", project.Name, string(env.ReadableId)),
			},
		}); err != nil {
			return err
		}

	} else {

		if err = d.workloadMessenger.SendAction("apply", d.getDispatchKafkaTopic(clusterId), string(instanceId), &op_crds.Resource{
			APIVersion: apiVersion,
			Kind:       kind,
			Metadata: op_crds.ResourceMetadata{
				Name:      name,
				Namespace: fmt.Sprintf("%s-%s", project.Name, string(env.ReadableId)),
			},
			Enabled: &instance.Enabled,
		}); err != nil {
			return err
		}
	}

	return nil
}

func (d *domain) UpdateInstance(ctx context.Context, instance *entities.ResInstance, project *entities.Project, jsonPatchList *createjsonpatch.JSONPatchList, enabled *bool, overrides *string) (*entities.ResInstance, error) {

	func() {
		if enabled != nil {
			instance.Enabled = *enabled
		}
		if overrides != nil {
			instance.Overrides = *overrides
		} else {
			instance.Overrides = "[]"
		}
	}()

	inst, err := d.instanceRepo.UpdateById(ctx, instance.Id, instance)
	if err != nil {
		return nil, err
	}

	env, err := d.environmentRepo.FindById(ctx, inst.EnvironmentId)
	if err != nil {
		return nil, err
	}

	clusterId, err := d.getClusterForAccount(ctx, project.AccountId)
	if err != nil {
		return nil, err
	}

	var apiVersion, kind, name string
	name = string(instance.ResourceId)
	isSelf := strings.HasPrefix(string(instance.ResourceId), ENV_INSTANCE)

	switch inst.ResourceType {
	case common.ResourceRouter:
		return nil, fmt.Errorf("not implemented")

	case common.ResourceConfig:
		apiVersion = op_crds.ConfigAPIVersion
		kind = op_crds.ConfigKind

		if !isSelf {
			if c, err := d.configRepo.FindById(ctx, instance.ResourceId); err != nil {
				return nil, err
			} else {
				name = string(c.Id)
			}
		}

	case common.ResourceSecret:
		apiVersion = op_crds.SecretAPIVersion
		kind = op_crds.SecretKind

		if !isSelf {
			if s, err := d.secretRepo.FindById(ctx, instance.ResourceId); err != nil {
				return nil, err
			} else {
				name = string(s.Id)
			}
		}

	case common.ResourceManagedService:
		apiVersion = op_crds.ManagedServiceAPIVersion
		kind = op_crds.ManagedServiceKind

		if !isSelf {
			if m, err := d.managedSvcRepo.FindById(ctx, instance.ResourceId); err != nil {
				return nil, err
			} else {
				name = string(m.Id)
			}
		}

	case common.ResourceManagedResource:
		apiVersion = op_crds.ManagedResourceAPIVersion
		kind = op_crds.ManagedResourceKind
		if !isSelf {
			if r, err := d.managedResRepo.FindById(ctx, instance.ResourceId); err != nil {
				return nil, err
			} else {
				name = string(r.Id)
			}
		}

	case common.ResourceApp:
		apiVersion = op_crds.AppAPIVersion
		kind = op_crds.AppKind
		if !isSelf {
			if a, err := d.appRepo.FindById(ctx, instance.ResourceId); err != nil {
				return nil, err
			} else {
				name = a.ReadableId
			}
		} else {
			if instance.ExtraDatas != nil {
				name = string(instance.ExtraDatas.ReadableId)
			} else {
				return nil, fmt.Errorf("readable_id not found to apply")
			}
		}

	default:
		return nil, fmt.Errorf("resource_type not found")
	} // switch end

	if err = d.workloadMessenger.SendAction("apply", d.getDispatchKafkaTopic(clusterId), string(inst.Id), &op_crds.Resource{
		APIVersion: apiVersion,
		Kind:       kind,
		Metadata: op_crds.ResourceMetadata{
			Name:      name,
			Namespace: fmt.Sprintf("%s-%s", project.Name, string(env.ReadableId)),
		},
		Overrides: &op_crds.Overrides{
			Patches: jsonPatchList.List(),
		},
		Enabled: enabled,
	}); err != nil {
		return nil, err
	}

	return nil, nil
}

func (d *domain) CreateResInstance(ctx context.Context, resourceId repos.ID, environmentId repos.ID, blueprintId repos.ID, resType string, overrides string, isSelf bool) (*entities.ResInstance, error) {
	return d.instanceRepo.Create(ctx,
		&entities.ResInstance{
			IsSelf:        isSelf,
			Overrides:     overrides,
			ResourceId:    resourceId,
			EnvironmentId: environmentId,
			BlueprintId:   blueprintId,
			ResourceType:  common.ResourceType(resType),
		})
}

func (d *domain) ReturnResInstance(ctx context.Context, instance *entities.ResInstance) *model.ResInstance {

	return &model.ResInstance{
		ID:            instance.Id,
		ResourceID:    instance.ResourceId,
		EnvironmentID: instance.EnvironmentId,
		BlueprintID:   instance.BlueprintId,
		Overrides:     &instance.Overrides,
		ResourceType:  string(instance.ResourceType),
	}

}

var types = map[string]common.ResourceType{
	"App":             common.ResourceApp,
	"Config":          common.ResourceConfig,
	"Secret":          common.ResourceSecret,
	"Lambda":          common.ResourceLambda,
	"Router":          common.ResourceRouter,
	"ManagedResource": common.ResourceManagedResource,
	"ManagedService":  common.ResourceManagedService,
}

func validateValues(response *op_crds.StatusUpdate) error {
	if response.Metadata.ResourceId == "" {
		return fmt.Errorf("resource id not provided")
	} else if response.Metadata.ProjectId == "" {
		return fmt.Errorf("project id/ blueprint id not provided")
	} else if response.Metadata.EnvironmentId == "" {
		return fmt.Errorf("environment id not provided")
	} else if types[response.Metadata.GroupVersionKind.Kind] == "" {
		return fmt.Errorf("group id not provided")
	}

	return nil
}

func (d *domain) OnUpdateInstance(ctx context.Context, response *op_crds.StatusUpdate) error {
	err := validateValues(response)
	if err != nil {
		fmt.Println(err)
		return err
	}

	if types[response.Metadata.GroupVersionKind.Kind] == "" {
		fmt.Println("unknown kind")
		return fmt.Errorf("unknown kind")
	}

	_, err = d.instanceRepo.Upsert(ctx, repos.Filter{
		"resource_id":    response.Metadata.ResourceId,
		"blueprint_id":   response.Metadata.ProjectId,
		"environment_id": response.Metadata.EnvironmentId,
		"resource_type":  string(types[response.Metadata.GroupVersionKind.Kind]),
	}, &entities.ResInstance{
		ResourceId:    repos.ID(response.Metadata.ResourceId),
		EnvironmentId: repos.ID(response.Metadata.EnvironmentId),
		BlueprintId:   repos.ID(response.Metadata.ProjectId),
		ResourceType:  types[response.Metadata.GroupVersionKind.Kind],
		Status: func() entities.InstanceStatus {
			if response.IsReady {
				return entities.InstanceStatus(entities.InstanceStateLive)
			}
			return entities.InstanceStateError
		}(),
		Conditions: response.Conditions,
	})
	if err != nil {
		fmt.Println(err)
	}

	return err

}

func (d *domain) OnDeleteInstance(ctx context.Context, response *op_crds.StatusUpdate) error {

	err := validateValues(response)
	if err != nil {
		return err
	}

	return d.instanceRepo.SilentUpdateMany(ctx, repos.Filter{
		"resource_id":    response.Metadata.ResourceId,
		"blueprint_id":   response.Metadata.ProjectId,
		"environment_id": response.Metadata.EnvironmentId,
		"resource_type":  string(types[response.Metadata.GroupVersionKind.Kind]),
	}, map[string]any{
		"is_deleted": true,
	})

}
