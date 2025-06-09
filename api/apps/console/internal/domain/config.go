package domain

import (
	"fmt"

	"github.com/kloudlite/api/apps/console/internal/entities"
	fc "github.com/kloudlite/api/apps/console/internal/entities/field-constants"
	"github.com/kloudlite/api/common"
	"github.com/kloudlite/api/common/fields"
	"github.com/kloudlite/api/pkg/errors"
	fn "github.com/kloudlite/api/pkg/functions"
	"github.com/kloudlite/api/pkg/repos"
	t "github.com/kloudlite/api/pkg/types"
	"github.com/kloudlite/operator/operators/resource-watcher/types"
)

func (d *domain) ListConfigs(ctx ResourceContext, search map[string]repos.MatchFilter, pq repos.CursorPagination) (*repos.PaginatedRecord[*entities.Config], error) {
	if err := d.canReadResourcesInEnvironment(ctx); err != nil {
		return nil, errors.NewE(err)
	}
	filter := ctx.DBFilters()
	return d.configRepo.FindPaginated(ctx, d.configRepo.MergeMatchFilters(filter, search), pq)
}

func (d *domain) findConfig(ctx ResourceContext, name string) (*entities.Config, error) {
	cfg, err := d.configRepo.FindOne(ctx, ctx.DBFilters().Add(fields.MetadataName, name))
	if err != nil {
		return nil, errors.NewE(err)
	}
	if cfg == nil {
		return nil, errors.Newf("no config with name (%q)", name)
	}
	return cfg, nil
}

func (d *domain) GetConfig(ctx ResourceContext, name string) (*entities.Config, error) {
	if err := d.canReadResourcesInEnvironment(ctx); err != nil {
		return nil, errors.NewE(err)
	}
	return d.findConfig(ctx, name)
}

// GetConfigEntries implements Domain.
func (d *domain) GetConfigEntries(ctx ResourceContext, keyrefs []ConfigKeyRef) ([]*ConfigKeyValueRef, error) {
	filters := ctx.DBFilters()
	names := make([]any, 0, len(keyrefs))
	for i := range keyrefs {
		names = append(names, keyrefs[i].ConfigName)
	}
	filters = d.configRepo.MergeMatchFilters(filters, map[string]repos.MatchFilter{
		fields.MetadataName: {
			MatchType: repos.MatchTypeArray,
			Array:     names,
		},
	})
	configs, err := d.configRepo.Find(ctx, repos.Query{Filter: filters})
	if err != nil {
		return nil, errors.NewE(err)
	}

	results := make([]*ConfigKeyValueRef, 0, len(configs))

	data := make(map[string]map[string]string)

	for i := range configs {
		data[configs[i].Name] = configs[i].Data
	}

	for i := range keyrefs {
		results = append(results, &ConfigKeyValueRef{
			ConfigName: keyrefs[i].ConfigName,
			Key:        keyrefs[i].Key,
			Value:      data[keyrefs[i].ConfigName][keyrefs[i].Key],
		})
	}

	return results, nil
}

