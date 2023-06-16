package domain

import (
	"fmt"
	"time"

	"github.com/kloudlite/operator/pkg/constants"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"kloudlite.io/apps/console/internal/domain/entities"
	"kloudlite.io/pkg/repos"
	t "kloudlite.io/pkg/types"
)

// environment:query

func (d *domain) findWorkspace(ctx ConsoleContext, namespace, name string) (*entities.Workspace, error) {
	env, err := d.workspaceRepo.FindOne(ctx, repos.Filter{
		"accountName":        ctx.AccountName,
		"clusterName":        ctx.ClusterName,
		"metadata.namespace": namespace,
		"metadata.name":      name,
	})

	if err != nil {
		return nil, err
	}
	if env == nil {
		return nil, fmt.Errorf("no environment with name=%q found", name)
	}
	return env, nil
}

func (d *domain) GetWorkspace(ctx ConsoleContext, namespace, name string) (*entities.Workspace, error) {
	if err := d.canReadResourcesInProject(ctx, namespace); err != nil {
		return nil, err
	}
	return d.findWorkspace(ctx, namespace, name)
}

func (d *domain) ListWorkspaces(ctx ConsoleContext, namespace string, pq t.CursorPagination) (*repos.PaginatedRecord[*entities.Workspace], error) {
	if err := d.canReadResourcesInProject(ctx, namespace); err != nil {
		return nil, err
	}

	return d.workspaceRepo.FindPaginated(ctx, repos.Filter{
		"accountName":        ctx.AccountName,
		"clusterName":        ctx.ClusterName,
		"metadata.namespace": namespace,
	}, pq)
}

func (d *domain) findWorkspaceByTargetNs(ctx ConsoleContext, targetNs string) (*entities.Workspace, error) {
	w, err := d.workspaceRepo.FindOne(ctx, repos.Filter{
		"accountName":          ctx.AccountName,
		"clusterName":          ctx.ClusterName,
		"spec.targetNamespace": targetNs,
	})
	if err != nil {
		return nil, err
	}

	if w == nil {
		return nil, fmt.Errorf("no workspace found for target namespace %q", targetNs)
	}

	return w, nil
}

// mutations

func (d *domain) CreateWorkspace(ctx ConsoleContext, env entities.Workspace) (*entities.Workspace, error) {
	env.EnsureGVK()
	if err := d.k8sExtendedClient.ValidateStruct(ctx, &env.Env); err != nil {
		return nil, err
	}

	if err := d.canMutateResourcesInProject(ctx, env.Namespace); err != nil {
		return nil, err
	}

	env.AccountName = ctx.AccountName
	env.ClusterName = ctx.ClusterName
	env.Generation = 1
	env.SyncStatus = t.GenSyncStatus(t.SyncActionApply, env.Generation)

	nEnv, err := d.workspaceRepo.Create(ctx, &env)
	if err != nil {
		if d.workspaceRepo.ErrAlreadyExists(err) {
			return nil, fmt.Errorf(
				"environment with name %q, namespace=%q already exists",
				env.Name,
				env.Namespace,
			)
		}
		return nil, err
	}

	if err := d.applyK8sResource(ctx, &corev1.Namespace{
		TypeMeta: metav1.TypeMeta{APIVersion: "v1", Kind: "Namespace"},
		ObjectMeta: metav1.ObjectMeta{
			Name: env.Spec.TargetNamespace,
			Labels: map[string]string{
				constants.EnvNameKey: env.Name,
			},
		},
	}); err != nil {
		return nil, err
	}

	if err := d.applyK8sResource(ctx, &env.Env); err != nil {
		return nil, err
	}

	return nEnv, nil
}

func (d *domain) UpdateWorkspace(ctx ConsoleContext, env entities.Workspace) (*entities.Workspace, error) {
	env.EnsureGVK()
	if err := d.k8sExtendedClient.ValidateStruct(ctx, &env.Env); err != nil {
		return nil, err
	}

	if err := d.canMutateResourcesInProject(ctx, env.Namespace); err != nil {
		return nil, err
	}

	exEnv, err := d.findWorkspace(ctx, env.Namespace, env.Name)
	if err != nil {
		return nil, err
	}

	if exEnv.GetDeletionTimestamp() != nil {
		return nil, errAlreadyMarkedForDeletion("environment", "", env.Name)
	}

	exEnv.Spec = env.Spec
	exEnv.Generation += 1
	exEnv.SyncStatus = t.GenSyncStatus(t.SyncActionApply, exEnv.Generation)

	upEnv, err := d.workspaceRepo.UpdateById(ctx, exEnv.Id, exEnv)
	if err != nil {
		return nil, err
	}

	if err := d.applyK8sResource(ctx, &upEnv.Env); err != nil {
		return nil, err
	}

	return upEnv, nil
}

func (d *domain) DeleteWorkspace(ctx ConsoleContext, namespace, name string) error {
	ws, err := d.findWorkspace(ctx, namespace, name)
	if err != nil {
		return err
	}

	if err := d.canMutateResourcesInProject(ctx, ws.Namespace); err != nil {
		return err
	}

	ws.SyncStatus = t.GenSyncStatus(t.SyncActionDelete, ws.Generation)
	if _, err := d.workspaceRepo.UpdateById(ctx, ws.Id, ws); err != nil {
		return err
	}

	return d.deleteK8sResource(ctx, &ws.Env)
}

func (d *domain) OnApplyWorkspaceError(ctx ConsoleContext, errMsg, namespace, name string) error {
	ws, err2 := d.findWorkspace(ctx, namespace, name)
	if err2 != nil {
		return err2
	}

	ws.SyncStatus.Error = &errMsg
	_, err := d.workspaceRepo.UpdateById(ctx, ws.Id, ws)
	return err
}

func (d *domain) OnDeleteEnvironmentMessage(ctx ConsoleContext, env entities.Workspace) error {
	p, err := d.findWorkspace(ctx, env.Namespace, env.Name)
	if err != nil {
		return err
	}

	return d.workspaceRepo.DeleteById(ctx, p.Id)
}

func (d *domain) OnUpdateEnvironmentMessage(ctx ConsoleContext, env entities.Workspace) error {
	ws, err := d.findWorkspace(ctx, env.Namespace, env.Name)
	if err != nil {
		return err
	}

	ws.Status = env.Status
	ws.SyncStatus.Error = nil
	ws.SyncStatus.LastSyncedAt = time.Now()
	ws.SyncStatus.Generation = env.Generation
	ws.SyncStatus.State = t.ParseSyncState(env.Status.IsReady)

	_, err = d.workspaceRepo.UpdateById(ctx, ws.Id, ws)
	return err
}

func (d *domain) ResyncWorkspace(ctx ConsoleContext, namespace, name string) error {
	if err := d.canMutateResourcesInProject(ctx, namespace); err != nil {
		return err
	}

	e, err := d.findWorkspace(ctx, namespace, name)
	if err != nil {
		return err
	}

	if err := d.resyncK8sResource(ctx, t.SyncActionApply, &corev1.Namespace{
		TypeMeta: metav1.TypeMeta{APIVersion: "v1", Kind: "Namespace"},
		ObjectMeta: metav1.ObjectMeta{
			Name: e.Spec.TargetNamespace,
			Labels: map[string]string{
				constants.EnvNameKey: e.Name,
			},
		},
	}); err != nil {
		return err
	}

	return d.resyncK8sResource(ctx, e.SyncStatus.Action, &e.Env)
}
