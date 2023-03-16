package domain

import (
	"context"
	"fmt"
	"kloudlite.io/apps/consolev2.old/internal/domain/entities"
	"kloudlite.io/pkg/errors"
	"kloudlite.io/pkg/repos"
)

func (d *domain) CreateConfig(ctx context.Context, config entities.Config) (*entities.Config, error) {
	exists, err := d.configRepo.Exists(ctx, repos.Filter{"metadata.name": config.Name, "metadata.namespace": config.Namespace})
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, errors.Newf("config %s already exists", config.Name)
	}

	clusterId, err := d.getClusterIdForNamespace(ctx, config.Namespace)
	if err != nil {
		return nil, err
	}
	cfg, err := d.configRepo.Create(ctx, &config)
	if err != nil {
		return nil, err
	}
	if err := d.workloadMessenger.SendAction(ActionApply, d.getDispatchKafkaTopic(clusterId), string(cfg.Id), cfg.Config); err != nil {
		return nil, err
	}
	return cfg, nil
}

func (d *domain) UpdateConfig(ctx context.Context, config entities.Config) (*entities.Config, error) {
	cfg, err := d.configRepo.FindOne(ctx, repos.Filter{"metadata.name": config.Name, "metadata.namespace": config.Namespace})
	if err != nil {
		return nil, err
	}
	if cfg != nil {
		return nil, errors.Newf("secret  %s already exists", config.Name)
	}

	clusterId, err := d.getClusterForProject(ctx, config.ProjectName)
	cfg.Config = config.Config

	uCfg, err := d.configRepo.UpdateById(ctx, cfg.Id, cfg)
	if err != nil {
		return nil, err
	}
	if err := d.workloadMessenger.SendAction(ActionApply, d.getDispatchKafkaTopic(clusterId), string(cfg.Id), cfg.Config); err != nil {
		return nil, err
	}
	return uCfg, nil
}

func (d *domain) GetConfigs(ctx context.Context, namespace string, search *string) ([]*entities.Config, error) {
	if search == nil {
		return d.configRepo.Find(ctx, repos.Query{Filter: repos.Filter{"metadata.namespace": namespace}})
	}
	return d.configRepo.Find(ctx, repos.Query{Filter: repos.Filter{"metadata.namespace": namespace, "metadata.name": fmt.Sprintf("/%s/", *search)}})
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
