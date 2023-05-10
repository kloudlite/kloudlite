package domain

import (
	"fmt"
	"time"

	"kloudlite.io/apps/console/internal/domain/entities"
	"kloudlite.io/pkg/repos"
	t "kloudlite.io/pkg/types"
)

// query
func (d *domain) ListConfigs(ctx ConsoleContext, namespace string) ([]*entities.Config, error) {
	if err := d.canReadResourcesInProject(ctx, namespace); err != nil {
		return nil, err
	}
	return d.configRepo.Find(ctx, repos.Query{Filter: repos.Filter{
		"accountName":        ctx.AccountName,
		"clusterName":        ctx.ClusterName,
		"metadata.namespace": namespace,
	}})
}

func (d *domain) findConfig(ctx ConsoleContext, namespace string, name string) (*entities.Config, error) {
	cfg, err := d.configRepo.FindOne(ctx, repos.Filter{
		"clusterName":        ctx.ClusterName,
		"accountName":        ctx.AccountName,
		"metadata.namespace": namespace,
		"metadata.name":      name,
	})
	if err != nil {
		return nil, err
	}
	if cfg == nil {
		return nil, fmt.Errorf("no config with name=%q,namespace=%q found", name, namespace)
	}
	return cfg, nil
}

func (d *domain) GetConfig(ctx ConsoleContext, namespace string, name string) (*entities.Config, error) {
	if err := d.canReadResourcesInProject(ctx, namespace); err != nil {
		return nil, err
	}
	return d.findConfig(ctx, namespace, name)
}

// mutations

func (d *domain) CreateConfig(ctx ConsoleContext, config entities.Config) (*entities.Config, error) {
	if err := d.canMutateResourcesInProject(ctx, config.Namespace); err != nil {
		return nil, err
	}

	config.EnsureGVK()
	if err := d.k8sExtendedClient.ValidateStruct(ctx, &config.Config); err != nil {
		return nil, err
	}

	config.AccountName = ctx.AccountName
	config.ClusterName = ctx.ClusterName
	config.SetGeneration(1)
	config.SyncStatus = t.GetSyncStatusForCreation()

	c, err := d.configRepo.Create(ctx, &config)
	if err != nil {
		if d.configRepo.ErrAlreadyExists(err) {
			return nil, fmt.Errorf("config with name %q already exists", config.Name)
		}
		return nil, err
	}

	if err := d.applyK8sResource(ctx, &c.Config); err != nil {
		return c, err
	}

	return c, nil
}

func (d *domain) UpdateConfig(ctx ConsoleContext, config entities.Config) (*entities.Config, error) {
	if err := d.canMutateResourcesInProject(ctx, config.Namespace); err != nil {
		return nil, err
	}

	config.EnsureGVK()
	if err := d.k8sExtendedClient.ValidateStruct(ctx, &config.Config); err != nil {
		return nil, err
	}

	c, err := d.findConfig(ctx, config.Namespace, config.Name)
	if err != nil {
		return nil, err
	}

	c.Config = config.Config
	c.Generation += 1
	c.SyncStatus = t.GetSyncStatusForUpdation(c.Generation)

	upConfig, err := d.configRepo.UpdateById(ctx, c.Id, c)
	if err != nil {
		return nil, err
	}

	if err := d.applyK8sResource(ctx, &upConfig.Config); err != nil {
		return upConfig, err
	}

	return upConfig, nil
}

func (d *domain) DeleteConfig(ctx ConsoleContext, namespace string, name string) error {
	if err := d.canMutateResourcesInProject(ctx, namespace); err != nil {
		return err
	}

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

func (d *domain) OnApplyConfigError(ctx ConsoleContext, err error, namespace, name string) error {
	c, err2 := d.findConfig(ctx, namespace, name)
	if err2 != nil {
		return err2
	}

	c.SyncStatus.Error = err.Error()
	_, err = d.configRepo.UpdateById(ctx, c.Id, c)
	return err
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
	c.SyncStatus.Generation = config.Generation
	c.SyncStatus.State = t.ParseSyncState(config.Status.IsReady)

	_, err = d.configRepo.UpdateById(ctx, c.Id, c)
	return err
}

func (d *domain) ResyncConfig(ctx ConsoleContext, namespace, name string) error {
	if err := d.canMutateResourcesInProject(ctx, namespace); err != nil {
		return err
	}

	c, err := d.findConfig(ctx, namespace, name)
	if err != nil {
		return err
	}

	return d.resyncK8sResource(ctx, c.SyncStatus.Action, &c.Config)
}
