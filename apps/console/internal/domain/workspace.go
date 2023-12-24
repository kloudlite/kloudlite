package domain

import (
	"github.com/kloudlite/api/pkg/errors"
	"time"

	iamT "github.com/kloudlite/api/apps/iam/types"
	"github.com/kloudlite/api/common"
	"github.com/kloudlite/api/grpc-interfaces/kloudlite.io/rpc/iam"

	"github.com/kloudlite/api/constants"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/kloudlite/api/apps/console/internal/entities"
	fn "github.com/kloudlite/api/pkg/functions"
	"github.com/kloudlite/api/pkg/repos"
	t "github.com/kloudlite/api/pkg/types"
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
		return nil, errors.NewE(err)
	}
	if ws == nil {
		return nil, errors.Newf("no workspace with name=%q, namespace=%q found", name, namespace)
	}
	return ws, nil
}

func (d *domain) GetWorkspace(ctx ConsoleContext, namespace, name string) (*entities.Workspace, error) {
	if err := d.canReadResourcesInProject(ctx, namespace); err != nil {
		return nil, errors.NewE(err)
	}
	return d.findWorkspace(ctx, namespace, name)
}

func (d *domain) ListWorkspaces(ctx ConsoleContext, namespace string, search map[string]repos.MatchFilter, pq repos.CursorPagination) (*repos.PaginatedRecord[*entities.Workspace], error) {
	if err := d.canReadResourcesInProject(ctx, namespace); err != nil {
		return nil, errors.NewE(err)
	}

	return d.listWorkspaces(ctx, namespace, search, pq)
}

func (d *domain) listWorkspaces(ctx ConsoleContext, namespace string, search map[string]repos.MatchFilter, pq repos.CursorPagination) (*repos.PaginatedRecord[*entities.Workspace], error) {
	filter := repos.Filter{
		"accountName":        ctx.AccountName,
		"clusterName":        ctx.ClusterName,
		"metadata.namespace": namespace,
	}

	if _, ok := search["spec.isEnvironment"]; !ok {
		filter["$or"] = []map[string]any{
			{"spec.isEnvironment": map[string]any{"$exists": false}},
			{"spec.isEnvironment": false},
		}
	}

	return d.workspaceRepo.FindPaginated(ctx, d.workspaceRepo.MergeMatchFilters(filter, search), pq)
}

func (d *domain) findWorkspaceByTargetNs(ctx ConsoleContext, targetNs string) (*entities.Workspace, error) {
	w, err := d.workspaceRepo.FindOne(ctx, repos.Filter{
		"accountName":          ctx.AccountName,
		"clusterName":          ctx.ClusterName,
		"spec.targetNamespace": targetNs,
	})
	if err != nil {
		return nil, errors.NewE(err)
	}

	if w == nil {
		return nil, errors.Newf("no workspace found for target namespace %q", targetNs)
	}

	return w, nil
}

// mutations

func (d *domain) CreateWorkspace(ctx ConsoleContext, ws entities.Workspace) (*entities.Workspace, error) {
	if err := d.canMutateResourcesInProject(ctx, ws.Namespace); err != nil {
		return nil, errors.NewE(err)
	}

	p, err := d.findProjectByTargetNs(ctx, ws.Namespace)
	if err != nil {
		return nil, errors.NewE(err)
	}

	if err := d.checkProjectAccess(ctx, p.Name, iamT.CreateWorkspace); err != nil {
		return nil, errors.NewE(err)
	}

	if ws.Spec.IsEnvironment != nil {
		return nil, errors.Newf(".Spec.IsEnvironment can not be set, to create environments, use CreateEnvironment")
	}

	ws.ProjectName = p.Name
	return d.createWorkspace(ctx, ws)
}

func (d *domain) createWorkspace(ctx ConsoleContext, ws entities.Workspace) (*entities.Workspace, error) {
	if ws.ProjectName == "" {
		return nil, errors.Newf(".ProjectName can not be empty")
	}

	ws.EnsureGVK()
	if err := d.k8sClient.ValidateObject(ctx, &ws.Workspace); err != nil {
		return nil, errors.NewE(err)
	}

	ws.IncrementRecordVersion()

	ws.CreatedBy = common.CreatedOrUpdatedBy{
		UserId:    ctx.UserId,
		UserName:  ctx.UserName,
		UserEmail: ctx.UserEmail,
	}
	ws.LastUpdatedBy = ws.CreatedBy

	ws.AccountName = ctx.AccountName
	ws.ClusterName = ctx.ClusterName
	ws.SyncStatus = t.GenSyncStatus(t.SyncActionApply, ws.RecordVersion)

	nWs, err := d.workspaceRepo.Create(ctx, &ws)
	if err != nil {
		if d.workspaceRepo.ErrAlreadyExists(err) {
			// TODO: better insights into error, when it is being caused by duplicated indexes
			return nil, errors.NewE(err)
		}
		return nil, errors.NewE(err)
	}

	if _,err:=d.iamClient.AddMembership(ctx, &iam.AddMembershipIn{
		UserId:       string(ctx.UserId),
		ResourceType: string(iamT.ResourceWorkspace),
		ResourceRef:  iamT.NewResourceRef(ctx.AccountName, iamT.ResourceWorkspace, nWs.Name),
		Role:         string(iamT.RoleResourceOwner),
	}); err != nil {
		d.logger.Errorf(err, "error while adding membership")
	}

	if err := d.applyK8sResource(ctx, &nWs.Workspace, nWs.RecordVersion); err != nil {
		return nil, errors.NewE(err)
	}

	return nWs, nil
}

