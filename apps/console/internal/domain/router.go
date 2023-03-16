package domain

import (
	"fmt"

	"kloudlite.io/apps/console/internal/domain/entities"
	"kloudlite.io/pkg/repos"
)

func (d *domain) CreateRouter(ctx ConsoleContext, router entities.Router) (*entities.Router, error) {
	router.EnsureGVK()
	if err := d.k8sExtendedClient.ValidateStruct(ctx, &router.Router); err != nil {
		return nil, err
	}

	router.AccountName = ctx.accountName
	router.ClusterName = ctx.clusterName
	r, err := d.routerRepo.Create(ctx, &router)
	if err != nil {
		if d.routerRepo.ErrAlreadyExists(err) {
			return nil, fmt.Errorf("router with name '%s' already exists", router.Name)
		}
		return nil, err
	}

	if err := d.applyK8sResource(ctx, &r.Router); err != nil {
		return r, err
	}

	return r, nil
}

func (d *domain) DeleteRouter(ctx ConsoleContext, namespace string, name string) error {
	r, err := d.findRouter(ctx, namespace, name)
	if err != nil {
		return err
	}
	return d.deleteK8sResource(ctx, &r.Router)
}

func (d *domain) GetRouter(ctx ConsoleContext, namespace string, name string) (*entities.Router, error) {
	return d.findRouter(ctx, namespace, name)
}

func (d *domain) ListRouters(ctx ConsoleContext, namespace string) ([]*entities.Router, error) {
	return d.routerRepo.Find(ctx, repos.Query{Filter: repos.Filter{
		"clusterName":        ctx.clusterName,
		"accountName":        ctx.accountName,
		"metadata.namespace": namespace,
	}})
}

func (d *domain) UpdateRouter(ctx ConsoleContext, router entities.Router) (*entities.Router, error) {
	router.EnsureGVK()
	if err := d.k8sExtendedClient.ValidateStruct(ctx, &router.Router); err != nil {
		return nil, err
	}

	r, err := d.findRouter(ctx, router.Namespace, router.Name)
	if err != nil {
		return nil, err
	}

	status := r.Status
	r.Router = router.Router
	r.Status = status

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
		"accountName":        ctx.accountName,
		"clusterName":        ctx.clusterName,
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
