package domain

import (
	"github.com/kloudlite/api/apps/console/internal/entities"
	iamT "github.com/kloudlite/api/apps/iam/types"
	"github.com/kloudlite/api/pkg/errors"
	fn "github.com/kloudlite/api/pkg/functions"
	"github.com/kloudlite/api/pkg/repos"
)

func (d *domain) findEnvironment(ctx ConsoleContext, namespace, name string) (*entities.Workspace, error) {
	ws, err := d.workspaceRepo.FindOne(ctx, repos.Filter{
		"accountName":        ctx.AccountName,
		"clusterName":        ctx.ClusterName,
		"metadata.namespace": namespace,
		"metadata.name":      name,
		"spec.isEnvironment": true,
	})

	if err != nil {
		return nil, err
	}
	if ws == nil {
		return nil, errors.Newf("no environment with name=%q, namespace=%q found", name, namespace)
	}
	return ws, nil
}

func (d *domain) ListEnvironments(ctx ConsoleContext, namespace string, search map[string]repos.MatchFilter, pq repos.CursorPagination) (*repos.PaginatedRecord[*entities.Workspace], error) {
	if err := d.canReadResourcesInProject(ctx, namespace); err != nil {
		return nil, err
	}

	if search == nil {
		search = map[string]repos.MatchFilter{}
	}

	search["spec.isEnvironment"] = repos.MatchFilter{
		MatchType: repos.MatchTypeExact,
		Exact:     true,
	}

	return d.listWorkspaces(ctx, namespace, search, pq)
}

func (d *domain) GetEnvironment(ctx ConsoleContext, namespace, name string) (*entities.Workspace, error) {
	return d.findEnvironment(ctx, namespace, name)
}

func (d *domain) CreateEnvironment(ctx ConsoleContext, ws entities.Workspace) (*entities.Workspace, error) {
	if err := d.canMutateResourcesInProject(ctx, ws.Namespace); err != nil {
		return nil, err
	}

	p, err := d.findProjectByTargetNs(ctx, ws.Namespace)
	if err != nil {
		return nil, err
	}

	if err := d.checkProjectAccess(ctx, p.Name, iamT.CreateEnvironment); err != nil {
		return nil, err
	}

	ws.ProjectName = p.Name
	ws.Spec.IsEnvironment = fn.New(true)

	return d.createWorkspace(ctx, ws)
}

func (d *domain) UpdateEnvironment(ctx ConsoleContext, ws entities.Workspace) (*entities.Workspace, error) {
	if err := d.canMutateResourcesInProject(ctx, ws.Namespace); err != nil {
		return nil, err
	}

	return d.updateWorkspace(ctx, ws)
}

func (d *domain) DeleteEnvironment(ctx ConsoleContext, namespace, name string) error {
	if err := d.canMutateResourcesInProject(ctx, namespace); err != nil {
		return err
	}

	return d.deleteWorkspace(ctx, namespace, name)
}

func (d *domain) ResyncEnvironment(ctx ConsoleContext, namespace, name string) error {
	if err := d.canMutateResourcesInProject(ctx, namespace); err != nil {
		return err
	}

	return d.resyncWorkspace(ctx, namespace, name)
}
