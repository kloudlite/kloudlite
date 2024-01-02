package domain

import (
	"github.com/kloudlite/api/apps/console/internal/entities"
	"github.com/kloudlite/api/common"
	"github.com/kloudlite/api/pkg/errors"
	"github.com/kloudlite/api/pkg/repos"
	t "github.com/kloudlite/api/pkg/types"
	"github.com/kloudlite/operator/operators/resource-watcher/types"
)

func (d *domain) ListRouters(ctx ConsoleContext, namespace string, search map[string]repos.MatchFilter, pq repos.CursorPagination) (*repos.PaginatedRecord[*entities.Router], error) {
	if err := d.canReadResourcesInProject(ctx, namespace); err != nil {
		return nil, errors.NewE(err)
	}

	filter := repos.Filter{
		"clusterName":        ctx.ClusterName,
		"accountName":        ctx.AccountName,
		"metadata.namespace": namespace,
	}

	return d.routerRepo.FindPaginated(ctx, d.routerRepo.MergeMatchFilters(filter, search), pq)
}

func (d *domain) findRouter(ctx ConsoleContext, namespace string, name string) (*entities.Router, error) {
	router, err := d.routerRepo.FindOne(ctx, repos.Filter{
		"accountName":        ctx.AccountName,
		"clusterName":        ctx.ClusterName,
		"metadata.namespace": namespace,
		"metadata.name":      name,
	})
	if err != nil {
		return nil, errors.NewE(err)
	}
	if router == nil {
		return nil, errors.Newf("no router with name=%q,namespace=%q found", name, namespace)
	}
	return router, nil
}

func (d *domain) GetRouter(ctx ConsoleContext, namespace string, name string) (*entities.Router, error) {
	if err := d.canReadResourcesInProject(ctx, namespace); err != nil {
		return nil, errors.NewE(err)
	}
	return d.findRouter(ctx, namespace, name)
}

// mutations

func (d *domain) CreateRouter(ctx ConsoleContext, router entities.Router) (*entities.Router, error) {
	if err := d.canMutateResourcesInProject(ctx, router.Namespace); err != nil {
		return nil, errors.NewE(err)
	}

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
	router.ClusterName = ctx.ClusterName
	router.SyncStatus = t.GenSyncStatus(t.SyncActionApply, router.RecordVersion)

	r, err := d.routerRepo.Create(ctx, &router)
	if err != nil {
		if d.routerRepo.ErrAlreadyExists(err) {
			// TODO: better insights into error, when it is being caused by duplicated indexes
			return nil, errors.NewE(err)
		}
		return nil, errors.NewE(err)
	}
	d.resourceEventPublisher.PublishRouterEvent(&router, PublishAdd)

	if err := d.applyK8sResource(ctx, &r.Router, 0); err != nil {
		return r, errors.NewE(err)
	}

	return r, nil
}

func (d *domain) UpdateRouter(ctx ConsoleContext, router entities.Router) (*entities.Router, error) {
	if err := d.canMutateResourcesInProject(ctx, router.Namespace); err != nil {
		return nil, errors.NewE(err)
	}

	router.EnsureGVK()
	if err := d.k8sClient.ValidateObject(ctx, &router.Router); err != nil {
		return nil, errors.NewE(err)
	}

	exRouter, err := d.findRouter(ctx, router.Namespace, router.Name)
	if err != nil {
		return nil, errors.NewE(err)
	}

	exRouter.IncrementRecordVersion()
	exRouter.LastUpdatedBy = common.CreatedOrUpdatedBy{
		UserId:    ctx.UserId,
		UserName:  ctx.UserName,
		UserEmail: ctx.UserEmail,
	}
	exRouter.DisplayName = router.DisplayName

	exRouter.Annotations = router.Annotations
	exRouter.Labels = router.Labels

	exRouter.Spec = router.Spec
	exRouter.SyncStatus = t.GenSyncStatus(t.SyncActionApply, exRouter.RecordVersion)

	upRouter, err := d.routerRepo.UpdateById(ctx, exRouter.Id, exRouter)
	if err != nil {
		return nil, errors.NewE(err)
	}
	d.resourceEventPublisher.PublishRouterEvent(upRouter, PublishUpdate)

	if err := d.applyK8sResource(ctx, &upRouter.Router, upRouter.RecordVersion); err != nil {
		return upRouter, errors.NewE(err)
	}

	return upRouter, nil
}

