package domain

import (
	"context"

	"kloudlite.io/apps/console/internal/domain/entities"
	"kloudlite.io/pkg/repos"
)

func (d *domain) CreateEnvironment(ctx context.Context, blueprintID *repos.ID, name string) (*entities.Environment, error) {

	return d.environmentRepo.Create(ctx, &entities.Environment{
		BlueprintId: blueprintID,
		Name:        name,
	})

}

func (d *domain) GetEnvironment(ctx context.Context, envId repos.ID) (*entities.Environment, error) {
	return d.environmentRepo.FindById(ctx, envId)
}
