package domain

import (
	"context"
	"kloudlite.io/apps/consolev2.old/internal/domain/entities"
	"kloudlite.io/pkg/errors"
	"kloudlite.io/pkg/repos"
)

func (d *domain) GetEnvironments(ctx context.Context, projectName string) ([]*entities.Environment, error) {
	return d.environmentRepo.Find(ctx, repos.Query{Filter: repos.Filter{"spec.projectName": projectName}})
}

func (d *domain) GetEnvironment(ctx context.Context, envName string) (*entities.Environment, error) {
	return d.environmentRepo.FindOne(ctx, repos.Filter{"metadata.name": envName})
}

func (d *domain) CreateEnvironment(ctx context.Context, env entities.Environment) (*entities.Environment, error) {
	// check for existing environment
	exists, err := d.environmentRepo.Exists(ctx, repos.Filter{"metadata.name": env.Name})
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, errors.Newf("environment with (name = %s) already exists", env.Name)
	}
	return d.environmentRepo.Create(ctx, &env)
}
