package domain

import (
	"context"
	"fmt"

	"kloudlite.io/apps/console/internal/domain/entities"
	"kloudlite.io/pkg/repos"
)

func (d *domain) CreateConfig(ctx context.Context, config entities.Config) (*entities.Config, error) {
	config.EnsureGVK()
	if err := d.k8sExtendedClient.ValidateStruct(ctx, &config.Config); err != nil {
		return nil, err
	}

	c, err := d.configRepo.Create(ctx, &config)
	if err != nil {
		return nil, err
	}

	if err := d.applyK8sResource(ctx, &c.Config); err != nil {
		return c, err
	}

	return c, nil
}

func (d *domain) DeleteConfig(ctx context.Context, namespace string, name string) error {
	c, err := d.findConfig(ctx, namespace, name)
	if err != nil {
		return err
	}
	return d.k8sYamlClient.DeleteResource(ctx, &c.Config)
}

func (d *domain) GetConfig(ctx context.Context, namespace string, name string) (*entities.Config, error) {
	return d.findConfig(ctx, namespace, name)
}

func (d *domain) GetConfigs(ctx context.Context, namespace string) ([]*entities.Config, error) {
	return d.configRepo.Find(ctx, repos.Query{Filter: repos.Filter{"metadata.namespace": namespace}})
}

func (d *domain) UpdateConfig(ctx context.Context, config entities.Config) (*entities.Config, error) {
	config.EnsureGVK()
	if err := d.k8sExtendedClient.ValidateStruct(ctx, &config.Config); err != nil {
		return nil, err
	}

	c, err := d.findConfig(ctx, config.Namespace, config.Name)
	if err != nil {
		return nil, err
	}

	status := c.Status
	c.Config = config.Config
	c.Status = status

	upConfig, err := d.configRepo.UpdateById(ctx, c.Id, c)
	if err != nil {
		return nil, err
	}

	if err := d.applyK8sResource(ctx, &upConfig.Config); err != nil {
		return upConfig, err
	}

	return upConfig, nil
}

func (d *domain) findConfig(ctx context.Context, namespace string, name string) (*entities.Config, error) {
	cfg, err := d.configRepo.FindOne(ctx, repos.Filter{"metadata.namespace": namespace, "metadata.name": name})
	if err != nil {
		return nil, err
	}
	if cfg == nil {
		return nil, fmt.Errorf("no config with name=%s,namespace=%s found", name, namespace)
	}
	return cfg, nil
}
