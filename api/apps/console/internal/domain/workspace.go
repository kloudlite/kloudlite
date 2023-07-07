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

// workspace:query

func (d *domain) findWorkspace(ctx ConsoleContext, namespace, name string) (*entities.Workspace, error) {
	ws, err := d.workspaceRepo.FindOne(ctx, repos.Filter{
		"accountName":        ctx.AccountName,
		"clusterName":        ctx.ClusterName,
		"metadata.namespace": namespace,
		"metadata.name":      name,
	})

	if err != nil {
		return nil, err
	}
	if ws == nil {
		return nil, fmt.Errorf("no workspace with name=%q, namespace=%q found", name, namespace)
	}
	return ws, nil
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

func (d *domain) CreateWorkspace(ctx ConsoleContext, ws entities.Workspace) (*entities.Workspace, error) {
	ws.EnsureGVK()
	if err := d.k8sExtendedClient.ValidateStruct(ctx, &ws.Workspace); err != nil {
		return nil, err
	}

	if err := d.canMutateResourcesInProject(ctx, ws.Namespace); err != nil {
		return nil, err
	}

	ws.IncrementRecordVersion()
	ws.AccountName = ctx.AccountName
	ws.ClusterName = ctx.ClusterName
	ws.SyncStatus = t.GenSyncStatus(t.SyncActionApply, ws.RecordVersion)

	nWs, err := d.workspaceRepo.Create(ctx, &ws)
	if err != nil {
		if d.workspaceRepo.ErrAlreadyExists(err) {
			// TODO: better insights into error, when it is being caused by duplicated indexes
			return nil, err
		}
		return nil, err
	}

	// if err := d.applyK8sResource(ctx, &corev1.Namespace{
	// 	TypeMeta: metav1.TypeMeta{APIVersion: "v1", Kind: "Namespace"},
	// 	ObjectMeta: metav1.ObjectMeta{
	// 		Name: ws.Spec.TargetNamespace,
	// 		Labels: map[string]string{
	// 			constants.EnvNameKey: ws.Name,
	// 		},
	// 	},
	// }, 0); err != nil {
	// 	return nil, err
	// }

	if err := d.applyK8sResource(ctx, &nWs.Workspace, nWs.RecordVersion); err != nil {
		return nil, err
	}

	return nWs, nil
}

func (d *domain) UpdateWorkspace(ctx ConsoleContext, ws entities.Workspace) (*entities.Workspace, error) {
	ws.EnsureGVK()
	if err := d.k8sExtendedClient.ValidateStruct(ctx, &ws.Workspace); err != nil {
		return nil, err
	}

	if err := d.canMutateResourcesInProject(ctx, ws.Namespace); err != nil {
		return nil, err
	}

	exWs, err := d.findWorkspace(ctx, ws.Namespace, ws.Name)
	if err != nil {
		return nil, err
	}

	if exWs.GetDeletionTimestamp() != nil {
		return nil, errAlreadyMarkedForDeletion("workspace", "", ws.Name)
	}

	exWs.Labels = ws.Labels
	exWs.Annotations = ws.Annotations
	exWs.Spec = ws.Spec
	exWs.SyncStatus = t.GenSyncStatus(t.SyncActionApply, exWs.RecordVersion)

	upWs, err := d.workspaceRepo.UpdateById(ctx, exWs.Id, exWs)
	if err != nil {
		return nil, err
	}

	if err := d.applyK8sResource(ctx, &upWs.Workspace, upWs.RecordVersion); err != nil {
		return nil, err
	}

	return upWs, nil
}

func (d *domain) DeleteWorkspace(ctx ConsoleContext, namespace, name string) error {
	ws, err := d.findWorkspace(ctx, namespace, name)
	if err != nil {
		return err
	}

	if err := d.canMutateResourcesInProject(ctx, ws.Namespace); err != nil {
		return err
	}

	ws.SyncStatus = t.GenSyncStatus(t.SyncActionDelete, ws.RecordVersion)
	if _, err := d.workspaceRepo.UpdateById(ctx, ws.Id, ws); err != nil {
		return err
	}

	return d.deleteK8sResource(ctx, &ws.Workspace)
}

func (d *domain) OnApplyWorkspaceError(ctx ConsoleContext, errMsg, namespace, name string) error {
	ws, err2 := d.findWorkspace(ctx, namespace, name)
	if err2 != nil {
		return err2
	}

	ws.SyncStatus.State = t.SyncStateErroredAtAgent
	ws.SyncStatus.LastSyncedAt = time.Now()
	ws.SyncStatus.Error = &errMsg
	_, err := d.workspaceRepo.UpdateById(ctx, ws.Id, ws)
	return err
}

func (d *domain) OnDeleteWorkspaceMessage(ctx ConsoleContext, ws entities.Workspace) error {
	exWs, err := d.findWorkspace(ctx, ws.Namespace, ws.Name)
	if err != nil {
		return err
	}

	if err := d.MatchRecordVersion(ws.Annotations, exWs.RecordVersion); err != nil {
		return err
	}

	return d.workspaceRepo.DeleteById(ctx, exWs.Id)
}

func (d *domain) OnUpdateWorkspaceMessage(ctx ConsoleContext, ws entities.Workspace) error {
	exWs, err := d.findWorkspace(ctx, ws.Namespace, ws.Name)
	if err != nil {
		return err
	}

	annotatedVersion, err := d.parseRecordVersionFromAnnotations(ws.Annotations)
	if err != nil {
		return d.resyncK8sResource(ctx, exWs.SyncStatus.Action, &exWs.Workspace, exWs.RecordVersion)
	}

	if annotatedVersion != exWs.RecordVersion {
		if err := d.resyncK8sResource(ctx, exWs.SyncStatus.Action, &exWs.Workspace, exWs.RecordVersion); err != nil {
			return err
		}
		return nil
	}

	exWs.CreationTimestamp = ws.CreationTimestamp
	exWs.Labels = ws.Labels
	exWs.Annotations = ws.Annotations
	exWs.Generation = ws.Generation

	exWs.Status = ws.Status

	exWs.SyncStatus.State = t.SyncStateReceivedUpdateFromAgent
	exWs.SyncStatus.RecordVersion = annotatedVersion
	exWs.SyncStatus.Error = nil
	exWs.SyncStatus.LastSyncedAt = time.Now()

	_, err = d.workspaceRepo.UpdateById(ctx, exWs.Id, exWs)
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
	}, 0); err != nil {
		return err
	}

	return d.resyncK8sResource(ctx, e.SyncStatus.Action, &e.Workspace, e.RecordVersion)
}
