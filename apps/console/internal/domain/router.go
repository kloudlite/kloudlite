package domain

import (
	"fmt"
	"time"

	"kloudlite.io/apps/console/internal/domain/entities"
	"kloudlite.io/pkg/repos"
	t "kloudlite.io/pkg/types"
)

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
	router.SyncStatus = t.GetSyncStatusForCreation()

	r, err := d.routerRepo.Create(ctx, &router)
	if err != nil {
		if d.routerRepo.ErrAlreadyExists(err) {
			return nil, fmt.Errorf("router with name %q already exists", router.Name)
		}
		return nil, err
	}

	if err := d.applyK8sResource(ctx, &r.Router); err != nil {
		return r, err
	}

	return r, nil
}

func (d *domain) DeleteRouter(ctx ConsoleContext, namespace string, name string) error {
	if err := d.canMutateResourcesInProject(ctx, namespace); err != nil {
		return err
	}

	r, err := d.findRouter(ctx, namespace, name)
	if err != nil {
		return err
	}

	r.SyncStatus = t.GetSyncStatusForDeletion(r.Generation)
	if _, err := d.routerRepo.UpdateById(ctx, r.Id, r); err != nil {
		return err
	}

	return d.deleteK8sResource(ctx, &r.Router)
}

func (d *domain) GetRouter(ctx ConsoleContext, namespace string, name string) (*entities.Router, error) {
	if err := d.canReadResourcesInProject(ctx, namespace); err != nil {
		return nil, err
	}
	return d.findRouter(ctx, namespace, name)
}

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
	r.SyncStatus = t.GetSyncStatusForUpdation(r.Generation + 1)

	upRouter, err := d.routerRepo.UpdateById(ctx, r.Id, r)
	if err != nil {
		return nil, err
	}

	if err := d.applyK8sResource(ctx, &upRouter.Router); err != nil {
		return upRouter, err
	}

	return upRouter, nil
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
	r.SyncStatus.LastSyncedAt = time.Now()
	r.SyncStatus.State = t.ParseSyncState(router.Status.IsReady)

	_, err = d.routerRepo.UpdateById(ctx, r.Id, r)
	return err
}

func (d *domain) OnApplyRouterError(ctx ConsoleContext, err error, namespace, name string) error {
	m, err2 := d.findRouter(ctx, namespace, name)
	if err2 != nil {
		return err2
	}

	m.SyncStatus.Error = err.Error()
	_, err = d.routerRepo.UpdateById(ctx, m.Id, m)
	return err
}
