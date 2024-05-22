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

func (d *domain) ListExternalApps(ctx ResourceContext, search map[string]repos.MatchFilter, pq repos.CursorPagination) (*repos.PaginatedRecord[*entities.ExternalApp], error) {
	if err := d.canReadResourcesInEnvironment(ctx); err != nil {
		return nil, errors.NewE(err)
	}

	filters := ctx.DBFilters()

	return d.externalAppRepo.FindPaginated(ctx, d.externalAppRepo.MergeMatchFilters(filters, search), pq)
}

func (d *domain) findExternalApp(ctx ResourceContext, name string) (*entities.ExternalApp, error) {
	app, err := d.externalAppRepo.FindOne(
		ctx,
		ctx.DBFilters().Add(fields.MetadataName, name),
	)
	if err != nil {
		return nil, errors.NewE(err)
	}
	if app == nil {
		return nil, errors.Newf("no external app with name (%s), found in resource context (%s)", name, ctx)
	}
	return app, nil
}

func (d *domain) GetExternalApp(ctx ResourceContext, name string) (*entities.ExternalApp, error) {
	if err := d.canReadResourcesInEnvironment(ctx); err != nil {
		return nil, errors.NewE(err)
	}
	return d.findExternalApp(ctx, name)
}

func (d *domain) applyExternalApp(ctx ResourceContext, externalApp *entities.ExternalApp) error {
	addTrackingId(&externalApp.ExternalApp, externalApp.Id)
	return d.applyK8sResource(ctx, externalApp.EnvironmentName, &externalApp.ExternalApp, externalApp.RecordVersion)
}

func (d *domain) CreateExternalApp(ctx ResourceContext, externalApp entities.ExternalApp) (*entities.ExternalApp, error) {
	if err := d.canMutateResourcesInEnvironment(ctx); err != nil {
		return nil, errors.NewE(err)
	}

	env, err := d.findEnvironment(ctx.ConsoleContext, ctx.EnvironmentName)
	if err != nil {
		return nil, errors.NewE(err)
	}

	externalApp.Namespace = env.Spec.TargetNamespace
	externalApp.EnsureGVK()
	if err := d.k8sClient.ValidateObject(ctx, &externalApp.ExternalApp); err != nil {
		return nil, errors.NewE(err)
	}

	externalApp.IncrementRecordVersion()

	externalApp.CreatedBy = common.CreatedOrUpdatedBy{
		UserId:    ctx.UserId,
		UserName:  ctx.UserName,
		UserEmail: ctx.UserEmail,
	}
	externalApp.LastUpdatedBy = externalApp.CreatedBy

	externalApp.AccountName = ctx.AccountName
	externalApp.EnvironmentName = ctx.EnvironmentName
	externalApp.SyncStatus = t.GenSyncStatus(t.SyncActionApply, externalApp.RecordVersion)

	return d.createAndApplyExternalApp(ctx, &externalApp)
}

func (d *domain) createAndApplyExternalApp(ctx ResourceContext, externalApp *entities.ExternalApp) (*entities.ExternalApp, error) {
	if _, err := d.upsertEnvironmentResourceMapping(ctx, externalApp); err != nil {
		return nil, errors.NewE(err)
	}

	nextapp, err := d.externalAppRepo.Create(ctx, externalApp)
	if err != nil {
		if d.externalAppRepo.ErrAlreadyExists(err) {
			// TODO: better insights into error, when it is being caused by duplicated indexes
			return nil, errors.NewE(err)
		}
		return nil, err
	}

	d.resourceEventPublisher.PublishResourceEvent(ctx, entities.ResourceTypeExternalApp, nextapp.Name, PublishAdd)

	if err := d.applyExternalApp(ctx, nextapp); err != nil {
		return nil, errors.NewE(err)
	}
	return nextapp, nil
}

func (d *domain) UpdateExternalApp(ctx ResourceContext, externalAppIn entities.ExternalApp) (*entities.ExternalApp, error) {
	if err := d.canMutateResourcesInEnvironment(ctx); err != nil {
		return nil, errors.NewE(err)
	}

	externalAppIn.Namespace = "trest"
	externalAppIn.EnsureGVK()
	if err := d.k8sClient.ValidateObject(ctx, &externalAppIn.ExternalApp); err != nil {
		return nil, errors.NewE(err)
	}

	patchForUpdate := common.PatchForUpdate(
		ctx,
		&externalAppIn,
		common.PatchOpts{
			XPatch: repos.Document{
				fc.ExternalAppSpec: externalAppIn.Spec,
			},
		},
	)

	upExternalApp, err := d.externalAppRepo.Patch(
		ctx,
		ctx.DBFilters().Add(fields.MetadataName, externalAppIn.Name),
		patchForUpdate,
	)
	if err != nil {
		return nil, errors.NewE(err)
	}
	d.resourceEventPublisher.PublishResourceEvent(ctx, entities.ResourceTypeExternalApp, upExternalApp.Name, PublishUpdate)

	if err := d.applyExternalApp(ctx, upExternalApp); err != nil {
		return nil, errors.NewE(err)
	}

	return upExternalApp, nil
}

