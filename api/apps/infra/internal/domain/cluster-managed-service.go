package domain

import (
	iamT "github.com/kloudlite/api/apps/iam/types"
	"github.com/kloudlite/api/apps/infra/internal/entities"
	fc "github.com/kloudlite/api/apps/infra/internal/entities/field-constants"
	"github.com/kloudlite/api/common"
	"github.com/kloudlite/api/common/fields"
	"github.com/kloudlite/api/pkg/errors"
	"github.com/kloudlite/api/pkg/repos"
	t "github.com/kloudlite/api/pkg/types"
	"github.com/kloudlite/operator/operators/resource-watcher/types"
)

func (d *domain) ListClusterManagedServices(ctx InfraContext, search map[string]repos.MatchFilter, pagination repos.CursorPagination) (*repos.PaginatedRecord[*entities.ClusterManagedService], error) {
	if err := d.canPerformActionInAccount(ctx, iamT.ListClusterManagedServices); err != nil {
		return nil, errors.NewE(err)
	}

	f := repos.Filter{
		fields.AccountName: ctx.AccountName,
	}

	pr, err := d.clusterManagedServiceRepo.FindPaginated(ctx, d.secretRepo.MergeMatchFilters(f, search), pagination)
	if err != nil {
		return nil, errors.NewE(err)
	}

	return pr, nil
}

func (d *domain) findClusterManagedService(ctx InfraContext, name string) (*entities.ClusterManagedService, error) {
	cmsvc, err := d.clusterManagedServiceRepo.FindOne(ctx, repos.Filter{
		fields.AccountName:  ctx.AccountName,
		fields.MetadataName: name,
	})
	if err != nil {
		return nil, errors.NewE(err)
	}

	if cmsvc == nil {
		return nil, errors.Newf("cmsvc with name %q not found", name)
	}
	return cmsvc, nil
}

func (d *domain) GetClusterManagedService(ctx InfraContext, serviceName string) (*entities.ClusterManagedService, error) {
	if err := d.canPerformActionInAccount(ctx, iamT.GetClusterManagedService); err != nil {
		return nil, errors.NewE(err)
	}

	c, err := d.findClusterManagedService(ctx, serviceName)
	if err != nil {
		return nil, errors.NewE(err)
	}

	return c, nil
}

func (d *domain) applyClusterManagedService(ctx InfraContext, cmsvc *entities.ClusterManagedService) error {
	addTrackingId(&cmsvc.ClusterManagedService, cmsvc.Id)
	return d.resDispatcher.ApplyToTargetCluster(ctx, cmsvc.ClusterName, &cmsvc.ClusterManagedService, cmsvc.RecordVersion)
}

func (d *domain) CreateClusterManagedService(ctx InfraContext, cmsvc entities.ClusterManagedService) (*entities.ClusterManagedService, error) {
	if err := d.canPerformActionInAccount(ctx, iamT.CreateClusterManagedService); err != nil {
		return nil, errors.NewE(err)
	}

	cmsvc.IncrementRecordVersion()

	cmsvc.CreatedBy = common.CreatedOrUpdatedBy{
		UserId:    ctx.UserId,
		UserName:  ctx.UserName,
		UserEmail: ctx.UserEmail,
	}

	cmsvc.LastUpdatedBy = cmsvc.CreatedBy

	existing, err := d.clusterManagedServiceRepo.FindOne(ctx, repos.Filter{
		fields.AccountName:  ctx.AccountName,
		fields.MetadataName: cmsvc.Name,
	})
	if err != nil {
		return nil, errors.NewE(err)
	}

	if existing != nil {
		return nil, errors.Newf("cluster managed service with name %q already exists", cmsvc.ClusterName)
	}

	cmsvc.AccountName = ctx.AccountName
	cmsvc.SyncStatus = t.GenSyncStatus(t.SyncActionApply, cmsvc.RecordVersion)

	cmsvc.EnsureGVK()

	if err := d.k8sClient.ValidateObject(ctx, &cmsvc.ClusterManagedService); err != nil {
		return nil, errors.NewE(err)
	}

	ncms, err := d.clusterManagedServiceRepo.Create(ctx, &cmsvc)
	if err != nil {
		return nil, errors.NewE(err)
	}

	if err := d.applyClusterManagedService(ctx, &cmsvc); err != nil {
		return nil, errors.NewE(err)
	}

	d.resourceEventPublisher.PublishResourceEvent(ctx, cmsvc.ClusterName, ResourceTypeClusterManagedService, ncms.Name, PublishAdd)

	return ncms, nil
}