func (d *domain) CreateConfig(ctx ResourceContext, config entities.Config) (*entities.Config, error) {
	if err := d.canMutateResourcesInEnvironment(ctx); err != nil {
		return nil, errors.NewE(err)
	}

	config.SetGroupVersionKind(fn.GVK("v1", "ConfigMap"))

	var err error
	config.Namespace, err = d.envTargetNamespace(ctx.ConsoleContext, ctx.EnvironmentName)
	if err != nil {
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
	config.EnvironmentName = ctx.EnvironmentName

	if config.Annotations == nil {
		config.Annotations = make(map[string]string, len(types.ConfigWatchingAnnotation))
	}
	for k, v := range types.ConfigWatchingAnnotation {
		config.Annotations[k] = v
	}

	return d.createAndApplyConfig(ctx, &config)
}

func (d *domain) createAndApplyConfig(ctx ResourceContext, config *entities.Config) (*entities.Config, error) {
	config.SyncStatus = t.GenSyncStatus(t.SyncActionApply, 0)

	if _, err := d.upsertEnvironmentResourceMapping(ctx, config); err != nil {
		return nil, errors.NewE(err)
	}

	cfg, err := d.configRepo.Create(ctx, config)
	if err != nil {
		if d.configRepo.ErrAlreadyExists(err) {
			// TODO: better insights into error, when it is being caused by duplicated indexes
			return nil, errors.NewE(err)
		}
		return nil, errors.NewE(err)
	}

	if err := d.applyK8sResource(ctx, ctx.EnvironmentName, &cfg.ConfigMap, cfg.RecordVersion); err != nil {
		return nil, errors.NewE(err)
	}

	return cfg, nil
}

func (d *domain) UpdateConfig(ctx ResourceContext, config entities.Config) (*entities.Config, error) {
	if err := d.canMutateResourcesInEnvironment(ctx); err != nil {
		return nil, errors.NewE(err)
	}

	config.SetGroupVersionKind(fn.GVK("v1", "ConfigMap"))

	if config.Annotations == nil {
		config.Annotations = make(map[string]string, len(types.ConfigWatchingAnnotation))
	}

	for k, v := range types.ConfigWatchingAnnotation {
		config.Annotations[k] = v
	}

	upConfig, err := d.configRepo.Patch(ctx, ctx.DBFilters().Add(fields.MetadataName, config.Name),
		common.PatchForUpdate(ctx, &config, common.PatchOpts{XPatch: repos.Document{
			fc.ConfigData: config.Data,
		}}),
	)
	if err != nil {
		return nil, errors.NewE(err)
	}

	d.resourceEventPublisher.PublishResourceEvent(ctx, entities.ResourceTypeConfig, upConfig.Name, PublishUpdate)

	if err := d.applyK8sResource(ctx, upConfig.EnvironmentName, &upConfig.ConfigMap, upConfig.RecordVersion); err != nil {
		return nil, errors.NewE(err)
	}

	return upConfig, nil
}

func (d *domain) applyConfigToK8s(ctx ResourceContext, config *entities.Config) error {
	return d.applyK8sResource(ctx, ctx.EnvironmentName, &config.ConfigMap, config.RecordVersion)
}

func (d *domain) deleteConfigFromK8s(ctx ResourceContext, config *entities.Config) error {
	return d.deleteK8sResource(ctx, ctx.EnvironmentName, &config.ConfigMap)
}

func (d *domain) DeleteConfig(ctx ResourceContext, name string) error {
	if err := d.canMutateResourcesInEnvironment(ctx); err != nil {
		return errors.NewE(err)
	}

	uc, err := d.configRepo.Patch(
		ctx,
		ctx.DBFilters().Add(fields.MetadataName, name),
		common.PatchForMarkDeletion(),
	)
	if err != nil {
		return errors.NewE(err)
	}
	d.resourceEventPublisher.PublishResourceEvent(ctx, entities.ResourceTypeConfig, uc.Name, PublishUpdate)

	if err := d.deleteConfigFromK8s(ctx, uc); err != nil {
		if errors.Is(err, ErrNoClusterAttached) {
			return d.configRepo.DeleteById(ctx, uc.Id)
		}
		return errors.NewE(err)
	}
	return nil
}

func (d *domain) OnConfigApplyError(ctx ResourceContext, errMsg, name string, opts UpdateAndDeleteOpts) error {
	uc, err := d.configRepo.Patch(
		ctx,
		ctx.DBFilters().Add(fields.MetadataName, name),
		common.PatchForErrorFromAgent(
			errMsg,
			common.PatchOpts{
				MessageTimestamp: opts.MessageTimestamp,
			},
		),
	)
	if err != nil {
		return errors.NewE(err)
	}

	d.resourceEventPublisher.PublishResourceEvent(ctx, entities.ResourceTypeConfig, uc.Name, PublishDelete)
	return nil
}

func (d *domain) OnConfigDeleteMessage(ctx ResourceContext, config entities.Config) error {
	err := d.configRepo.DeleteOne(ctx, ctx.DBFilters().Add(fields.MetadataName, config.Name).Add(fields.MetadataNamespace, config.Namespace))
	if err != nil {
		return errors.NewE(err)
	}
	d.resourceEventPublisher.PublishResourceEvent(ctx, entities.ResourceTypeConfig, config.Name, PublishDelete)
	return nil
}

func (d *domain) OnConfigUpdateMessage(ctx ResourceContext, configIn entities.Config, status types.ResourceStatus, opts UpdateAndDeleteOpts) error {
	if v, ok := configIn.GetLabels()["app.kubernetes.io/managed-by"]; ok && v == "Helm" {
		// INFO: configmap created with Helm, we should just upsert it

		ctx.DBFilters().Add(fc.MetadataName, configIn.Name).Add(fc.MetadataNamespace, configIn.Namespace)
		// meta.helm.sh/release-name: playground
		// meta.helm.sh/release-namespace: env-nxt17-env-1

		configIn.AccountName = ctx.AccountName
		configIn.EnvironmentName = ctx.EnvironmentName
		configIn.CreatedByHelm = fn.New(fmt.Sprintf("%s/%s", configIn.GetAnnotations()["meta.helm.sh/release-namespace"], configIn.GetAnnotations()["meta.helm.sh/release-name"]))
		configIn.ResourceMetadata = common.ResourceMetadata{
			DisplayName:   configIn.Name,
			CreatedBy:     common.CreatedOrUpdatedByResourceSync,
			LastUpdatedBy: common.CreatedOrUpdatedByResourceSync,
		}
		configIn.SyncStatus = t.SyncStatus{
			LastSyncedAt:  opts.MessageTimestamp,
			Action:        t.SyncActionApply,
			RecordVersion: 0,
			State:         t.SyncStateAppliedAtAgent,
			Error:         nil,
		}

		_, err := d.configRepo.Upsert(ctx, repos.Filter{
			fc.MetadataName:      configIn.Name,
			fc.MetadataNamespace: configIn.Namespace,
			fc.EnvironmentName:   ctx.EnvironmentName,
		}, &configIn)
		if err != nil {
			return errors.NewE(err)
		}

		return nil
	}

	xconfig, err := d.findConfig(ctx, configIn.Name)
	if err != nil {
		return errors.NewE(err)
	}

	if xconfig == nil {
		return errors.Newf("no config found")
	}

	recordVersion, err := d.MatchRecordVersion(configIn.Annotations, xconfig.RecordVersion)
	if err != nil {
		return d.resyncK8sResource(ctx, xconfig.EnvironmentName, xconfig.SyncStatus.Action, &xconfig.ConfigMap, xconfig.RecordVersion)
	}

	uc, err := d.configRepo.PatchById(ctx, xconfig.Id, common.PatchForSyncFromAgent(&configIn, recordVersion, status, common.PatchOpts{
		MessageTimestamp: opts.MessageTimestamp,
	}))
	d.resourceEventPublisher.PublishResourceEvent(ctx, uc.GetResourceType(), uc.GetName(), PublishUpdate)
	return errors.NewE(err)
}

func (d *domain) ResyncConfig(ctx ResourceContext, name string) error {
	if err := d.canMutateResourcesInEnvironment(ctx); err != nil {
		return errors.NewE(err)
	}

	cfg, err := d.findConfig(ctx, name)
	if err != nil {
		return errors.NewE(err)
	}

	return d.resyncK8sResource(ctx, cfg.EnvironmentName, cfg.SyncStatus.Action, &cfg.ConfigMap, cfg.RecordVersion)
}
