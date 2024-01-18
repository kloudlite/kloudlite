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

func (d *domain) ListRouters(ctx ResourceContext, search map[string]repos.MatchFilter, pq repos.CursorPagination) (*repos.PaginatedRecord[*entities.Router], error) {
	if err := d.canReadResourcesInEnvironment(ctx); err != nil {
		return nil, errors.NewE(err)
	}

	filter := ctx.DBFilters()
	return d.routerRepo.FindPaginated(ctx, d.routerRepo.MergeMatchFilters(filter, search), pq)
}

func (d *domain) findRouter(ctx ResourceContext, name string) (*entities.Router, error) {
	filter := ctx.DBFilters()
	filter.Add("metadata.name", name)

	router, err := d.routerRepo.FindOne(
		ctx,
		ctx.DBFilters().Add(fields.MetadataName, name),
	)
	if err != nil {
		return nil, errors.NewE(err)
	}
	if router == nil {
		return nil, errors.Newf("no router with name (%s) found", name)
	}
	return router, nil
}

func (d *domain) GetRouter(ctx ResourceContext, name string) (*entities.Router, error) {
	if err := d.canReadResourcesInEnvironment(ctx); err != nil {
		return nil, errors.NewE(err)
	}
	return d.findRouter(ctx, name)
}

// mutations

func (d *domain) CreateRouter(ctx ResourceContext, router entities.Router) (*entities.Router, error) {
	if err := d.canMutateResourcesInEnvironment(ctx); err != nil {
		return nil, errors.NewE(err)
	}

	namespace, err := d.envTargetNamespace(ctx.ConsoleContext, ctx.ProjectName, ctx.EnvironmentName)
	if err != nil {
		return nil, err
	}

	router.Namespace = namespace

	router.EnsureGVK()
	if err := d.k8sClient.ValidateObject(ctx, &router.Router); err != nil {
		return nil, errors.NewE(err)
	}

	router.IncrementRecordVersion()
	router.CreatedBy = common.CreatedOrUpdatedBy{
		UserId:    ctx.UserId,
		UserName:  ctx.UserName,
		UserEmail: ctx.UserEmail,
	}
	router.LastUpdatedBy = router.CreatedBy

	router.AccountName = ctx.AccountName
	router.ProjectName = ctx.ProjectName
	router.EnvironmentName = ctx.EnvironmentName
	router.SyncStatus = t.GenSyncStatus(t.SyncActionApply, router.RecordVersion)

	router.Spec.Https = &crdsv1.Https{
		Enabled:       true,
		ForceRedirect: true,
	}

	if _, err := d.upsertEnvironmentResourceMapping(ctx, &router); err != nil {
		return nil, errors.NewE(err)
	}

	nrouter, err := d.routerRepo.Create(ctx, &router)
	if err != nil {
		if d.routerRepo.ErrAlreadyExists(err) {
			// TODO: better insights into error, when it is being caused by duplicated indexes
			return nil, errors.NewE(err)
		}
		return nil, errors.NewE(err)
	}
	d.resourceEventPublisher.PublishResourceEvent(ctx, entities.ResourceTypeRouter, nrouter.Name, PublishAdd)

	if err := d.applyK8sResource(ctx, nrouter.ProjectName, &nrouter.Router, nrouter.RecordVersion); err != nil {
		return nrouter, errors.NewE(err)
	}

	return nrouter, nil
}