func (d *domain) UpdateClusterManagedService(ctx InfraContext, cmsvc entities.ClusterManagedService) (*entities.ClusterManagedService, error) {
	if err := d.canPerformActionInAccount(ctx, iamT.UpdateClusterManagedService); err != nil {
		return nil, errors.NewE(err)
	}

	cmsvc.EnsureGVK()
	if err := d.k8sClient.ValidateObject(ctx, &cmsvc.ClusterManagedService); err != nil {
		return nil, errors.NewE(err)
	}

	patchForUpdate := common.PatchForUpdate(
		ctx,
		&cmsvc,
		common.PatchOpts{
			XPatch: repos.Document{
				fc.ClusterManagedServiceSpecMsvcSpec: cmsvc.Spec,
			},
		})

	ucmsvc, err := d.clusterManagedServiceRepo.Patch(ctx, repos.Filter{fields.AccountName: ctx.AccountName, fields.MetadataName: cmsvc.Name}, patchForUpdate)
	if err != nil {
		return nil, errors.NewE(err)
	}

	d.resourceEventPublisher.PublishResourceEvent(ctx, ucmsvc.ClusterName, ResourceTypeClusterManagedService, ucmsvc.Name, PublishUpdate)

	if err := d.applyClusterManagedService(ctx, ucmsvc); err != nil {
		return nil, errors.NewE(err)
	}

	return ucmsvc, nil
}

func (d *domain) DeleteClusterManagedService(ctx InfraContext, name string) error {
	if err := d.canPerformActionInAccount(ctx, iamT.DeleteClusterManagedService); err != nil {
		return errors.NewE(err)
	}

	ucmsvc, err := d.clusterManagedServiceRepo.Patch(ctx, repos.Filter{fields.AccountName: ctx.AccountName, fields.MetadataName: name}, common.PatchForMarkDeletion())
	if err != nil {
		return errors.NewE(err)
	}

	d.resourceEventPublisher.PublishResourceEvent(ctx, ucmsvc.ClusterName, ResourceTypeClusterManagedService, ucmsvc.Name, PublishUpdate)

	return d.resDispatcher.DeleteFromTargetCluster(ctx, ucmsvc.ClusterName, &ucmsvc.ClusterManagedService)
}

func (d *domain) OnClusterManagedServiceApplyError(ctx InfraContext, clusterName, name, errMsg string, opts UpdateAndDeleteOpts) error {
	ucmsvc, err := d.clusterManagedServiceRepo.Patch(
		ctx,
		repos.Filter{
			fields.ClusterName:  clusterName,
			fields.AccountName:  ctx.AccountName,
			fields.MetadataName: name,
		},
		common.PatchForErrorFromAgent(
			errMsg,
			common.PatchOpts{
				MessageTimestamp: opts.MessageTimestamp,
			},
		),
	)
	if err != nil {
		return errors.NewE(err)
	}

	d.resourceEventPublisher.PublishResourceEvent(ctx, clusterName, ResourceTypeClusterManagedService, ucmsvc.Name, PublishDelete)
	return errors.NewE(err)
}

func (d *domain) OnClusterManagedServiceDeleteMessage(ctx InfraContext, clusterName string, service entities.ClusterManagedService) error {
	err := d.clusterManagedServiceRepo.DeleteOne(
		ctx,
		repos.Filter{
			fields.ClusterName:  clusterName,
			fields.AccountName:  ctx.AccountName,
			fields.MetadataName: service.Name,
		},
	)
	if err != nil {
		return errors.NewE(err)
	}
	d.resourceEventPublisher.PublishResourceEvent(ctx, clusterName, ResourceTypeClusterManagedService, service.Name, PublishDelete)
	return err
}

func (d *domain) OnClusterManagedServiceUpdateMessage(ctx InfraContext, clusterName string, service entities.ClusterManagedService, status types.ResourceStatus, opts UpdateAndDeleteOpts) error {
	xService, err := d.findClusterManagedService(ctx, service.Name)
	if err != nil {
		return errors.NewE(err)
	}

	if xService == nil {
		return errors.Newf("no cluster manage service found")
	}

	if _, err := d.matchRecordVersion(service.Annotations, xService.RecordVersion); err != nil {
		return d.resyncToTargetCluster(ctx, xService.SyncStatus.Action, clusterName, xService, xService.RecordVersion)
	}

	recordVersion, err := d.matchRecordVersion(service.Annotations, xService.RecordVersion)
	if err != nil {
		return errors.NewE(err)
	}

	ucmsvc, err := d.clusterManagedServiceRepo.PatchById(
		ctx,
		xService.Id,
		common.PatchForSyncFromAgent(&service, recordVersion, status, common.PatchOpts{
			MessageTimestamp: opts.MessageTimestamp,
		}))
	if err != nil {
		return errors.NewE(err)
	}

	d.resourceEventPublisher.PublishResourceEvent(ctx, clusterName, ResourceTypeClusterManagedService, ucmsvc.GetName(), PublishUpdate)
	return nil
}
