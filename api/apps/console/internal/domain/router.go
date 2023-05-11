package domain

import (
	"fmt"
	"time"

	"kloudlite.io/apps/console/internal/domain/entities"
	"kloudlite.io/pkg/repos"
	t "kloudlite.io/pkg/types"
)

// query

func (d *domain) ListRouters(ctx ConsoleContext, namespace string) ([]*entities.Router, error) {
	if err := d.canReadResourcesInProject(ctx, namespace); err != nil {
		return nil, err
	}
	return d.routerRepo.Find(ctx, repos.Query{Filter: repos.Filter{
		"clusterName":        ctx.ClusterName,
		"accountName":        ctx.AccountName,
		"metadata.namespace": namespace,
	}})
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
		return nil, fmt.Errorf("no secret with name=%s,namespace=%s found", name, namespace)
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

	router.AccountName = ctx.AccountName
	router.ClusterName = ctx.ClusterName
	router.Generation = 1
	router.SyncStatus = t.GenSyncStatus(t.SyncActionApply, router.Generation)

	r, err := d.routerRepo.Create(ctx, &router)
	if err != nil {
		if d.routerRepo.ErrAlreadyExists(err) {
			return nil, fmt.Errorf("router with name=%q,namespace=%q already exists", router.Name, router.Namespace)
		}
		return nil, err
	}

	if err := d.applyK8sResource(ctx, &r.Router); err != nil {
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

	r, err := d.findRouter(ctx, router.Namespace, router.Name)
	if err != nil {
		return nil, err
	}

	r.Spec = router.Spec
	r.Generation += 1
	r.SyncStatus = t.GenSyncStatus(t.SyncActionApply, r.Generation)

	upRouter, err := d.routerRepo.UpdateById(ctx, r.Id, r)
	if err != nil {
		return nil, err
	}

	if err := d.applyK8sResource(ctx, &upRouter.Router); err != nil {
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

	r.SyncStatus = t.GenSyncStatus(t.SyncActionDelete, r.Generation)
	if _, err := d.routerRepo.UpdateById(ctx, r.Id, r); err != nil {
		return err
	}

	return d.deleteK8sResource(ctx, &r.Router)
}

func (d *domain) OnDeleteRouterMessage(ctx ConsoleContext, app entities.Router) error {
	a, err := d.findRouter(ctx, app.Namespace, app.Name)
	if err != nil {
		return err
	}

	return d.routerRepo.DeleteById(ctx, a.Id)
}

func (d *domain) OnUpdateRouterMessage(ctx ConsoleContext, router entities.Router) error {
	r, err := d.findRouter(ctx, router.Namespace, router.Name)
	if err != nil {
		return err
	}

	r.Status = router.Status
	r.SyncStatus.Error = nil
	r.SyncStatus.LastSyncedAt = time.Now()
	r.SyncStatus.Generation = router.Generation
	r.SyncStatus.State = t.ParseSyncState(router.Status.IsReady)

	_, err = d.routerRepo.UpdateById(ctx, r.Id, r)
	return err
}

func (d *domain) OnApplyRouterError(ctx ConsoleContext, errMsg string, namespace string, name string) error {
	m, err2 := d.findRouter(ctx, namespace, name)
	if err2 != nil {
		return err2
	}

	m.SyncStatus.Error = &errMsg
	_, err := d.routerRepo.UpdateById(ctx, m.Id, m)
	return err
}

func (d *domain) ResyncRouter(ctx ConsoleContext, namespace, name string) error {
	r, err := d.findRouter(ctx, namespace, name)
	if err != nil {
		return err
	}

	return d.resyncK8sResource(ctx, r.SyncStatus.Action, &r.Router)
}
