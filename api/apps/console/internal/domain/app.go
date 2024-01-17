package domain

import (
	"time"

	"github.com/kloudlite/api/apps/console/internal/entities"
	fc "github.com/kloudlite/api/apps/console/internal/entities/field-constants"
	"github.com/kloudlite/api/common"
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
	filters := ctx.DBFilters()
	filters.Add(fc.MetadataName, name)

	app, err := d.appRepo.FindOne(ctx, filters)
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

	d.resourceEventPublisher.PublishAppEvent(napp, PublishAdd)

	if err := d.applyApp(ctx, napp); err != nil {
		return nil, errors.NewE(err)
	}

	return napp, nil
}

func (d *domain) DeleteApp(ctx ResourceContext, name string) error {
	if err := d.canMutateResourcesInEnvironment(ctx); err != nil {
		return errors.NewE(err)
	}

	app, err := d.findApp(ctx, name)
	if err != nil {
		return errors.NewE(err)
	}

	if app == nil {
		return errors.Newf("apps not found")
	}

	if _, err := d.appRepo.PatchById(ctx, app.Id, repos.Document{
		fc.MarkedForDeletion:         true,
		fc.SyncStatusSyncScheduledAt: time.Now(),
		fc.SyncStatusAction:          t.SyncActionDelete,
		fc.SyncStatusState:           t.SyncStateIdle,
	}); err != nil {
		return errors.NewE(err)
	}
	d.resourceEventPublisher.PublishAppEvent(app, PublishUpdate)

	if err := d.deleteK8sResource(ctx, app.ProjectName, &app.App); err != nil {
		if errors.Is(err, ErrNoClusterAttached) {
			return d.appRepo.DeleteById(ctx, app.Id)
		}
		return errors.NewE(err)
	}
	return nil
}

func (d *domain) UpdateApp(ctx ResourceContext, appIn entities.App) (*entities.App, error) {
	if err := d.canMutateResourcesInEnvironment(ctx); err != nil {
		return nil, errors.NewE(err)
	}

	xapp, err := d.findApp(ctx, appIn.Name)
	if err != nil {
		return nil, errors.NewE(err)
	}

	appIn.Namespace = xapp.Namespace
	appIn.EnsureGVK()
	if err := d.k8sClient.ValidateObject(ctx, &appIn.App); err != nil {
		return nil, errors.NewE(err)
	}

	upApp, err := d.appRepo.PatchById(ctx, xapp.Id, repos.Document{
		fc.MetadataLabels:      appIn.Labels,
		fc.MetadataAnnotations: appIn.Annotations,

		fc.RecordVersion: xapp.RecordVersion + 1,
		fc.DisplayName:   appIn.DisplayName,
		fc.LastUpdatedBy: common.CreatedOrUpdatedBy{
			UserId:    ctx.UserId,
			UserName:  ctx.UserName,
			UserEmail: ctx.UserEmail,
		},

		fc.AppSpec: appIn.Spec,

		fc.SyncStatusSyncScheduledAt: time.Now(),
		fc.SyncStatusState:           t.SyncStateInQueue,
		fc.SyncStatusAction:          t.SyncActionApply,
	})
	if err != nil {
		return nil, errors.NewE(err)
	}
	d.resourceEventPublisher.PublishAppEvent(upApp, PublishUpdate)

	if err := d.applyApp(ctx, upApp); err != nil {
		return nil, errors.NewE(err)
	}

	return upApp, nil
}

// InterceptApp implements Domain.
func (d *domain) InterceptApp(ctx ResourceContext, appName string, deviceName string, intercept bool) (bool, error) {
	app, err := d.findApp(ctx, appName)
	if err != nil {
		return false, err
	}

	if app == nil {
		return false, errors.Newf("no aps found")
	}

	// intercepted := app.Spec.Intercept != nil && app.Spec.Intercept.Enabled

	// if intercepted && app.Spec.Intercept.ToDevice != deviceName {
	// 	return false, errors.Newf("device (%s) is already intercepting app (%s)", app.Spec.Intercept.ToDevice, appName)
	// }

	uApp, err := d.appRepo.PatchById(ctx, app.Id, repos.Document{
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

	if err := d.MatchRecordVersion(app.Annotations, xApp.RecordVersion); err != nil {
		return errors.NewE(err)
	}

	uapp, err := d.appRepo.PatchById(ctx, xApp.Id, repos.Document{
		fc.MetadataCreationTimestamp: app.CreationTimestamp,
		fc.MetadataLabels:            app.Labels,
		fc.MetadataAnnotations:       app.Annotations,
		fc.MetadataGeneration:        app.Generation,

		fc.Status: app.Status,

		fc.SyncStatusState: func() t.SyncState {
			if status == types.ResourceStatusDeleting {
				return t.SyncStateDeletingAtAgent
			}
			return t.SyncStateUpdatedAtAgent
		}(),
		fc.SyncStatusRecordVersion: xApp.RecordVersion,
		fc.SyncStatusLastSyncedAt:  opts.MessageTimestamp,
		fc.SyncStatusError:         nil,
	})
	d.resourceEventPublisher.PublishAppEvent(uapp, PublishUpdate)
	return errors.NewE(err)
}

func (d *domain) OnAppDeleteMessage(ctx ResourceContext, app entities.App) error {
	a, err := d.findApp(ctx, app.Name)
	if err != nil {
		return errors.NewE(err)
	}

	err = d.appRepo.DeleteById(ctx, a.Id)
	if err != nil {
		return errors.NewE(err)
	}
	d.resourceEventPublisher.PublishAppEvent(a, PublishDelete)

	return nil
}

func (d *domain) OnAppApplyError(ctx ResourceContext, errMsg string, name string, opts UpdateAndDeleteOpts) error {
	app, err := d.findApp(ctx, name)
	if err != nil {
		return errors.NewE(err)
	}

	uapp, err := d.appRepo.PatchById(ctx, app.Id, repos.Document{
		fc.SyncStatusState:        t.SyncStateErroredAtAgent,
		fc.SyncStatusLastSyncedAt: opts.MessageTimestamp,
		fc.SyncStatusError:        errMsg,
	})
	if err != nil {
		return errors.NewE(err)
	}
	d.resourceEventPublisher.PublishAppEvent(uapp, PublishDelete)
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
