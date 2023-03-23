package domain

import (
	"fmt"
	"time"

	"kloudlite.io/apps/console/internal/domain/entities"
	"kloudlite.io/pkg/repos"
	t "kloudlite.io/pkg/types"
)

func (d *domain) CreateConfig(ctx ConsoleContext, config entities.Config) (*entities.Config, error) {
	config.EnsureGVK()
	if err := d.k8sExtendedClient.ValidateStruct(ctx, &config.Config); err != nil {
		return nil, err
	}

	config.AccountName = ctx.accountName
	config.ClusterName = ctx.clusterName
	config.SyncStatus = t.GetSyncStatusForCreation()

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

	c.SyncStatus = t.GetSyncStatusForDeletion(c.Generation)
	if _, err := d.configRepo.UpdateById(ctx, c.Id, c); err != nil {
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

	c.Config = config.Config
	c.SyncStatus = t.GetSyncStatusForUpdation(c.Generation + 1)

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

func (d *domain) OnDeleteConfigMessage(ctx ConsoleContext, config entities.Config) error {
	a, err := d.findConfig(ctx, config.Namespace, config.Name)
	if err != nil {
		return err
	}

	return d.configRepo.DeleteById(ctx, a.Id)
}

func (d *domain) OnUpdateConfigMessage(ctx ConsoleContext, config entities.Config) error {
	c, err := d.findConfig(ctx, config.Namespace, config.Name)
	if err != nil {
		return err
	}

	c.Status = config.Status
	c.SyncStatus.LastSyncedAt = time.Now()
	c.SyncStatus.State = t.ParseSyncState(config.Status.IsReady)

	_, err = d.configRepo.UpdateById(ctx, c.Id, c)
	return err
}

func (d *domain) OnApplyConfigError(ctx ConsoleContext, err error, namespace, name string) error {
	c, err2 := d.findConfig(ctx, namespace, name)
	if err2 != nil {
		return err2
	}

	c.SyncStatus.Error = err.Error()
	_, err = d.configRepo.UpdateById(ctx, c.Id, c)
	return err
}
