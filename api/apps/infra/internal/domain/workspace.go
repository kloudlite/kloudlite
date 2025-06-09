package domain

import (
	"github.com/kloudlite/api/apps/infra/internal/entities"
	fc "github.com/kloudlite/api/apps/infra/internal/entities/field-constants"
	"github.com/kloudlite/api/common"
	"github.com/kloudlite/api/common/fields"
	"github.com/kloudlite/api/pkg/errors"
	"github.com/kloudlite/api/pkg/repos"
	"github.com/kloudlite/operator/operators/resource-watcher/types"
)

func (d *domain) applyWorkspace(ctx InfraContext, ws *entities.Workspace) error {
	addTrackingId(&ws.Workspace, ws.Id)
	return d.resDispatcher.ApplyToTargetCluster(ctx, ws.DispatchAddr, &ws.Workspace, ws.RecordVersion)
}

func (d *domain) findWorkspace(ctx InfraContext, workmachineName string, clusterName string, name string) (*entities.Workspace, error) {
	ws, err := d.workspaceRepo.FindOne(ctx, repos.Filter{
		fc.AccountName:              ctx.AccountName,
		fc.MetadataName:             name,
		fc.ClusterName:              clusterName,
		fc.WorkspaceSpecWorkMachine: workmachineName,
	})
	if err != nil {
		return nil, errors.NewE(err)
	}
	if ws == nil {
		return nil, errors.Newf("no workspace with name=%q found", name)
	}
	return ws, nil
}

func (d *domain) CreateWorkspace(ctx InfraContext, workmachineName string, clusterName string, workspace entities.Workspace) (*entities.Workspace, error) {
	workspace.AccountName = ctx.AccountName
	workspace.ClusterName = clusterName

	workspace.DispatchAddr = &entities.DispatchAddr{
		AccountName: ctx.AccountName,
		ClusterName: clusterName}

	workspace.CreatedBy = common.CreatedOrUpdatedBy{
		UserId:    ctx.UserId,
		UserName:  ctx.UserName,
		UserEmail: ctx.UserEmail,
	}

	workspace.LastUpdatedBy = workspace.CreatedBy
	workspace.Spec.WorkMachine = workmachineName
	workspace.EnsureGVK()
	if err := d.k8sClient.ValidateObject(ctx, &workspace.Workspace); err != nil {
		return nil, errors.NewE(err)
	}

	ws, err := d.workspaceRepo.Create(ctx, &workspace)
	if err != nil {
		return nil, errors.NewE(err)
	}

	d.resourceEventPublisher.PublishResourceEvent(ctx, workspace.ClusterName, ResourceTypeWorkspace, ws.Name, PublishAdd)

	if err := d.applyWorkspace(ctx, ws); err != nil {
		return nil, errors.NewE(err)
	}

	return ws, nil
}

func (d *domain) UpdateWorkspace(ctx InfraContext, workmachineName string, clusterName string, workspace entities.Workspace) (*entities.Workspace, error) {
	patchForUpdate := repos.Document{
		fc.DisplayName: workspace.DisplayName,
		fc.LastUpdatedBy: common.CreatedOrUpdatedBy{
			UserId:    ctx.UserId,
			UserName:  ctx.UserName,
			UserEmail: ctx.UserEmail,
		},
	}

	upWorkspace, err := d.workspaceRepo.Patch(
		ctx,
		repos.Filter{
			fc.AccountName:              ctx.AccountName,
			fc.MetadataName:             workspace.Name,
			fields.ClusterName:          clusterName,
			fc.WorkspaceSpecWorkMachine: workmachineName,
		},
		patchForUpdate,
	)
	if err != nil {
		return nil, errors.NewE(err)
	}

	d.resourceEventPublisher.PublishResourceEvent(ctx, workspace.ClusterName, ResourceTypeWorkspace, upWorkspace.Name, PublishUpdate)

	if err := d.applyWorkspace(ctx, upWorkspace); err != nil {
		return nil, errors.NewE(err)
	}

	return upWorkspace, nil
}

