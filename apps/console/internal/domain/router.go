package domain

import (
	"context"
	"fmt"
	"kloudlite.io/apps/console/internal/domain/entities"
	op_crds "kloudlite.io/apps/console/internal/domain/op-crds"
	"kloudlite.io/pkg/repos"
)

func (d *domain) GetRouter(ctx context.Context, routerID repos.ID) (*entities.Router, error) {
	router, err := d.routerRepo.FindById(ctx, routerID)

	if err = mongoError(err, "router not found"); err != nil {
		return nil, err
	}

	if err = d.checkProjectAccess(ctx, router.ProjectId, ReadProject); err != nil {
		return nil, err
	}

	return router, nil
}

func (d *domain) GetRouters(ctx context.Context, projectID repos.ID) ([]*entities.Router, error) {

	if err := d.checkProjectAccess(ctx, projectID, ReadProject); err != nil {
		return nil, err
	}

	routers, err := d.routerRepo.Find(ctx, repos.Query{
		Filter: repos.Filter{
			"project_id": projectID,
		},
	})
	if err != nil {
		return nil, err
	}

	return routers, nil
}

func (d *domain) OnUpdateRouter(ctx context.Context, response *op_crds.StatusUpdate) error {
	one, err := d.routerRepo.FindOne(ctx, repos.Filter{
		"id": response.Metadata.ResourceId,
	})
	if err != nil {
		return err
	}
	if one == nil {
		return fmt.Errorf("router not found")
	}
	newStatus := one.Status
	if response.IsReady {
		one.Status = entities.RouteStateLive
	}
	shouldUpdate := newStatus != one.Status
	one.Conditions = response.ChildConditions
	_, err = d.routerRepo.UpdateById(ctx, one.Id, one)
	if shouldUpdate {
		err = d.notifier.Notify(one.Id)
		if err != nil {
			return err
		}
	}
	return err
}

func (d *domain) OnDeleteRouter(ctx context.Context, response *op_crds.StatusUpdate) error {
	return d.routerRepo.DeleteById(ctx, repos.ID(response.Metadata.ResourceId))
}

func (d *domain) CreateRouter(ctx context.Context, projectId repos.ID, routerName string, domains []string, routes []*entities.Route) (*entities.Router, error) {

	if err := d.checkProjectAccess(ctx, projectId, UpdateProject); err != nil {
		return nil, err
	}

	prj, err := d.projectRepo.FindById(ctx, projectId)
	if err != nil {
		return nil, err
	}
	if prj == nil {
		return nil, fmt.Errorf("project not found")
	}
	create, err := d.routerRepo.Create(ctx, &entities.Router{
		ProjectId: projectId,
		Name:      routerName,
		Namespace: prj.Name,
		Domains:   domains,
		Routes:    routes,
	})
	if err != nil {
		return nil, err
	}

	err = d.workloadMessenger.SendAction("apply", string(create.Id), &op_crds.Router{
		APIVersion: op_crds.RouterAPIVersion,
		Kind:       op_crds.RouterKind,
		Metadata: op_crds.RouterMetadata{
			Name:      string(create.Id),
			Namespace: create.Namespace,
			Labels: map[string]string{
				"kloudlite.io/account-ref": string(prj.AccountId),
			},
		},
		Spec: op_crds.RouterSpec{
			Https: struct {
				Enabled       bool `json:"enabled"`
				ForceRedirect bool `json:"forceRedirect"`
			}(struct {
				Enabled       bool
				ForceRedirect bool
			}{Enabled: true, ForceRedirect: true}),
			Domains: create.Domains,
			Routes: func() []op_crds.Route {
				i := make([]op_crds.Route, 0)
				for _, r := range create.Routes {
					i = append(i, op_crds.Route{
						Path: r.Path,
						App:  r.AppName,
						Port: r.Port,
					})
				}
				return i
			}(),
		},
	})
	if err != nil {
		return nil, err
	}
	return create, nil
}
func (d *domain) UpdateRouter(ctx context.Context, id repos.ID, domains []string, entries []*entities.Route) (bool, error) {
	router, err := d.routerRepo.FindById(ctx, id)
	if err = mongoError(err, "router not found"); err != nil {
		return false, err
	}
	if err = d.checkProjectAccess(ctx, router.ProjectId, UpdateProject); err != nil {
		return false, err
	}
	prj, err := d.projectRepo.FindById(ctx, router.ProjectId)
	if err != nil {
		return false, err
	}
	if domains != nil {
		router.Domains = domains
	}
	if entries != nil {
		router.Routes = entries
	}
	_, err = d.routerRepo.UpdateById(ctx, id, router)
	if err != nil {
		return false, err
	}
	err = d.workloadMessenger.SendAction("apply", string(router.Id), &op_crds.Router{
		APIVersion: op_crds.RouterAPIVersion,
		Kind:       op_crds.RouterKind,
		Metadata: op_crds.RouterMetadata{
			Name:      string(router.Id),
			Namespace: router.Namespace,
			Labels: map[string]string{
				"kloudlite.io/account-ref": string(prj.AccountId),
			},
		},
		Spec: op_crds.RouterSpec{
			Https: struct {
				Enabled       bool `json:"enabled"`
				ForceRedirect bool `json:"forceRedirect"`
			}(struct {
				Enabled       bool
				ForceRedirect bool
			}{Enabled: true, ForceRedirect: true}),
			Domains: router.Domains,
			Routes: func() []op_crds.Route {
				i := make([]op_crds.Route, 0)
				for _, r := range router.Routes {
					i = append(i, op_crds.Route{
						Path: r.Path,
						App:  r.AppName,
						Port: r.Port,
					})
				}
				return i
			}(),
		},
	})
	if err != nil {
		return false, err
	}
	if err != nil {
		return false, err
	}
	return true, nil
}
func (d *domain) DeleteRouter(ctx context.Context, routerID repos.ID) (bool, error) {
	r, err := d.routerRepo.FindById(ctx, routerID)
	if err = mongoError(err, "router not found"); err != nil {
		return false, err
	}

	if err = d.checkProjectAccess(ctx, r.ProjectId, UpdateProject); err != nil {
		return false, err
	}

	err = d.routerRepo.DeleteById(ctx, routerID)
	if err != nil {
		return false, err
	}

	err = d.workloadMessenger.SendAction("delete", string(r.Id), &op_crds.Router{
		APIVersion: op_crds.RouterAPIVersion,
		Kind:       op_crds.RouterKind,
		Metadata: op_crds.RouterMetadata{
			Name:      string(r.Id),
			Namespace: r.Namespace,
		},
	})

	if err != nil {
		return false, err
	}

	return true, nil
}
