package domain

import (
	"fmt"

	"kloudlite.io/apps/console/internal/domain/entities"
	"kloudlite.io/pkg/repos"
)

func (d *domain) CreateConfig(ctx ConsoleContext, config entities.Config) (*entities.Config, error) {
	config.EnsureGVK()
	if err := d.k8sExtendedClient.ValidateStruct(ctx, &config.Config); err != nil {
		return nil, err
	}

	config.AccountName = ctx.accountName
	config.ClusterName = ctx.clusterName
	c, err := d.configRepo.Create(ctx, &config)
	if err != nil {
		if d.configRepo.ErrAlreadyExists(err) {
			return nil, fmt.Errorf("config with name '%s' already exists", config.Name)
		}
		return nil, err
	}

	if err := d.applyK8sResource(ctx, &c.Config); err != nil {
		return c, err
	}

	return c, nil
}

func (d *domain) DeleteConfig(ctx ConsoleContext, namespace string, name string) error {
	c, err := d.findConfig(ctx, namespace, name)
	if err != nil {
		return err
	}
	return d.deleteK8sResource(ctx, &c.Config)
}

func (d *domain) GetConfig(ctx ConsoleContext, namespace string, name string) (*entities.Config, error) {
	return d.findConfig(ctx, namespace, name)
}

func (d *domain) ListConfigs(ctx ConsoleContext, namespace string) ([]*entities.Config, error) {
	return d.configRepo.Find(ctx, repos.Query{Filter: repos.Filter{
		"accountName":        ctx.accountName,
		"clusterName":        ctx.clusterName,
		"metadata.namespace": namespace,
	}})
}

func (d *domain) UpdateConfig(ctx ConsoleContext, config entities.Config) (*entities.Config, error) {
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

func (d *domain) findConfig(ctx ConsoleContext, namespace string, name string) (*entities.Config, error) {
	cfg, err := d.configRepo.FindOne(ctx, repos.Filter{
		"clusterName":        ctx.clusterName,
		"accountName":        ctx.accountName,
		"metadata.namespace": namespace,
		"metadata.name":      name,
	})
	if err != nil {
		return nil, err
	}
	if cfg == nil {
		return nil, fmt.Errorf("no config with name=%s,namespace=%s found", name, namespace)
	}
	return cfg, nil
}
