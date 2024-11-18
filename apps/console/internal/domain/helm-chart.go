package domain

import (
	"github.com/kloudlite/api/apps/console/internal/entities"
	fc "github.com/kloudlite/api/apps/console/internal/entities/field-constants"
	"github.com/kloudlite/api/common"
	"github.com/kloudlite/api/common/fields"
	"github.com/kloudlite/api/pkg/errors"
	"github.com/kloudlite/api/pkg/repos"
	t "github.com/kloudlite/api/pkg/types"
	"github.com/kloudlite/operator/operators/resource-watcher/types"
)

func (d *domain) createAndApplyHelmChart(ctx ResourceContext, helmc *entities.HelmChart) (*entities.HelmChart, error) {
	if _, err := d.upsertEnvironmentResourceMapping(ctx, helmc); err != nil {
		return nil, errors.NewE(err)
	}

	nhelmc, err := d.helmChartRepo.Create(ctx, helmc)
	if err != nil {
		if d.helmChartRepo.ErrAlreadyExists(err) {
			// TODO: better insights into error, when it is being caused by duplicated indexes
			return nil, errors.NewE(err)
		}
		return nil, err
	}

	d.resourceEventPublisher.PublishResourceEvent(ctx, entities.ResourceTypeHelmChart, nhelmc.Name, PublishAdd)

	if err := d.applyHelmChart(ctx, nhelmc); err != nil {
		return nil, errors.NewE(err)
	}
	return nhelmc, nil
}

func (d *domain) findHelmCharts(ctx ResourceContext, name string) (*entities.HelmChart, error) {
	helmc, err := d.helmChartRepo.FindOne(
		ctx,
		ctx.DBFilters().Add(fields.MetadataName, name),
	)
	if err != nil {
		return nil, errors.NewE(err)
	}
	if helmc == nil {
		return nil, errors.Newf("no helm chart with name (%s), found in resource context (%s)", name, ctx)
	}
	return helmc, nil
}

func (d *domain) ListHelmCharts(ctx ResourceContext, search map[string]repos.MatchFilter, pq repos.CursorPagination) (*repos.PaginatedRecord[*entities.HelmChart], error) {
	if err := d.canReadResourcesInEnvironment(ctx); err != nil {
		return nil, errors.NewE(err)
	}

	filters := ctx.DBFilters()

	return d.helmChartRepo.FindPaginated(ctx, d.helmChartRepo.MergeMatchFilters(filters, search), pq)
}

func (d *domain) GetHelmChart(ctx ResourceContext, name string) (*entities.HelmChart, error) {
	if err := d.canReadResourcesInEnvironment(ctx); err != nil {
		return nil, errors.NewE(err)
	}
	return d.findHelmCharts(ctx, name)
}

func (d *domain) applyHelmChart(ctx ResourceContext, helmc *entities.HelmChart) error {
	addTrackingId(&helmc.HelmChart, helmc.Id)
	return d.applyK8sResource(ctx, helmc.EnvironmentName, &helmc.HelmChart, helmc.RecordVersion)
}

func (d *domain) CreateHelmChart(ctx ResourceContext, helmc entities.HelmChart) (*entities.HelmChart, error) {
	if err := d.canMutateResourcesInEnvironment(ctx); err != nil {
		return nil, errors.NewE(err)
	}

	env, err := d.findEnvironment(ctx.ConsoleContext, ctx.EnvironmentName)
	if err != nil {
		return nil, errors.NewE(err)
	}

	helmc.Namespace = env.Spec.TargetNamespace
	helmc.EnsureGVK()
	if err := d.k8sClient.ValidateObject(ctx, &helmc.HelmChart); err != nil {
		return nil, errors.NewE(err)
	}

	helmc.IncrementRecordVersion()

	helmc.CreatedBy = common.CreatedOrUpdatedBy{
		UserId:    ctx.UserId,
		UserName:  ctx.UserName,
		UserEmail: ctx.UserEmail,
	}
	helmc.LastUpdatedBy = helmc.CreatedBy

	helmc.AccountName = ctx.AccountName
	helmc.EnvironmentName = ctx.EnvironmentName
	helmc.SyncStatus = t.GenSyncStatus(t.SyncActionApply, helmc.RecordVersion)

	return d.createAndApplyHelmChart(ctx, &helmc)
}

