package domain

import (
	"time"

	"github.com/kloudlite/api/apps/console/internal/entities"
	"github.com/kloudlite/api/common"
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

	router.Spec.Https = &crdsv1.Https{
		Enabled:       true,
		ForceRedirect: true,
	}

	if _, err := d.upsertEnvironmentResourceMapping(ctx, &router); err != nil {
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

	if router.Spec.Https == nil {
		router.Spec.Https = &crdsv1.Https{
			Enabled:       true,
			ForceRedirect: true,
		}
	}

	patch := repos.Document{
		"recordVersion": xRouter.RecordVersion + 1,
		"displayName":   router.DisplayName,
		"lastUpdatedBy": common.CreatedOrUpdatedBy{
			UserId:    ctx.UserId,
			UserName:  ctx.UserName,
			UserEmail: ctx.UserEmail,
		},
		"metadata.labels":       router.Labels,
		"metadata.annnotations": router.Annotations,

		"spec": router.Spec,

		"syncStatus.state":           t.SyncStateInQueue,
		"syncStatus.syncScheduledAt": time.Now(),
		"syncStatus.action":          t.SyncActionApply,
	}

	upRouter, err := d.routerRepo.PatchById(ctx, xRouter.Id, patch)
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

	patch := repos.Document{
		"markedForDeletion":          true,
		"syncStatus.syncScheduledAt": time.Now(),
		"syncStatus.action":          t.SyncActionDelete,
		"syncStatus.state":           t.SyncStateInQueue,
	}

	urouter, err := d.routerRepo.PatchById(ctx, r.Id, patch)
	if err != nil {
		return errors.NewE(err)
	}
	d.resourceEventPublisher.PublishRouterEvent(urouter, PublishUpdate)

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

	patch := repos.Document{
		"metadata.creationTimestamp": router.CreationTimestamp,
		"metadata.labels":            router.Labels,
		"metadata.annotations":       router.Annotations,
		"metadata.generation":        router.Generation,

		"status": router.Status,

		"syncStatus.state": func() t.SyncState {
			if status == types.ResourceStatusDeleting {
				return t.SyncStateDeletingAtAgent
			}
			return t.SyncStateUpdatedAtAgent
		}(),
		"syncStatus.recordVersion": xRouter.RecordVersion,
		"syncStatus.lastSyncedAt":  opts.MessageTimestamp,
		"syncStatus.error":         nil,
	}

	urouter, err := d.routerRepo.PatchById(ctx, xRouter.Id, patch)
	if err != nil {
		return errors.NewE(err)
	}
	d.resourceEventPublisher.PublishRouterEvent(urouter, PublishUpdate)
	return nil
}

func (d *domain) OnRouterApplyError(ctx ResourceContext, errMsg string, name string, opts UpdateAndDeleteOpts) error {
	m, err := d.findRouter(ctx, name)
	if err != nil {
		return errors.NewE(err)
	}

	patch := repos.Document{
		"syncStatus.state":        t.SyncStateErroredAtAgent,
		"syncStatus.lastSyncedAt": opts.MessageTimestamp,
		"syncStatus.error":        errMsg,
	}

	urouter, err := d.routerRepo.PatchById(ctx, m.Id, patch)
	if err != nil {
		return err
	}

	d.resourceEventPublisher.PublishRouterEvent(urouter, PublishUpdate)

	return nil
}

func (d *domain) ResyncRouter(ctx ResourceContext, name string) error {
	router, err := d.findRouter(ctx, name)
	if err != nil {
		return errors.NewE(err)
	}

	return d.resyncK8sResource(ctx, router.ProjectName, router.SyncStatus.Action, &router.Router, router.RecordVersion)
}
