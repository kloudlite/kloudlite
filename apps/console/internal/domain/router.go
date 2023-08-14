package domain

import (
	"fmt"
	"time"

	"kloudlite.io/apps/console/internal/domain/entities"
	"kloudlite.io/pkg/repos"
	t "kloudlite.io/pkg/types"
)

// query

func (d *domain) ListRouters(ctx ConsoleContext, namespace string, search map[string]repos.MatchFilter, pq repos.CursorPagination) (*repos.PaginatedRecord[*entities.Router], error) {
	if err := d.canReadResourcesInProject(ctx, namespace); err != nil {
		return nil, err
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
		return nil, err
	}
	if router == nil {
		return nil, fmt.Errorf("no router with name=%q,namespace=%q found", name, namespace)
	}
	return router, nil
}

func (d *domain) GetRouter(ctx ConsoleContext, namespace string, name string) (*entities.Router, error) {
	if err := d.canReadResourcesInProject(ctx, namespace); err != nil {
		return nil, err
	}
	return d.findRouter(ctx, namespace, name)
}

// mutations

func (d *domain) CreateRouter(ctx ConsoleContext, router entities.Router) (*entities.Router, error) {
	if err := d.canMutateResourcesInProject(ctx, router.Namespace); err != nil {
		return nil, err
	}

	router.EnsureGVK()
	if err := d.k8sExtendedClient.ValidateStruct(ctx, &router.Router); err != nil {
		return nil, err
	}

	router.IncrementRecordVersion()
	router.AccountName = ctx.AccountName
	router.ClusterName = ctx.ClusterName
	router.SyncStatus = t.GenSyncStatus(t.SyncActionApply, router.RecordVersion)

	r, err := d.routerRepo.Create(ctx, &router)
	if err != nil {
		if d.routerRepo.ErrAlreadyExists(err) {
			// TODO: better insights into error, when it is being caused by duplicated indexes
			return nil, err
		}
		return nil, err
	}

	if err := d.applyK8sResource(ctx, &r.Router, 0); err != nil {
		return r, err
	}

	return r, nil
}

func (d *domain) UpdateRouter(ctx ConsoleContext, router entities.Router) (*entities.Router, error) {
	if err := d.canMutateResourcesInProject(ctx, router.Namespace); err != nil {
		return nil, err
	}

	router.EnsureGVK()
	if err := d.k8sExtendedClient.ValidateStruct(ctx, &router.Router); err != nil {
		return nil, err
	}

	exRouter, err := d.findRouter(ctx, router.Namespace, router.Name)
	if err != nil {
		return nil, err
	}

	exRouter.IncrementRecordVersion()
	exRouter.Annotations = router.Annotations
	exRouter.Labels = router.Labels

	exRouter.Spec = router.Spec
	exRouter.SyncStatus = t.GenSyncStatus(t.SyncActionApply, exRouter.RecordVersion)

	upRouter, err := d.routerRepo.UpdateById(ctx, exRouter.Id, exRouter)
	if err != nil {
		return nil, err
	}

	if err := d.applyK8sResource(ctx, &upRouter.Router, upRouter.RecordVersion); err != nil {
		return upRouter, err
	}

	return upRouter, nil
}

func (d *domain) DeleteRouter(ctx ConsoleContext, namespace string, name string) error {
	if err := d.canMutateResourcesInProject(ctx, namespace); err != nil {
		return err
	}

	r, err := d.findRouter(ctx, namespace, name)
	if err != nil {
		return err
	}

	r.SyncStatus = t.GenSyncStatus(t.SyncActionDelete, r.RecordVersion)
	if _, err := d.routerRepo.UpdateById(ctx, r.Id, r); err != nil {
		return err
	}

	return d.deleteK8sResource(ctx, &r.Router)
}

func (d *domain) OnDeleteRouterMessage(ctx ConsoleContext, router entities.Router) error {
	exRouter, err := d.findRouter(ctx, router.Namespace, router.Name)
	if err != nil {
		return err
	}

	if err := d.MatchRecordVersion(router.Annotations, exRouter.RecordVersion); err != nil {
		return err
	}

	return d.routerRepo.DeleteById(ctx, exRouter.Id)
}

func (d *domain) OnUpdateRouterMessage(ctx ConsoleContext, router entities.Router) error {
	exRouter, err := d.findRouter(ctx, router.Namespace, router.Name)
	if err != nil {
		return err
	}

	annotatedVersion, err := d.parseRecordVersionFromAnnotations(router.Annotations)
	if err != nil {
		return d.resyncK8sResource(ctx, exRouter.SyncStatus.Action, &exRouter.Router, exRouter.RecordVersion)
	}

	if annotatedVersion != exRouter.RecordVersion {
		return d.resyncK8sResource(ctx, exRouter.SyncStatus.Action, &exRouter.Router, exRouter.RecordVersion)
	}

	exRouter.CreationTimestamp = router.CreationTimestamp
	exRouter.Labels = router.Labels
	exRouter.Annotations = router.Annotations
	exRouter.Generation = router.Generation

	exRouter.Status = router.Status

	exRouter.SyncStatus.State = t.SyncStateReceivedUpdateFromAgent
	exRouter.SyncStatus.RecordVersion = annotatedVersion
	exRouter.SyncStatus.Error = nil
	exRouter.SyncStatus.LastSyncedAt = time.Now()

	_, err = d.routerRepo.UpdateById(ctx, exRouter.Id, exRouter)
	return err
}

func (d *domain) OnApplyRouterError(ctx ConsoleContext, errMsg string, namespace string, name string) error {
	m, err := d.findRouter(ctx, namespace, name)
	if err != nil {
		return err
	}

	m.SyncStatus.State = t.SyncStateErroredAtAgent
	m.SyncStatus.LastSyncedAt = time.Now()
	m.SyncStatus.Error = &errMsg

	_, err = d.routerRepo.UpdateById(ctx, m.Id, m)
	return err
}

func (d *domain) ResyncRouter(ctx ConsoleContext, namespace, name string) error {
	r, err := d.findRouter(ctx, namespace, name)
	if err != nil {
		return err
	}

	return d.resyncK8sResource(ctx, r.SyncStatus.Action, &r.Router, r.RecordVersion)
}