func (d *domain) DeleteExternalApp(ctx ResourceContext, name string) error {
	if err := d.canMutateResourcesInEnvironment(ctx); err != nil {
		return errors.NewE(err)
	}
	uapp, err := d.externalAppRepo.Patch(
		ctx,
		ctx.DBFilters().Add(fields.MetadataName, name),
		common.PatchForMarkDeletion(),
	)
	if err != nil {
		return errors.NewE(err)
	}
	d.resourceEventPublisher.PublishResourceEvent(ctx, entities.ResourceTypeExternalApp, uapp.Name, PublishUpdate)
	if err := d.deleteK8sResource(ctx, uapp.EnvironmentName, &uapp.ExternalApp); err != nil {
		if errors.Is(err, ErrNoClusterAttached) {
			return d.externalAppRepo.DeleteById(ctx, uapp.Id)
		}
		return errors.NewE(err)
	}
	return nil
}

func (d *domain) InterceptExternalApp(ctx ResourceContext, externalAppName string, deviceName string, intercept bool, portMappings []crdsv1.AppInterceptPortMappings) (bool, error) {
	if err := d.canMutateResourcesInEnvironment(ctx); err != nil {
		return false, errors.NewE(err)
	}

	patch := repos.Document{
		fc.ExternalAppSpecInterceptEnabled:  intercept,
		fc.ExternalAppSpecInterceptToDevice: deviceName,
	}

	if portMappings != nil {
		patch[fc.ExternalAppSpecInterceptPortMappings] = portMappings
	}

	uExternalApp, err := d.externalAppRepo.Patch(ctx, ctx.DBFilters().Add(fields.MetadataName, externalAppName), patch)
	if err != nil {
		return false, errors.NewE(err)
	}
	if err := d.applyExternalApp(ctx, uExternalApp); err != nil {
		return false, errors.NewE(err)
	}
	return true, nil
}

func (d *domain) OnExternalAppUpdateMessage(ctx ResourceContext, externalApp entities.ExternalApp, status types.ResourceStatus, opts UpdateAndDeleteOpts) error {
	xExtApp, err := d.findExternalApp(ctx, externalApp.Name)
	if err != nil {
		return errors.NewE(err)
	}

	if xExtApp == nil {
		return errors.Newf("no external apps found")
	}
	recordVersion, err := d.MatchRecordVersion(externalApp.Annotations, xExtApp.RecordVersion)
	if err != nil {
		return errors.NewE(err)
	}

	uExtApp, err := d.externalAppRepo.PatchById(
		ctx,
		xExtApp.Id,
		common.PatchForSyncFromAgent(&externalApp, recordVersion, status, common.PatchOpts{
			MessageTimestamp: opts.MessageTimestamp,
		}))
	d.resourceEventPublisher.PublishResourceEvent(ctx, uExtApp.GetResourceType(), uExtApp.GetName(), PublishUpdate)
	return errors.NewE(err)
}

func (d *domain) OnExternalAppDeleteMessage(ctx ResourceContext, externalApp entities.ExternalApp) error {
	err := d.externalAppRepo.DeleteOne(
		ctx,
		ctx.DBFilters().Add(fields.MetadataName, externalApp.Name),
	)
	if err != nil {
		return errors.NewE(err)
	}
	d.resourceEventPublisher.PublishResourceEvent(ctx, entities.ResourceTypeExternalApp, externalApp.Name, PublishDelete)
	return nil
}

func (d *domain) OnExternalAppApplyError(ctx ResourceContext, errMsg string, name string, opts UpdateAndDeleteOpts) error {
	uExtApp, err := d.externalAppRepo.Patch(
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

	d.resourceEventPublisher.PublishResourceEvent(ctx, entities.ResourceTypeExternalApp, uExtApp.Name, PublishDelete)
	return errors.NewE(err)
}

func (d *domain) ResyncExternalApp(ctx ResourceContext, name string) error {
	if err := d.canMutateResourcesInEnvironment(ctx); err != nil {
		return errors.NewE(err)
	}
	extApp, err := d.findExternalApp(ctx, name)
	if err != nil {
		return errors.NewE(err)
	}
	return d.resyncK8sResource(ctx, extApp.EnvironmentName, extApp.SyncStatus.Action, &extApp.ExternalApp, extApp.RecordVersion)
}