func (d *domain) DeleteHelmChart(ctx ResourceContext, name string) error {
	if err := d.canMutateResourcesInEnvironment(ctx); err != nil {
		return errors.NewE(err)
	}

	uhelmc, err := d.helmChartRepo.Patch(
		ctx,
		ctx.DBFilters().Add(fields.MetadataName, name),
		common.PatchForMarkDeletion(),
	)
	if err != nil {
		return errors.NewE(err)
	}

	d.resourceEventPublisher.PublishResourceEvent(ctx, entities.ResourceTypeHelmChart, uhelmc.Name, PublishUpdate)
	if err := d.deleteK8sResource(ctx, uhelmc.EnvironmentName, &uhelmc.HelmChart); err != nil {
		if errors.Is(err, ErrNoClusterAttached) {
			return d.helmChartRepo.DeleteById(ctx, uhelmc.Id)
		}
		return errors.NewE(err)
	}
	return nil
}

func (d *domain) UpdateHelmChart(ctx ResourceContext, helmcIn entities.HelmChart) (*entities.HelmChart, error) {
	if err := d.canMutateResourcesInEnvironment(ctx); err != nil {
		return nil, errors.NewE(err)
	}

	helmcIn.Namespace = "trest"
	helmcIn.EnsureGVK()
	if err := d.k8sClient.ValidateObject(ctx, &helmcIn.HelmChart); err != nil {
		return nil, errors.NewE(err)
	}

	xhelmc, err := d.helmChartRepo.FindOne(ctx, ctx.DBFilters().Add(fields.MetadataName, helmcIn.Name))
	if err != nil {
		return nil, errors.NewE(err)
	}

	if xhelmc == nil {
		return nil, errors.Newf("helm-chart does not exist")
	}

	patchDoc := repos.Document{
		fc.HelmChartSpecChartVersion: helmcIn.Spec.ChartVersion,
		fc.HelmChartSpecValues:       helmcIn.Spec.Values,
	}

	patchForUpdate := common.PatchForUpdate(
		ctx,
		&helmcIn,
		common.PatchOpts{
			XPatch: patchDoc,
		})

	upHelmc, err := d.helmChartRepo.Patch(ctx, ctx.DBFilters().Add(fields.MetadataName, helmcIn.Name), patchForUpdate)
	if err != nil {
		return nil, errors.NewE(err)
	}

	d.resourceEventPublisher.PublishResourceEvent(ctx, entities.ResourceTypeHelmChart, upHelmc.Name, PublishUpdate)

	if err := d.applyHelmChart(ctx, upHelmc); err != nil {
		return nil, errors.NewE(err)
	}

	return upHelmc, nil
}

func (d *domain) OnHelmChartUpdateMessage(ctx ResourceContext, app entities.HelmChart, status types.ResourceStatus, opts UpdateAndDeleteOpts) error {
	xApp, err := d.findHelmCharts(ctx, app.Name)
	if err != nil {
		return errors.NewE(err)
	}

	if xApp == nil {
		return errors.Newf("no helm charts found")
	}

	recordVersion, err := d.MatchRecordVersion(app.Annotations, xApp.RecordVersion)
	if err != nil {
		return errors.NewE(err)
	}

	uhelmc, err := d.helmChartRepo.PatchById(ctx, xApp.Id,
		common.PatchForSyncFromAgent(&app, recordVersion, status, common.PatchOpts{
			MessageTimestamp: opts.MessageTimestamp,
		}))
	d.resourceEventPublisher.PublishResourceEvent(ctx, uhelmc.GetResourceType(), uhelmc.GetName(), PublishUpdate)
	return errors.NewE(err)
}

func (d *domain) OnHelmChartDeleteMessage(ctx ResourceContext, app entities.HelmChart) error {
	err := d.helmChartRepo.DeleteOne(
		ctx,
		ctx.DBFilters().Add(fields.MetadataName, app.Name),
	)
	if err != nil {
		return errors.NewE(err)
	}
	d.resourceEventPublisher.PublishResourceEvent(ctx, entities.ResourceTypeApp, app.Name, PublishDelete)
	return nil
}

func (d *domain) OnHelmChartApplyError(ctx ResourceContext, errMsg string, name string, opts UpdateAndDeleteOpts) error {
	uapp, err := d.helmChartRepo.Patch(
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

	d.resourceEventPublisher.PublishResourceEvent(ctx, entities.ResourceTypeApp, uapp.Name, PublishDelete)
	return errors.NewE(err)
}

func (d *domain) ResyncHelmChart(ctx ResourceContext, name string) error {
	if err := d.canMutateResourcesInEnvironment(ctx); err != nil {
		return errors.NewE(err)
	}
	a, err := d.findHelmCharts(ctx, name)
	if err != nil {
		return errors.NewE(err)
	}
	return d.resyncK8sResource(ctx, a.EnvironmentName, a.SyncStatus.Action, &a.HelmChart, a.RecordVersion)
}
