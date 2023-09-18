package domain

import (
	"fmt"
	"time"

	"kloudlite.io/apps/console/internal/entities"
	"kloudlite.io/common"
	"kloudlite.io/pkg/repos"
	t "kloudlite.io/pkg/types"
)

func (d *domain) ListConfigs(ctx ConsoleContext, namespace string, search map[string]repos.MatchFilter, pq repos.CursorPagination) (*repos.PaginatedRecord[*entities.Config], error) {
	//
	if err := d.canReadResourcesInWorkspace(ctx, namespace); err != nil {
		return nil, err
	}
	filter := repos.Filter{
		"accountName":        ctx.AccountName,
		"clusterName":        ctx.ClusterName,
		"metadata.namespace": namespace,
	}

	return d.configRepo.FindPaginated(ctx, d.configRepo.MergeMatchFilters(filter, search), pq)
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
	if err := d.canReadResourcesInWorkspace(ctx, namespace); err != nil {
		return nil, err
	}
	return d.findConfig(ctx, namespace, name)
}

// mutations

func (d *domain) CreateConfig(ctx ConsoleContext, config entities.Config) (*entities.Config, error) {
	if err := d.canMutateResourcesInWorkspace(ctx, config.Namespace); err != nil {
		return nil, err
	}

	config.EnsureGVK()
	if err := d.k8sExtendedClient.ValidateStruct(ctx, &config.Config); err != nil {
		return nil, err
	}

	config.IncrementRecordVersion()

	config.CreatedBy = common.CreatedOrUpdatedBy{
		UserId:    ctx.UserId,
		UserName:  ctx.UserName,
		UserEmail: ctx.UserEmail,
	}
	config.LastUpdatedBy = config.CreatedBy

	config.AccountName = ctx.AccountName
	config.ClusterName = ctx.ClusterName
	config.SyncStatus = t.GenSyncStatus(t.SyncActionApply, config.RecordVersion)

	c, err := d.configRepo.Create(ctx, &config)
	if err != nil {
		if d.configRepo.ErrAlreadyExists(err) {
			// TODO: better insights into error, when it is being caused by duplicated indexes
			return nil, err
		}
		return nil, err
	}

	if err := d.applyK8sResource(ctx, &c.Config, c.RecordVersion); err != nil {
		return c, err
	}

	return c, nil
}

func (d *domain) UpdateConfig(ctx ConsoleContext, config entities.Config) (*entities.Config, error) {
	if err := d.canMutateResourcesInWorkspace(ctx, config.Namespace); err != nil {
		return nil, err
	}

	config.EnsureGVK()
	if err := d.k8sExtendedClient.ValidateStruct(ctx, &config.Config); err != nil {
		return nil, err
	}

	currConfig, err := d.findConfig(ctx, config.Namespace, config.Name)
	if err != nil {
		return nil, err
	}

	currConfig.IncrementRecordVersion()

	currConfig.LastUpdatedBy = common.CreatedOrUpdatedBy{
		UserId:    ctx.UserId,
		UserName:  ctx.UserName,
		UserEmail: ctx.UserEmail,
	}
	currConfig.DisplayName = config.DisplayName

	currConfig.Labels = config.Labels
	currConfig.Annotations = config.Annotations
	currConfig.Data = config.Data

	currConfig.SyncStatus = t.GenSyncStatus(t.SyncActionApply, currConfig.RecordVersion)

	upConfig, err := d.configRepo.UpdateById(ctx, currConfig.Id, currConfig)
	if err != nil {
		return nil, err
	}

	if err := d.applyK8sResource(ctx, &upConfig.Config, upConfig.RecordVersion); err != nil {
		return upConfig, err
	}

	return upConfig, nil
}

func (d *domain) DeleteConfig(ctx ConsoleContext, namespace string, name string) error {
	if err := d.canMutateResourcesInWorkspace(ctx, namespace); err != nil {
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

func (d *domain) OnApplyConfigError(ctx ConsoleContext, errMsg, namespace, name string) error {
	c, err := d.findConfig(ctx, namespace, name)
	if err != nil {
		return err
	}

	c.SyncStatus.State = t.SyncStateErroredAtAgent
	c.SyncStatus.LastSyncedAt = time.Now()
	c.SyncStatus.Error = &errMsg

	_, err = d.configRepo.UpdateById(ctx, c.Id, c)
	return err
}

func (d *domain) OnDeleteConfigMessage(ctx ConsoleContext, config entities.Config) error {
	c, err := d.findConfig(ctx, config.Namespace, config.Name)
	if err != nil {
		return err
	}

	if err := d.MatchRecordVersion(config.Annotations, c.RecordVersion); err != nil {
		return d.resyncK8sResource(ctx, c.SyncStatus.Action, &c.Config, c.RecordVersion)
	}

	return d.configRepo.DeleteById(ctx, c.Id)
}

func (d *domain) OnUpdateConfigMessage(ctx ConsoleContext, config entities.Config) error {
	exConfig, err := d.findConfig(ctx, config.Namespace, config.Name)
	if err != nil {
		return err
	}

	annotatedVersion, err := d.parseRecordVersionFromAnnotations(config.Annotations)
	if err != nil {
		return d.resyncK8sResource(ctx, exConfig.SyncStatus.Action, &exConfig.Config, exConfig.RecordVersion)
	}

	if annotatedVersion != exConfig.RecordVersion {
		return d.resyncK8sResource(ctx, exConfig.SyncStatus.Action, &exConfig.Config, exConfig.RecordVersion)
	}

	exConfig.CreationTimestamp = config.CreationTimestamp
	exConfig.Labels = config.Labels
	exConfig.Annotations = config.Annotations
	exConfig.Generation = config.Generation

	exConfig.Status = config.Status

	exConfig.SyncStatus.State = t.SyncStateReceivedUpdateFromAgent
	exConfig.SyncStatus.RecordVersion = annotatedVersion
	exConfig.SyncStatus.Error = nil
	exConfig.SyncStatus.LastSyncedAt = time.Now()

	_, err = d.configRepo.UpdateById(ctx, exConfig.Id, exConfig)
	return err
}

func (d *domain) ResyncConfig(ctx ConsoleContext, namespace, name string) error {
	if err := d.canMutateResourcesInWorkspace(ctx, namespace); err != nil {
		return err
	}

	c, err := d.findConfig(ctx, namespace, name)
	if err != nil {
		return err
	}

	return d.resyncK8sResource(ctx, c.SyncStatus.Action, &c.Config, c.RecordVersion)
}
