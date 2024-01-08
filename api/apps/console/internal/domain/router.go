package domain

import (
	"github.com/kloudlite/api/apps/console/internal/entities"
	"github.com/kloudlite/api/common"
	"github.com/kloudlite/api/pkg/errors"
	"github.com/kloudlite/api/pkg/repos"
	t "github.com/kloudlite/api/pkg/types"
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

	router, err := d.routerRepo.FindOne(ctx, filter)
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

	if _, err := d.upsertResourceMapping(ctx, &router); err != nil {
		return nil, errors.NewE(err)
	}

	r, err := d.routerRepo.Create(ctx, &router)
	if err != nil {
		if d.routerRepo.ErrAlreadyExists(err) {
			// TODO: better insights into error, when it is being caused by duplicated indexes
			return nil, errors.NewE(err)
		}
		return nil, errors.NewE(err)
	}
	d.resourceEventPublisher.PublishRouterEvent(&router, PublishAdd)

	if err := d.applyK8sResource(ctx, router.ProjectName, &router.Router, router.RecordVersion); err != nil {
		return r, errors.NewE(err)
	}

	return r, nil
}

func (d *domain) UpdateRouter(ctx ResourceContext, router entities.Router) (*entities.Router, error) {
	if err := d.canMutateResourcesInEnvironment(ctx); err != nil {
		return nil, errors.NewE(err)
	}

	xRouter, err := d.findRouter(ctx, router.Name)
	if err != nil {
		return nil, errors.NewE(err)
	}

	router.EnsureGVK()
	router.Namespace = xRouter.Namespace
	if err := d.k8sClient.ValidateObject(ctx, &router.Router); err != nil {
		return nil, errors.NewE(err)
	}

	xRouter.IncrementRecordVersion()
	xRouter.LastUpdatedBy = common.CreatedOrUpdatedBy{
		UserId:    ctx.UserId,
		UserName:  ctx.UserName,
		UserEmail: ctx.UserEmail,
	}
	xRouter.DisplayName = router.DisplayName

	xRouter.Annotations = router.Annotations
	xRouter.Labels = router.Labels

	xRouter.Spec = router.Spec
	xRouter.SyncStatus = t.GenSyncStatus(t.SyncActionApply, xRouter.RecordVersion)

	upRouter, err := d.routerRepo.UpdateById(ctx, xRouter.Id, xRouter)
	if err != nil {
		return nil, errors.NewE(err)
	}
	d.resourceEventPublisher.PublishRouterEvent(upRouter, PublishUpdate)

	if err := d.applyK8sResource(ctx, upRouter.ProjectName, &upRouter.Router, upRouter.RecordVersion); err != nil {
		return upRouter, errors.NewE(err)
	}

	return upRouter, nil
}

func (d *domain) DeleteRouter(ctx ResourceContext, name string) error {
	if err := d.canMutateResourcesInEnvironment(ctx); err != nil {
		return errors.NewE(err)
	}

	r, err := d.findRouter(ctx, name)
	if err != nil {
		return errors.NewE(err)
	}

	r.SyncStatus = t.GenSyncStatus(t.SyncActionDelete, r.RecordVersion)
	if _, err := d.routerRepo.UpdateById(ctx, r.Id, r); err != nil {
		return errors.NewE(err)
	}
	d.resourceEventPublisher.PublishRouterEvent(r, PublishUpdate)

	if err := d.deleteK8sResource(ctx, r.ProjectName, &r.Router); err != nil {
		if errors.Is(err, ErrNoClusterAttached) {
			return d.routerRepo.DeleteById(ctx, r.Id)
		}
		return err
	}

	return nil
}

func (d *domain) OnRouterDeleteMessage(ctx ResourceContext, router entities.Router) error {
	xRouter, err := d.findRouter(ctx, router.Name)
	if err != nil {
		return errors.NewE(err)
	}

	if err := d.MatchRecordVersion(router.Annotations, xRouter.RecordVersion); err != nil {
		return d.resyncK8sResource(ctx, xRouter.ProjectName, xRouter.SyncStatus.Action, &xRouter.Router, xRouter.RecordVersion)
	}

	err = d.routerRepo.DeleteById(ctx, xRouter.Id)
	if err != nil {
		return errors.NewE(err)
	}
	d.resourceEventPublisher.PublishRouterEvent(xRouter, PublishDelete)
	return nil
}

func (d *domain) OnRouterUpdateMessage(ctx ResourceContext, router entities.Router, status types.ResourceStatus, opts UpdateAndDeleteOpts) error {
	xRouter, err := d.findRouter(ctx, router.Name)
	if err != nil {
		return errors.NewE(err)
	}

	if err := d.MatchRecordVersion(router.Annotations, xRouter.RecordVersion); err != nil {
		return d.resyncK8sResource(ctx, xRouter.ProjectName, xRouter.SyncStatus.Action, &xRouter.Router, xRouter.RecordVersion)
	}

	xRouter.CreationTimestamp = router.CreationTimestamp
	xRouter.Labels = router.Labels
	xRouter.Annotations = router.Annotations
	xRouter.Generation = router.Generation

	xRouter.Status = router.Status

	xRouter.SyncStatus.State = func() t.SyncState {
		if status == types.ResourceStatusDeleting {
			return t.SyncStateDeletingAtAgent
		}
		return t.SyncStateUpdatedAtAgent
	}()
	xRouter.SyncStatus.RecordVersion = xRouter.RecordVersion
	xRouter.SyncStatus.Error = nil
	xRouter.SyncStatus.LastSyncedAt = opts.MessageTimestamp

	_, err = d.routerRepo.UpdateById(ctx, xRouter.Id, xRouter)
	d.resourceEventPublisher.PublishRouterEvent(xRouter, PublishUpdate)
	return errors.NewE(err)
}

func (d *domain) OnRouterApplyError(ctx ResourceContext, errMsg string, name string, opts UpdateAndDeleteOpts) error {
	m, err := d.findRouter(ctx, name)
	if err != nil {
		return errors.NewE(err)
	}

	m.SyncStatus.State = t.SyncStateErroredAtAgent
	m.SyncStatus.LastSyncedAt = opts.MessageTimestamp
	m.SyncStatus.Error = &errMsg

	_, err = d.routerRepo.UpdateById(ctx, m.Id, m)
	return errors.NewE(err)
}

func (d *domain) ResyncRouter(ctx ResourceContext, name string) error {
	router, err := d.findRouter(ctx, name)
	if err != nil {
		return errors.NewE(err)
	}

	return d.resyncK8sResource(ctx, router.ProjectName, router.SyncStatus.Action, &router.Router, router.RecordVersion)
}