func (d *domain) UpdateRouter(ctx ResourceContext, router entities.Router) (*entities.Router, error) {
	if err := d.canMutateResourcesInEnvironment(ctx); err != nil {
		return nil, errors.NewE(err)
	}
	router.EnsureGVK()
	router.Namespace = "trest"
	if err := d.k8sClient.ValidateObject(ctx, &router.Router); err != nil {
		return nil, errors.NewE(err)
	}

	if router.Spec.Https == nil {
		router.Spec.Https = &crdsv1.Https{
			Enabled:       true,
			ForceRedirect: true,
		}
	}

	patchForUpdate := common.PatchForUpdate(
		ctx,
		&router,
		common.PatchOpts{
			XPatch: repos.Document{
				fc.AppSpec: router.Spec,
			},
		})

	upRouter, err := d.routerRepo.Patch(
		ctx,
		ctx.DBFilters().Add(fields.MetadataName, router.Name),
		patchForUpdate,
	)
	if err != nil {
		return nil, errors.NewE(err)
	}
	d.resourceEventPublisher.PublishResourceEvent(ctx, entities.ResourceTypeRouter, upRouter.Name, PublishUpdate)

	if err := d.applyK8sResource(ctx, upRouter.ProjectName, &upRouter.Router, upRouter.RecordVersion); err != nil {
		return upRouter, errors.NewE(err)
	}

	return upRouter, nil
}

func (d *domain) DeleteRouter(ctx ResourceContext, name string) error {
	if err := d.canMutateResourcesInEnvironment(ctx); err != nil {
		return errors.NewE(err)
	}

	urouter, err := d.routerRepo.Patch(
		ctx,
		ctx.DBFilters().Add(fields.MetadataName, name),
		common.PatchForMarkDeletion(),
	)

	if err != nil {
		return errors.NewE(err)
	}

	d.resourceEventPublisher.PublishResourceEvent(ctx, entities.ResourceTypeRouter, urouter.Name, PublishUpdate)

	if err := d.deleteK8sResource(ctx, urouter.ProjectName, &urouter.Router); err != nil {
		if errors.Is(err, ErrNoClusterAttached) {
			return d.routerRepo.DeleteById(ctx, urouter.Id)
		}
		return errors.NewE(err)
	}
	return nil
}

func (d *domain) OnRouterDeleteMessage(ctx ResourceContext, router entities.Router) error {
	err := d.routerRepo.DeleteOne(
		ctx,
		ctx.DBFilters().Add(fields.MetadataName, router.Name),
	)

	if err != nil {
		return errors.NewE(err)
	}

	d.resourceEventPublisher.PublishResourceEvent(ctx, entities.ResourceTypeRouter, router.Name, PublishDelete)
	return nil
}

func (d *domain) OnRouterUpdateMessage(ctx ResourceContext, router entities.Router, status types.ResourceStatus, opts UpdateAndDeleteOpts) error {
	xRouter, err := d.findRouter(ctx, router.Name)
	if err != nil {
		return errors.NewE(err)
	}

	if xRouter == nil {
		return errors.Newf("no router found")
	}

	if err := d.MatchRecordVersion(router.Annotations, xRouter.RecordVersion); err != nil {
		return d.resyncK8sResource(ctx, xRouter.ProjectName, xRouter.SyncStatus.Action, &xRouter.Router, xRouter.RecordVersion)
	}

	urouter, err := d.routerRepo.PatchById(
		ctx,
		xRouter.Id,
		common.PatchForSyncFromAgent(&router, status, common.PatchOpts{
			MessageTimestamp: opts.MessageTimestamp,
		}))

	if err != nil {
		return errors.NewE(err)
	}

	d.resourceEventPublisher.PublishResourceEvent(ctx, urouter.GetResourceType(), urouter.GetName(), PublishUpdate)
	return nil
}

func (d *domain) OnRouterApplyError(ctx ResourceContext, errMsg string, name string, opts UpdateAndDeleteOpts) error {
	urouter, err := d.routerRepo.Patch(
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
		return err
	}

	d.resourceEventPublisher.PublishResourceEvent(ctx, entities.ResourceTypeRouter, urouter.Name, PublishDelete)

	return nil
}

func (d *domain) ResyncRouter(ctx ResourceContext, name string) error {
	router, err := d.findRouter(ctx, name)
	if err != nil {
		return errors.NewE(err)
	}

	return d.resyncK8sResource(ctx, router.ProjectName, router.SyncStatus.Action, &router.Router, router.RecordVersion)
}
