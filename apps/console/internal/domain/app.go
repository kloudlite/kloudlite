package domain

import (
	"github.com/kloudlite/api/apps/console/internal/entities"
	"github.com/kloudlite/api/common"
	"github.com/kloudlite/api/pkg/errors"
	"github.com/kloudlite/api/pkg/repos"
	t "github.com/kloudlite/api/pkg/types"
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
	filters.Add("metadata.name", name)

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

	if _, err := d.upsertResourceMapping(ctx, &app); err != nil {
		return nil, errors.NewE(err)
	}

	nApp, err := d.appRepo.Create(ctx, &app)
	if err != nil {
		if d.appRepo.ErrAlreadyExists(err) {
			// TODO: better insights into error, when it is being caused by duplicated indexes
			return nil, errors.NewE(err)
		}
		return nil, err
	}

	d.resourceEventPublisher.PublishAppEvent(&app, PublishAdd)

	if err := d.applyK8sResource(ctx, app.ProjectName, &nApp.App, nApp.RecordVersion); err != nil {
		return nil, errors.NewE(err)
	}

	return nApp, nil
}

func (d *domain) DeleteApp(ctx ResourceContext, name string) error {
	if err := d.canMutateResourcesInEnvironment(ctx); err != nil {
		return errors.NewE(err)
	}

	app, err := d.findApp(ctx, name)
	if err != nil {
		return errors.NewE(err)
	}

	app.SyncStatus = t.GenSyncStatus(t.SyncActionDelete, app.RecordVersion)

	if _, err := d.appRepo.UpdateById(ctx, app.Id, app); err != nil {
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

func (d *domain) UpdateApp(ctx ResourceContext, app entities.App) (*entities.App, error) {
	if err := d.canMutateResourcesInEnvironment(ctx); err != nil {
		return nil, errors.NewE(err)
	}

	app.EnsureGVK()
	if err := d.k8sClient.ValidateObject(ctx, &app.App); err != nil {
		return nil, errors.NewE(err)
	}

	exApp, err := d.findApp(ctx, app.Name)
	if err != nil {
		return nil, errors.NewE(err)
	}

	exApp.IncrementRecordVersion()

	exApp.LastUpdatedBy = common.CreatedOrUpdatedBy{
		UserId:    ctx.UserId,
		UserName:  ctx.UserName,
		UserEmail: ctx.UserEmail,
	}

	exApp.DisplayName = app.DisplayName

	exApp.Labels = app.Labels
	exApp.Annotations = app.Annotations
	exApp.Spec = app.Spec
	exApp.SyncStatus = t.GenSyncStatus(t.SyncActionApply, exApp.RecordVersion)

	upApp, err := d.appRepo.UpdateById(ctx, exApp.Id, exApp)
	if err != nil {
		return nil, errors.NewE(err)
	}
	d.resourceEventPublisher.PublishAppEvent(upApp, PublishUpdate)

	if err := d.applyK8sResource(ctx, ctx.AccountName, &upApp.App, upApp.RecordVersion); err != nil {
		return nil, errors.NewE(err)
	}

	return upApp, nil
}

func (d *domain) OnAppUpdateMessage(ctx ResourceContext, app entities.App, status types.ResourceStatus, opts UpdateAndDeleteOpts) error {
	xApp, err := d.findApp(ctx, app.Name)
	if err != nil {
		return errors.NewE(err)
	}

	if err := d.MatchRecordVersion(app.Annotations, xApp.RecordVersion); err != nil {
		return errors.NewE(err)
	}

	xApp.CreationTimestamp = app.CreationTimestamp
	xApp.Labels = app.Labels
	xApp.Annotations = app.Annotations
	xApp.Generation = app.Generation

	xApp.Status = app.Status

	xApp.SyncStatus.State = func() t.SyncState {
		if status == types.ResourceStatusDeleting {
			return t.SyncStateDeletingAtAgent
		}
		return t.SyncStateUpdatedAtAgent
	}()

	xApp.SyncStatus.RecordVersion = xApp.RecordVersion
	xApp.SyncStatus.Error = nil
	xApp.SyncStatus.LastSyncedAt = opts.MessageTimestamp

	_, err = d.appRepo.UpdateById(ctx, xApp.Id, xApp)
	d.resourceEventPublisher.PublishAppEvent(xApp, PublishUpdate)
	return errors.NewE(err)
}

func (d *domain) OnAppDeleteMessage(ctx ResourceContext, app entities.App) error {
	a, err := d.findApp(ctx, app.Name)
	if err != nil {
		return errors.NewE(err)
	}

	if err := d.MatchRecordVersion(app.Annotations, a.RecordVersion); err != nil {
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

	app.SyncStatus.State = t.SyncStateErroredAtAgent

	app.SyncStatus.LastSyncedAt = opts.MessageTimestamp
	app.SyncStatus.Error = &errMsg

	_, err = d.appRepo.UpdateById(ctx, app.Id, app)
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