func (d *domain) DeleteRouter(ctx ConsoleContext, namespace string, name string) error {
	if err := d.canMutateResourcesInProject(ctx, namespace); err != nil {
		return errors.NewE(err)
	}

	r, err := d.findRouter(ctx, namespace, name)
	if err != nil {
		return errors.NewE(err)
	}

	r.SyncStatus = t.GenSyncStatus(t.SyncActionDelete, r.RecordVersion)
	if _, err := d.routerRepo.UpdateById(ctx, r.Id, r); err != nil {
		return errors.NewE(err)
	}
	d.resourceEventPublisher.PublishRouterEvent(r, PublishUpdate)

	return d.deleteK8sResource(ctx, &r.Router)
}

func (d *domain) OnRouterDeleteMessage(ctx ConsoleContext, router entities.Router) error {
	exRouter, err := d.findRouter(ctx, router.Namespace, router.Name)
	if err != nil {
		return errors.NewE(err)
	}

	if err := d.MatchRecordVersion(router.Annotations, exRouter.RecordVersion); err != nil {
		return errors.NewE(err)
	}

	err = d.routerRepo.DeleteById(ctx, exRouter.Id)
	if err != nil {
		return errors.NewE(err)
	}
	d.resourceEventPublisher.PublishRouterEvent(exRouter, PublishDelete)
	return nil
}

func (d *domain) OnRouterUpdateMessage(ctx ConsoleContext, router entities.Router, status types.ResourceStatus, opts UpdateAndDeleteOpts) error {
	exRouter, err := d.findRouter(ctx, router.Namespace, router.Name)
	if err != nil {
		return errors.NewE(err)
	}

	if err := d.MatchRecordVersion(router.Annotations, exRouter.RecordVersion); err != nil {
		return d.resyncK8sResource(ctx, exRouter.SyncStatus.Action, &exRouter.Router, exRouter.RecordVersion)
	}

	exRouter.CreationTimestamp = router.CreationTimestamp
	exRouter.Labels = router.Labels
	exRouter.Annotations = router.Annotations
	exRouter.Generation = router.Generation

	exRouter.Status = router.Status

	exRouter.SyncStatus.State = func() t.SyncState {
		if status == types.ResourceStatusDeleting {
			return t.SyncStateDeletingAtAgent
		}
		return t.SyncStateUpdatedAtAgent
	}()
	exRouter.SyncStatus.RecordVersion = exRouter.RecordVersion
	exRouter.SyncStatus.Error = nil
	exRouter.SyncStatus.LastSyncedAt = opts.MessageTimestamp

	_, err = d.routerRepo.UpdateById(ctx, exRouter.Id, exRouter)
	d.resourceEventPublisher.PublishRouterEvent(exRouter, PublishUpdate)
	return errors.NewE(err)
}

func (d *domain) OnRouterApplyError(ctx ConsoleContext, errMsg string, namespace string, name string, opts UpdateAndDeleteOpts) error {
	m, err := d.findRouter(ctx, namespace, name)
	if err != nil {
		return errors.NewE(err)
	}

	m.SyncStatus.State = t.SyncStateErroredAtAgent
	m.SyncStatus.LastSyncedAt = opts.MessageTimestamp
	m.SyncStatus.Error = &errMsg

	_, err = d.routerRepo.UpdateById(ctx, m.Id, m)
	return errors.NewE(err)
}

func (d *domain) ResyncRouter(ctx ConsoleContext, namespace, name string) error {
	r, err := d.findRouter(ctx, namespace, name)
	if err != nil {
		return errors.NewE(err)
	}

	return d.resyncK8sResource(ctx, r.SyncStatus.Action, &r.Router, r.RecordVersion)
}
