package domain

import (
	"github.com/kloudlite/api/apps/console/internal/entities"
	"github.com/kloudlite/api/common"
	"github.com/kloudlite/api/pkg/errors"
	"github.com/kloudlite/api/pkg/repos"
	t "github.com/kloudlite/api/pkg/types"
	"github.com/kloudlite/operator/operators/resource-watcher/types"
)

func (d *domain) ListConfigs(ctx ConsoleContext, namespace string, search map[string]repos.MatchFilter, pq repos.CursorPagination) (*repos.PaginatedRecord[*entities.Config], error) {
	//
	if err := d.canReadResourcesInWorkspace(ctx, namespace); err != nil {
		return nil, errors.NewE(err)
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
		return nil, errors.NewE(err)
	}
	if cfg == nil {
		return nil, errors.Newf("no config with name=%q,namespace=%q found", name, namespace)
	}
	return cfg, nil
}

func (d *domain) GetConfig(ctx ConsoleContext, namespace string, name string) (*entities.Config, error) {
	if err := d.canReadResourcesInWorkspace(ctx, namespace); err != nil {
		return nil, errors.NewE(err)
	}
	return d.findConfig(ctx, namespace, name)
}

// mutations

func (d *domain) CreateConfig(ctx ConsoleContext, config entities.Config) (*entities.Config, error) {
	if err := d.canMutateResourcesInWorkspace(ctx, config.Namespace); err != nil {
		return nil, errors.NewE(err)
	}

	config.EnsureGVK()
	if err := d.k8sClient.ValidateObject(ctx, &config.Config); err != nil {
		return nil, errors.NewE(err)
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
			return nil, errors.NewE(err)
		}
		return nil, errors.NewE(err)
	}

	if err := d.applyK8sResource(ctx, &c.Config, c.RecordVersion); err != nil {
		return c, errors.NewE(err)
	}

	return c, nil
}

func (d *domain) UpdateConfig(ctx ConsoleContext, config entities.Config) (*entities.Config, error) {
	if err := d.canMutateResourcesInWorkspace(ctx, config.Namespace); err != nil {
		return nil, errors.NewE(err)
	}

	config.EnsureGVK()
	if err := d.k8sClient.ValidateObject(ctx, &config.Config); err != nil {
		return nil, errors.NewE(err)
	}

	currConfig, err := d.findConfig(ctx, config.Namespace, config.Name)
	if err != nil {
		return nil, errors.NewE(err)
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
		return nil, errors.NewE(err)
	}

	if err := d.applyK8sResource(ctx, &upConfig.Config, upConfig.RecordVersion); err != nil {
		return upConfig, errors.NewE(err)
	}

	return upConfig, nil
}

func (d *domain) DeleteConfig(ctx ConsoleContext, namespace string, name string) error {
	if err := d.canMutateResourcesInWorkspace(ctx, namespace); err != nil {
		return errors.NewE(err)
	}

	c, err := d.findConfig(ctx, namespace, name)
	if err != nil {
		return errors.NewE(err)
	}

	c.SyncStatus = t.GetSyncStatusForDeletion(c.Generation)
	if _, err := d.configRepo.UpdateById(ctx, c.Id, c); err != nil {
		return errors.NewE(err)
	}

	return d.deleteK8sResource(ctx, &c.Config)
}

func (d *domain) OnConfigApplyError(ctx ConsoleContext, errMsg, namespace, name string, opts UpdateAndDeleteOpts) error {
	c, err := d.findConfig(ctx, namespace, name)
	if err != nil {
		return errors.NewE(err)
	}

	c.SyncStatus.State = t.SyncStateErroredAtAgent
	c.SyncStatus.LastSyncedAt = opts.MessageTimestamp
	c.SyncStatus.Error = &errMsg

	_, err = d.configRepo.UpdateById(ctx, c.Id, c)
	return errors.NewE(err)
}

func (d *domain) OnConfigDeleteMessage(ctx ConsoleContext, config entities.Config) error {
	c, err := d.findConfig(ctx, config.Namespace, config.Name)
	if err != nil {
		return errors.NewE(err)
	}

	if err := d.MatchRecordVersion(config.Annotations, c.RecordVersion); err != nil {
		return d.resyncK8sResource(ctx, c.SyncStatus.Action, &c.Config, c.RecordVersion)
	}

	return d.configRepo.DeleteById(ctx, c.Id)
}

func (d *domain) OnConfigUpdateMessage(ctx ConsoleContext, config entities.Config, status types.ResourceStatus, opts UpdateAndDeleteOpts) error {
	exConfig, err := d.findConfig(ctx, config.Namespace, config.Name)
	if err != nil {
		return errors.NewE(err)
	}

	if err := d.MatchRecordVersion(config.Annotations, exConfig.RecordVersion); err != nil {
		return d.resyncK8sResource(ctx, exConfig.SyncStatus.Action, &exConfig.Config, exConfig.RecordVersion)
	}

	exConfig.CreationTimestamp = config.CreationTimestamp
	exConfig.Labels = config.Labels
	exConfig.Annotations = config.Annotations
	exConfig.Generation = config.Generation

	exConfig.Status = config.Status

	exConfig.SyncStatus.State = func() t.SyncState {
		if status == types.ResourceStatusDeleting {
			return t.SyncStateDeletingAtAgent
		}
		return t.SyncStateUpdatedAtAgent
	}()
	exConfig.SyncStatus.RecordVersion = exConfig.RecordVersion
	exConfig.SyncStatus.Error = nil
	exConfig.SyncStatus.LastSyncedAt = opts.MessageTimestamp

	_, err = d.configRepo.UpdateById(ctx, exConfig.Id, exConfig)
	return errors.NewE(err)
}

func (d *domain) ResyncConfig(ctx ConsoleContext, namespace, name string) error {
	if err := d.canMutateResourcesInWorkspace(ctx, namespace); err != nil {
		return errors.NewE(err)
	}

	c, err := d.findConfig(ctx, namespace, name)
	if err != nil {
		return errors.NewE(err)
	}

	return d.resyncK8sResource(ctx, c.SyncStatus.Action, &c.Config, c.RecordVersion)
}