func (d *domain) UpdateWorkspace(ctx ConsoleContext, ws entities.Workspace) (*entities.Workspace, error) {
	if err := d.canMutateResourcesInProject(ctx, ws.Namespace); err != nil {
		return nil, errors.NewE(err)
	}

	return d.updateWorkspace(ctx, ws)
}

func (d *domain) updateWorkspace(ctx ConsoleContext, ws entities.Workspace) (*entities.Workspace, error) {
	ws.EnsureGVK()
	if err := d.k8sClient.ValidateObject(ctx, &ws.Workspace); err != nil {
		return nil, errors.NewE(err)
	}

	exWs, err := d.findWorkspace(ctx, ws.Namespace, ws.Name)
	if err != nil {
		return nil, errors.NewE(err)
	}

	if exWs.GetDeletionTimestamp() != nil {
		return nil, errAlreadyMarkedForDeletion("workspace", "", ws.Name)
	}

	exWs.IncrementRecordVersion()
	exWs.LastUpdatedBy = common.CreatedOrUpdatedBy{
		UserId:    ctx.UserId,
		UserName:  ctx.UserName,
		UserEmail: ctx.UserEmail,
	}

	exWs.DisplayName = ws.DisplayName
	exWs.Labels = ws.Labels
	exWs.Annotations = ws.Annotations

	exWs.SyncStatus = t.GenSyncStatus(t.SyncActionApply, exWs.RecordVersion)

	upWs, err := d.workspaceRepo.UpdateById(ctx, exWs.Id, exWs)
	if err != nil {
		return nil, errors.NewE(err)
	}

	if err := d.applyK8sResource(ctx, &upWs.Workspace, upWs.RecordVersion); err != nil {
		return nil, errors.NewE(err)
	}

	return upWs, nil
}

func (d *domain) DeleteWorkspace(ctx ConsoleContext, namespace, name string) error {
	if err := d.canMutateResourcesInProject(ctx, namespace); err != nil {
		return errors.NewE(err)
	}

	return d.deleteWorkspace(ctx, namespace, name)
}

func (d *domain) deleteWorkspace(ctx ConsoleContext, namespace string, name string) error {
	ws, err := d.findWorkspace(ctx, namespace, name)
	if err != nil {
		return errors.NewE(err)
	}

	ws.MarkedForDeletion = fn.New(true)
	ws.SyncStatus = t.GenSyncStatus(t.SyncActionDelete, ws.RecordVersion)
	if _, err := d.workspaceRepo.UpdateById(ctx, ws.Id, ws); err != nil {
		return errors.NewE(err)
	}

	if err := d.deleteK8sResource(ctx, &ws.Workspace); err != nil {
		return errors.NewE(err)
	}

	if err := d.deleteK8sResource(ctx, &corev1.Namespace{
		TypeMeta: metav1.TypeMeta{APIVersion: "v1", Kind: "Namespace"},
		ObjectMeta: metav1.ObjectMeta{
			Name: ws.Spec.TargetNamespace,
		},
	}); err != nil {
		return errors.NewE(err)
	}

	return nil
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
	return errors.NewE(err)
}

func (d *domain) OnDeleteWorkspaceMessage(ctx ConsoleContext, ws entities.Workspace) error {
	exWs, err := d.findWorkspace(ctx, ws.Namespace, ws.Name)
	if err != nil {
		return errors.NewE(err)
	}

	if err := d.MatchRecordVersion(ws.Annotations, exWs.RecordVersion); err != nil {
		return errors.NewE(err)
	}

	return d.workspaceRepo.DeleteById(ctx, exWs.Id)
}

func (d *domain) OnUpdateWorkspaceMessage(ctx ConsoleContext, ws entities.Workspace) error {
	exWs, err := d.findWorkspace(ctx, ws.Namespace, ws.Name)
	if err != nil {
		return errors.NewE(err)
	}

	annotatedVersion, err := d.parseRecordVersionFromAnnotations(ws.Annotations)
	if err != nil {
		return d.resyncK8sResource(ctx, exWs.SyncStatus.Action, &exWs.Workspace, exWs.RecordVersion)
	}

	if annotatedVersion != exWs.RecordVersion {
		if err := d.resyncK8sResource(ctx, exWs.SyncStatus.Action, &exWs.Workspace, exWs.RecordVersion); err != nil {
			return errors.NewE(err)
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
	return errors.NewE(err)
}

func (d *domain) ResyncWorkspace(ctx ConsoleContext, namespace, name string) error {
	if err := d.canMutateResourcesInProject(ctx, namespace); err != nil {
		return errors.NewE(err)
	}

	return d.resyncWorkspace(ctx, namespace, name)
}

func (d *domain) resyncWorkspace(ctx ConsoleContext, namespace string, name string) error {
	e, err := d.findWorkspace(ctx, namespace, name)
	if err != nil {
		return errors.NewE(err)
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
		return errors.NewE(err)
	}

	return d.resyncK8sResource(ctx, e.SyncStatus.Action, &e.Workspace, e.RecordVersion)
}
