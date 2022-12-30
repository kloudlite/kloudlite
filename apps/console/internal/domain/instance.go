package domain

import (
	"context"

	"kloudlite.io/apps/console/internal/app/graph/model"
	"kloudlite.io/apps/console/internal/domain/entities"
	"kloudlite.io/common"
	"kloudlite.io/pkg/repos"
)

func (d *domain) GetResInstances(ctx context.Context, envId repos.ID, resType string) ([]*entities.ResInstance, error) {
	return d.instanceRepo.Find(ctx, repos.Query{
		Filter: repos.Filter{
			"environment_id": envId,
			"resource_type":  resType,
		},
	})
}

func (d *domain) ValidateResourecType(ctx context.Context, resType string) bool {
	switch common.ResourceType(resType) {
	case common.ResourceApp,
		common.ResourceRouter,
		common.ResourceConfig,
		common.ResourceSecret,
		common.ResourceManagedResource,
		common.ResourceManagedService:
		return true
	default:
		return false
	}
}

func (d *domain) GetResInstance(ctx context.Context, envID repos.ID, resID string) (*entities.ResInstance, error) {
	return d.instanceRepo.FindOne(ctx,
		repos.Filter{
			"environment_id": envID,
			"resource_id":    resID,
		})
}

func (d *domain) UpdateInstance(ctx context.Context, resID repos.ID, resType string, overrides string) (*entities.ResInstance, error) {
	return d.instanceRepo.UpdateById(ctx, resID, &entities.ResInstance{
		Overrides: overrides,
	})
}

func (d *domain) CreateResInstance(ctx context.Context, resourceId repos.ID, environmentId repos.ID, blueprintId *repos.ID, resType string, overrides string) (*entities.ResInstance, error) {
	return d.instanceRepo.Create(ctx,
		&entities.ResInstance{
			Overrides:     overrides,
			ResourceId:    resourceId,
			EnvironmentId: environmentId,
			BlueprintId:   blueprintId,
			ResourceType:  common.ResourceType(resType),
		})
}

func (d *domain) GetEnvironments(ctx context.Context, blueprintID repos.ID) ([]*entities.Environment, error) {
	return d.environmentRepo.Find(ctx, repos.Query{
		Filter: repos.Filter{
			"blueprint_id": blueprintID,
		},
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