func (d *domain) UpdateWorkspaceStatus(ctx InfraContext, workmachineName string, clusterName string, status bool, name string) (bool, error) {
	workspaceStatus := "OFF"
	if status {
		workspaceStatus = "ON"
	}

	patchForUpdate := repos.Document{
		fc.WorkspaceSpecState: workspaceStatus,
		fc.LastUpdatedBy: common.CreatedOrUpdatedBy{
			UserId:    ctx.UserId,
			UserName:  ctx.UserName,
			UserEmail: ctx.UserEmail,
		},
	}

	upWorkspace, err := d.workspaceRepo.Patch(
		ctx,
		repos.Filter{
			fc.AccountName:              ctx.AccountName,
			fc.MetadataName:             name,
			fields.ClusterName:          clusterName,
			fc.WorkspaceSpecWorkMachine: workmachineName,
		},
		patchForUpdate,
	)

	if err != nil {
		return false, errors.NewE(err)
	}

	d.resourceEventPublisher.PublishResourceEvent(ctx, clusterName, ResourceTypeWorkspace, upWorkspace.Name, PublishUpdate)

	if err := d.applyWorkspace(ctx, upWorkspace); err != nil {
		return false, errors.NewE(err)
	}

	return true, nil
}

func (d *domain) DeleteWorkspace(ctx InfraContext, workmachineName string, clusterName string, name string) error {
	uws, err := d.workspaceRepo.Patch(
		ctx,
		repos.Filter{
			fields.ClusterName:          clusterName,
			fields.AccountName:          ctx.AccountName,
			fields.MetadataName:         name,
			fc.WorkspaceSpecWorkMachine: workmachineName,
		},
		common.PatchForMarkDeletion(),
	)

	if err != nil {
		return errors.NewE(err)
	}
	d.resourceEventPublisher.PublishResourceEvent(ctx, clusterName, ResourceTypeWorkspace, uws.Name, PublishUpdate)
	return d.resDispatcher.DeleteFromTargetCluster(ctx, uws.DispatchAddr, &uws.Workspace)
}

func (d *domain) GetWorkspace(ctx InfraContext, workmachineName string, clusterName string, name string) (*entities.Workspace, error) {
	return d.findWorkspace(ctx, workmachineName, clusterName, name)
}

func (d *domain) ListWorkspaces(ctx InfraContext, workmachineName string, clusterName string, search map[string]repos.MatchFilter, pagination repos.CursorPagination) (*repos.PaginatedRecord[*entities.Workspace], error) {
	filter := repos.Filter{
		fc.AccountName:              ctx.AccountName,
		fc.WorkspaceSpecWorkMachine: workmachineName,
		fc.ClusterName:              clusterName,
	}
	return d.workspaceRepo.FindPaginated(ctx, d.workspaceRepo.MergeMatchFilters(filter, search), pagination)
}

func (d *domain) OnWorkspaceDeleteMessage(ctx InfraContext, clusterName string, workspace entities.Workspace) error {
	err := d.workspaceRepo.DeleteOne(
		ctx,
		repos.Filter{
			fields.AccountName:          ctx.AccountName,
			fields.ClusterName:          clusterName,
			fc.MetadataName:             workspace.Name,
			fc.WorkspaceSpecWorkMachine: workspace.Spec.WorkMachine,
		},
	)
	if err != nil {
		return errors.NewE(err)
	}
	d.resourceEventPublisher.PublishResourceEvent(ctx, clusterName, ResourceTypeWorkspace, workspace.Name, PublishDelete)
	return nil
}

func (d *domain) OnWorkspaceUpdateMessage(ctx InfraContext, clusterName string, workspace entities.Workspace, status types.ResourceStatus, opts UpdateAndDeleteOpts) error {
	ws, err := d.findWorkspace(ctx, workspace.Spec.WorkMachine, clusterName, workspace.Name)
	if err != nil {
		return errors.NewE(err)
	}

	if ws == nil {
		workspace.AccountName = ctx.AccountName
		workspace.ClusterName = clusterName

		workspace.CreatedBy = common.CreatedOrUpdatedBy{
			UserId:    ctx.UserId,
			UserName:  ctx.UserName,
			UserEmail: ctx.UserEmail,
		}
		workspace.LastUpdatedBy = workspace.CreatedBy

		ws, err = d.workspaceRepo.Create(ctx, &workspace)
		if err != nil {
			return errors.NewE(err)
		}
	}

	upWs, err := d.workspaceRepo.PatchById(
		ctx,
		ws.Id,
		common.PatchForSyncFromAgent(
			&workspace,
			workspace.RecordVersion,
			status,
			common.PatchOpts{
				MessageTimestamp: opts.MessageTimestamp,
			}))
	if err != nil {
		return errors.NewE(err)
	}
	d.resourceEventPublisher.PublishResourceEvent(ctx, clusterName, ResourceTypeWorkspace, upWs.Name, PublishUpdate)
	return nil
}
