package domain

import (
	"context"
	"kloudlite.io/apps/consolev2/internal/domain/entities"
	"kloudlite.io/pkg/errors"
	"kloudlite.io/pkg/repos"
)

func (d *domain) CreateConfig(ctx context.Context, config entities.Config) (*entities.Config, error) {
	exists, err := d.configRepo.Exists(ctx, repos.Filter{"metadata.name": config.Name, "metadata.namespace": config.Namespace})
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, errors.Newf("secret  %s already exists", config.Name)
	}

	clusterId, err := d.getClusterForProject(ctx, config.Spec.ProjectName)
	cfg, err := d.configRepo.Create(ctx, &config)
	if err != nil {
		return nil, err
	}
	if err := d.workloadMessenger.SendAction(ActionApply, d.getDispatchKafkaTopic(clusterId), string(cfg.Id), cfg.Config); err != nil {
		return nil, err
	}
	return cfg, nil
}

func (d *domain) UpdateConfig(ctx context.Context, config entities.Config) (bool, error) {
	cfg, err := d.configRepo.FindOne(ctx, repos.Filter{"metadata.name": config.Name, "metadata.namespace": config.Namespace})
	if err != nil {
		return false, err
	}
	if cfg != nil {
		return false, errors.Newf("secret  %s already exists", config.Name)
	}

	clusterId, err := d.getClusterForProject(ctx, config.Spec.ProjectName)
	cfg.Config = config.Config

	if _, err := d.configRepo.UpdateById(ctx, cfg.Id, cfg); err != nil {
		return false, err
	}
	if err := d.workloadMessenger.SendAction(ActionApply, d.getDispatchKafkaTopic(clusterId), string(cfg.Id), cfg.Config); err != nil {
		return false, err
	}
	return true, nil
}

func (d *domain) GetConfigs(ctx context.Context, namespace string) ([]*entities.Config, error) {
	return d.configRepo.Find(ctx, repos.Query{Filter: repos.Filter{"metadata.namespace": namespace}})
}

func (d *domain) GetConfig(ctx context.Context, namespace string, name string) (*entities.Config, error) {
	return d.configRepo.FindOne(ctx, repos.Filter{"metadata.namespace": namespace, "metadata.name": name})
}

func (d *domain) DeleteConfig(ctx context.Context, namespace string, name string) (bool, error) {
	if err := d.configRepo.DeleteOne(ctx, repos.Filter{"metadata.namespace": namespace, "metadata.name": name}); err != nil {
		return false, err
	}
	return true, nil
}
