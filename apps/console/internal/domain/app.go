package domain

import (
	"fmt"

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
	return d.applyK8sResource(ctx, app.EnvironmentName, &app.App, app.RecordVersion)
}

func (d *domain) CreateApp(ctx ResourceContext, app entities.App) (*entities.App, error) {
	if err := d.canMutateResourcesInEnvironment(ctx); err != nil {
		return nil, errors.NewE(err)
	}

	env, err := d.findEnvironment(ctx.ConsoleContext, ctx.EnvironmentName)
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
	app.EnvironmentName = ctx.EnvironmentName
	app.SyncStatus = t.GenSyncStatus(t.SyncActionApply, app.RecordVersion)

	return d.createAndApplyApp(ctx, &app)
}

func (d *domain) createAndApplyApp(ctx ResourceContext, app *entities.App) (*entities.App, error) {
	if _, err := d.upsertEnvironmentResourceMapping(ctx, app); err != nil {
		return nil, errors.NewE(err)
	}

	napp, err := d.appRepo.Create(ctx, app)
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
	if err := d.deleteK8sResource(ctx, uapp.EnvironmentName, &uapp.App); err != nil {
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

	xapp, err := d.appRepo.FindOne(ctx, ctx.DBFilters().Add(fields.MetadataName, appIn.Name))
	if err != nil {
		return nil, errors.NewE(err)
	}

	if xapp == nil {
		return nil, errors.Newf("app does not exist")
	}

	// FIXME: hotfix till volume mounts for PVCs are not added in UI
	pvcMounts := make(map[int][]crdsv1.ContainerVolume)
	for i := range xapp.Spec.Containers {
		for _, volume := range xapp.Spec.Containers[i].Volumes {
			if volume.Type == crdsv1.PVCType {
				pvcMounts[i] = append(pvcMounts[i], volume)
			}
		}
	}

	existingMounts := make(map[int]map[string]struct{})
	for i := range appIn.Spec.Containers {
		for _, volume := range appIn.Spec.Containers[i].Volumes {
			if existingMounts[i] == nil {
				existingMounts[i] = make(map[string]struct{})
			}
			existingMounts[i][fmt.Sprintf("%#v", volume)] = struct{}{}
		}
	}

	for i := range appIn.Spec.Containers {
		for _, pvcVolume := range pvcMounts[i] {
			if _, ok := existingMounts[i][fmt.Sprintf("%#v", pvcVolume)]; !ok {
				appIn.Spec.Containers[i].Volumes = append(appIn.Spec.Containers[i].Volumes, pvcVolume)
			}
		}
	}

	// readiness and liveness probes
	if xapp.Spec.Containers[0].LivenessProbe != nil && appIn.Spec.Containers[0].LivenessProbe == nil {
		appIn.Spec.Containers[0].LivenessProbe = xapp.Spec.Containers[0].LivenessProbe
	}

	if xapp.Spec.Containers[0].ReadinessProbe != nil {
		appIn.Spec.Containers[0].ReadinessProbe = xapp.Spec.Containers[0].ReadinessProbe
	}

	patchDoc := repos.Document{
		fc.AppCiBuildId: appIn.CIBuildId,
		fc.AppSpec:      appIn.Spec,
	}

	patchForUpdate := common.PatchForUpdate(
		ctx,
		&appIn,
		common.PatchOpts{
			XPatch: patchDoc,
		})

	upApp, err := d.appRepo.Patch(ctx, ctx.DBFilters().Add(fields.MetadataName, appIn.Name), patchForUpdate)
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
func (d *domain) InterceptApp(ctx ResourceContext, appName string, deviceName string, intercept bool, portMappings []crdsv1.AppInterceptPortMappings) (bool, error) {
	if err := d.canMutateResourcesInEnvironment(ctx); err != nil {
		return false, errors.NewE(err)
	}

	patch := repos.Document{
		fc.AppSpecInterceptEnabled:  intercept,
		fc.AppSpecInterceptToDevice: deviceName,
	}

	if portMappings != nil {
		patch[fc.AppSpecInterceptPortMappings] = portMappings
	}

	uApp, err := d.appRepo.Patch(ctx, ctx.DBFilters().Add(fields.MetadataName, appName), patch)
	if err != nil {
		return false, errors.NewE(err)
	}
	if err := d.applyApp(ctx, uApp); err != nil {
		return false, errors.NewE(err)
	}
	return true, nil
}

// InterceptApp implements Domain.
func (d *domain) InterceptAppOnLocalCluster(ctx ResourceContext, appName string, clusterName string, ipAddr string, intercept bool, portMappings []crdsv1.AppInterceptPortMappings) (bool, error) {
	if err := d.canMutateResourcesInEnvironment(ctx); err != nil {
		return false, errors.NewE(err)
	}

	patch := repos.Document{
		fc.AppSpecInterceptEnabled:  intercept,
		fc.AppSpecInterceptToDevice: clusterName,
		fc.AppSpecInterceptToIPAddr: ipAddr,
	}

	if portMappings != nil {
		patch[fc.AppSpecInterceptPortMappings] = portMappings
	}

	uApp, err := d.appRepo.Patch(ctx, ctx.DBFilters().Add(fields.MetadataName, appName), patch)
	if err != nil {
		return false, errors.NewE(err)
	}
	if err := d.applyApp(ctx, uApp); err != nil {
		return false, errors.NewE(err)
	}
	return true, nil
}

func (d *domain) RestartApp(ctx ResourceContext, appName string) error {
	if err := d.canMutateResourcesInEnvironment(ctx); err != nil {
		return errors.NewE(err)
	}

	app, err := d.findApp(ctx, appName)
	if err != nil {
		return err
	}

	if err := d.restartK8sResource(ctx, ctx.EnvironmentName, app.Namespace, app.GetEnsuredLabels()); err != nil {
		return err
	}

	return nil
}

func (d *domain) RemoveDeviceIntercepts(ctx ResourceContext, deviceName string) error {
	apps, err := d.appRepo.Find(ctx, repos.Query{
		Filter: repos.Filter{
			fields.AccountName:          ctx.AccountName,
			fields.EnvironmentName:      ctx.EnvironmentName,
			fc.AppSpecInterceptToDevice: deviceName,
		},
		Sort: nil,
	})
	if err != nil {
		return errors.NewE(err)
	}

	for i := range apps {
		patchForUpdate := repos.Document{
			fc.AppSpecInterceptEnabled: false,
		}

		up, err := d.appRepo.PatchById(ctx, apps[i].Id, patchForUpdate)
		if err != nil {
			return errors.NewE(err)
		}

		if err := d.applyApp(ctx, up); err != nil {
			return errors.NewE(err)
		}
	}

	return nil
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

	uapp, err := d.appRepo.PatchById(ctx, xApp.Id,
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
	return d.resyncK8sResource(ctx, a.EnvironmentName, a.SyncStatus.Action, &a.App, a.RecordVersion)
}
