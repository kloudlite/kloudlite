package domain

import (
	"github.com/kloudlite/api/apps/console/internal/entities"
	fc "github.com/kloudlite/api/apps/console/internal/entities/field-constants"
	"github.com/kloudlite/api/common"
	"github.com/kloudlite/api/common/fields"
	"github.com/kloudlite/api/pkg/errors"
	"github.com/kloudlite/api/pkg/repos"
	t "github.com/kloudlite/api/pkg/types"
	crdsv1 "github.com/kloudlite/operator/apis/crds/v1"
	"github.com/kloudlite/operator/operators/resource-watcher/types"
)

func (d *domain) ListApps(ctx ResourceContext, search map[string]repos.MatchFilter, pq repos.CursorPagination) (*repos.PaginatedRecord[*entities.App], error) {
	if err := d.canReadResourcesInEnvironment(ctx); err != nil {
		return nil, errors.NewE(err)
	}

	filters := ctx.DBFilters()

	return d.appRepo.FindPaginated(ctx, d.appRepo.MergeMatchFilters(filters, search), pq)
}

func (d *domain) findApp(ctx ResourceContext, name string) (*entities.App, error) {
	app, err := d.appRepo.FindOne(
		ctx,
		ctx.DBFilters().Add(fields.MetadataName, name),
	)
	if err != nil {
		return nil, errors.NewE(err)
	}
	if app == nil {
		return nil, errors.Newf("no app with name (%s), found in resource context (%s)", name, ctx)
	}
	return app, nil
}

func (d *domain) GetApp(ctx ResourceContext, name string) (*entities.App, error) {
	if err := d.canReadResourcesInEnvironment(ctx); err != nil {
		return nil, errors.NewE(err)
	}
	return d.findApp(ctx, name)
}

func (d *domain) applyApp(ctx ResourceContext, app *entities.App) error {
	addTrackingId(&app.App, app.Id)
	return d.applyK8sResource(ctx, app.ProjectName, &app.App, app.RecordVersion)
}

func (d *domain) CreateApp(ctx ResourceContext, app entities.App) (*entities.App, error) {
	if err := d.canMutateResourcesInEnvironment(ctx); err != nil {
		return nil, errors.NewE(err)
	}

	env, err := d.findEnvironment(ctx.ConsoleContext, ctx.ProjectName, ctx.EnvironmentName)
	if err != nil {
		return nil, errors.NewE(err)
	}

	app.Namespace = env.Spec.TargetNamespace
	app.EnsureGVK()
	if err := d.k8sClient.ValidateObject(ctx, &app.App); err != nil {
		return nil, errors.NewE(err)
	}

	app.IncrementRecordVersion()

	app.CreatedBy = common.CreatedOrUpdatedBy{
		UserId:    ctx.UserId,
		UserName:  ctx.UserName,
		UserEmail: ctx.UserEmail,
	}
	app.LastUpdatedBy = app.CreatedBy

	app.AccountName = ctx.AccountName
	app.ProjectName = ctx.ProjectName
	app.EnvironmentName = ctx.EnvironmentName
	app.SyncStatus = t.GenSyncStatus(t.SyncActionApply, app.RecordVersion)

	if _, err := d.upsertEnvironmentResourceMapping(ctx, &app); err != nil {
		return nil, errors.NewE(err)
	}

	napp, err := d.appRepo.Create(ctx, &app)
	if err != nil {
		if d.appRepo.ErrAlreadyExists(err) {
			// TODO: better insights into error, when it is being caused by duplicated indexes
			return nil, errors.NewE(err)
		}
		return nil, err
	}

	d.resourceEventPublisher.PublishResourceEvent(ctx, entities.ResourceTypeApp, napp.Name, PublishAdd)

	if err := d.applyApp(ctx, napp); err != nil {
		return nil, errors.NewE(err)
	}

	return napp, nil
}

func (d *domain) DeleteApp(ctx ResourceContext, name string) error {
	if err := d.canMutateResourcesInEnvironment(ctx); err != nil {
		return errors.NewE(err)
	}
	uapp, err := d.appRepo.Patch(
		ctx,
		ctx.DBFilters().Add(fields.MetadataName, name),
		common.PatchForMarkDeletion(),
	)
	if err != nil {
		return errors.NewE(err)
	}
	d.resourceEventPublisher.PublishResourceEvent(ctx, entities.ResourceTypeApp, uapp.Name, PublishUpdate)
	if err := d.deleteK8sResource(ctx, uapp.ProjectName, &uapp.App); err != nil {
		if errors.Is(err, ErrNoClusterAttached) {
			return d.appRepo.DeleteById(ctx, uapp.Id)
		}
		return errors.NewE(err)
	}
	return nil
}

