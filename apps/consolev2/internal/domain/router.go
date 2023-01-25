package domain

import (
	"context"
	"kloudlite.io/apps/consolev2/internal/domain/entities"
	"kloudlite.io/pkg/errors"
	"kloudlite.io/pkg/repos"
)

func (d *domain) GetRouters(ctx context.Context, namespace string) ([]*entities.Router, error) {
	return d.routerRepo.Find(ctx, repos.Query{Filter: repos.Filter{"metadata.namespace": namespace}})
}

func (d *domain) GetRouter(ctx context.Context, namespace string, name string) (*entities.Router, error) {
	return d.routerRepo.FindOne(ctx, repos.Filter{"metadata.namespace": namespace})
}

func (d *domain) DeleteRouter(ctx context.Context, namespace, name string) (bool, error) {
	return true, d.routerRepo.DeleteOne(ctx, repos.Filter{"metadata.namespace": namespace, "metadata.name": name})
}

func (d *domain) CreateRouter(ctx context.Context, router entities.Router) (*entities.Router, error) {
	exists, err := d.routerRepo.Exists(ctx, repos.Filter{"metadata.namespace": router.Namespace, "metadata.name": router.Name})
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, errors.Newf("router %s already exists", router.Name)
	}

	nRouter, err := d.routerRepo.Create(ctx, &router)
	if err != nil {
		return nil, err
	}

	clusterId, err := d.getClusterForProject(ctx, nRouter.Spec.ProjectName)
	if err != nil {
		return nil, err
	}

	if err := d.workloadMessenger.SendAction(ActionApply, d.getDispatchKafkaTopic(clusterId), string(nRouter.Id), nRouter.Router); err != nil {
		return nil, err
	}
	return nRouter, nil
}

func (d *domain) UpdateRouter(ctx context.Context, router entities.Router) (bool, error) {
	exRouter, err := d.routerRepo.FindOne(ctx, repos.Filter{"metadata.namespace": router.Namespace, "metadata.name": router.Name})
	if err != nil {
		return false, err
	}
	if exRouter == nil {
		return false, errors.Newf("router %s does not exist", router.Name)
	}

	exRouter.Router = router.Router

	if _, err := d.routerRepo.UpdateById(ctx, exRouter.Id, exRouter); err != nil {
		return false, err
	}

	clusterId, err := d.getClusterForProject(ctx, exRouter.Spec.ProjectName)
	if err != nil {
		return false, err
	}

	if err := d.workloadMessenger.SendAction(ActionApply, d.getDispatchKafkaTopic(clusterId), string(exRouter.Id), exRouter.Router); err != nil {
		return false, err
	}
	return true, nil
}