func (d *domain) UpdateApp(ctx ResourceContext, appIn entities.App) (*entities.App, error) {
	if err := d.canMutateResourcesInEnvironment(ctx); err != nil {
		return nil, errors.NewE(err)
	}

	appIn.Namespace = "trest"
	appIn.EnsureGVK()
	if err := d.k8sClient.ValidateObject(ctx, &appIn.App); err != nil {
		return nil, errors.NewE(err)
	}

	patchForUpdate := common.PatchForUpdate(
		ctx,
		&appIn,
		common.PatchOpts{
			XPatch: repos.Document{
				fc.AppSpec: appIn.Spec,
			},
		})

	upApp, err := d.appRepo.Patch(
		ctx,
		ctx.DBFilters().Add(fields.MetadataName, appIn.Name),
		patchForUpdate,
	)
	
	if err != nil {
		return nil, errors.NewE(err)
	}
	d.resourceEventPublisher.PublishResourceEvent(ctx, entities.ResourceTypeApp, upApp.Name, PublishUpdate)

	if err := d.applyApp(ctx, upApp); err != nil {
		return nil, errors.NewE(err)
	}

	return upApp, nil
}

// InterceptApp implements Domain.
func (d *domain) InterceptApp(ctx ResourceContext, appName string, deviceName string, intercept bool) (bool, error) {
	if err := d.canMutateResourcesInEnvironment(ctx); err != nil {
		return false, errors.NewE(err)
	}
	uApp, err := d.appRepo.Patch(ctx, ctx.DBFilters().Add(fields.MetadataName, appName), repos.Document{
		fc.AppSpecIntercept: crdsv1.Intercept{
			Enabled:  intercept,
			ToDevice: deviceName,
		},
	})
	if err != nil {
		return false, errors.NewE(err)
	}
	if err := d.applyApp(ctx, uApp); err != nil {
		return false, errors.NewE(err)
	}
	return true, nil
}

func (d *domain) OnAppUpdateMessage(ctx ResourceContext, app entities.App, status types.ResourceStatus, opts UpdateAndDeleteOpts) error {
	xApp, err := d.findApp(ctx, app.Name)
	if err != nil {
		return errors.NewE(err)
	}

	if xApp == nil {
		return errors.Newf("no apps found")
	}
	recordVersion, err := d.MatchRecordVersion(app.Annotations, xApp.RecordVersion)
	if err != nil {
		return errors.NewE(err)
	}

	uapp, err := d.appRepo.PatchById(
		ctx,
		xApp.Id,
		common.PatchForSyncFromAgent(&app, recordVersion, status, common.PatchOpts{
			MessageTimestamp: opts.MessageTimestamp,
		}))
	d.resourceEventPublisher.PublishResourceEvent(ctx, uapp.GetResourceType(), uapp.GetName(), PublishUpdate)
	return errors.NewE(err)
}

func (d *domain) OnAppDeleteMessage(ctx ResourceContext, app entities.App) error {
	err := d.appRepo.DeleteOne(
		ctx,
		ctx.DBFilters().Add(fields.MetadataName, app.Name),
	)
	if err != nil {
		return errors.NewE(err)
	}
	d.resourceEventPublisher.PublishResourceEvent(ctx, entities.ResourceTypeApp, app.Name, PublishDelete)
	return nil
}

func (d *domain) OnAppApplyError(ctx ResourceContext, errMsg string, name string, opts UpdateAndDeleteOpts) error {
	uapp, err := d.appRepo.Patch(
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

func (d *domain) ResyncApp(ctx ResourceContext, name string) error {
	if err := d.canMutateResourcesInEnvironment(ctx); err != nil {
		return errors.NewE(err)
	}
	a, err := d.findApp(ctx, name)
	if err != nil {
		return errors.NewE(err)
	}
	return d.resyncK8sResource(ctx, a.ProjectName, a.SyncStatus.Action, &a.App, a.RecordVersion)
}
