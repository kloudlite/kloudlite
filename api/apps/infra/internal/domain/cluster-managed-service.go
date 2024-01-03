package domain

import (
	iamT "github.com/kloudlite/api/apps/iam/types"
	"github.com/kloudlite/api/apps/infra/internal/entities"
	"github.com/kloudlite/api/common"
	"github.com/kloudlite/api/pkg/errors"
	fn "github.com/kloudlite/api/pkg/functions"
	"github.com/kloudlite/api/pkg/repos"
	t "github.com/kloudlite/api/pkg/types"
	"github.com/kloudlite/operator/operators/resource-watcher/types"
)

func (d *domain) ListClusterManagedServices(ctx InfraContext, clusterName string, mf map[string]repos.MatchFilter, pagination repos.CursorPagination) (*repos.PaginatedRecord[*entities.ClusterManagedService], error) {
	if err := d.canPerformActionInAccount(ctx, iamT.ListClusterManagedServices); err != nil {
		return nil, errors.NewE(err)
	}

	f := repos.Filter{
		"clusterName": clusterName,
		"accountName": ctx.AccountName,
	}

	pr, err := d.clusterManagedServiceRepo.FindPaginated(ctx, d.secretRepo.MergeMatchFilters(f, mf), pagination)
	if err != nil {
		return nil, errors.NewE(err)
	}

	return pr, nil
}

func (d *domain) findClusterManagedService(ctx InfraContext, clusterName string, svcName string) (*entities.ClusterManagedService, error) {
	cmsvc, err := d.clusterManagedServiceRepo.FindOne(ctx, repos.Filter{
		"clusterName":   clusterName,
		"accountName":   ctx.AccountName,
		"metadata.name": svcName,
	})
	if err != nil {
		return nil, errors.NewE(err)
	}

	if cmsvc == nil {
		return nil, errors.Newf("cmsvc with name %q not found", clusterName)
	}
	return cmsvc, nil
}

func (d *domain) GetClusterManagedService(ctx InfraContext, clusterName string, serviceName string) (*entities.ClusterManagedService, error) {
	if err := d.canPerformActionInAccount(ctx, iamT.GetClusterManagedService); err != nil {
		return nil, errors.NewE(err)
	}

	c, err := d.findClusterManagedService(ctx, clusterName, serviceName)
	if err != nil {
		return nil, errors.NewE(err)
	}

	return c, nil
}

func (d *domain) CreateClusterManagedService(ctx InfraContext, clusterName string, service entities.ClusterManagedService) (*entities.ClusterManagedService, error) {
	if err := d.canPerformActionInAccount(ctx, iamT.CreateClusterManagedService); err != nil {
		return nil, errors.NewE(err)
	}

	service.IncrementRecordVersion()

	service.CreatedBy = common.CreatedOrUpdatedBy{
		UserId:    ctx.UserId,
		UserName:  ctx.UserName,
		UserEmail: ctx.UserEmail,
	}

	service.LastUpdatedBy = service.CreatedBy

	existing, err := d.clusterManagedServiceRepo.FindOne(ctx, repos.Filter{
		"clusterName":   clusterName,
		"accountName":   ctx.AccountName,
		"metadata.name": service.Name,
	})
	if err != nil {
		return nil, errors.NewE(err)
	}

	if existing != nil {
		return nil, errors.Newf("cluster managed service with name %q already exists", clusterName)
	}

	service.AccountName = ctx.AccountName
	service.ClusterName = clusterName
	service.SyncStatus = t.GenSyncStatus(t.SyncActionApply, service.RecordVersion)

	service.EnsureGVK()

	if err := d.k8sClient.ValidateObject(ctx, &service.ClusterManagedService); err != nil {
		return nil, errors.NewE(err)
	}

	if err := d.resDispatcher.ApplyToTargetCluster(ctx, clusterName, &service.ClusterManagedService, 1); err != nil {
		return nil, errors.NewE(err)
	}

	if cms, err := d.clusterManagedServiceRepo.Create(ctx, &service); err != nil {
		return nil, errors.NewE(err)
	} else {
		d.resourceEventPublisher.PublishCMSEvent(&service, PublishAdd)

		return cms, nil
	}
}

func (d *domain) UpdateClusterManagedService(ctx InfraContext, clusterName string, service entities.ClusterManagedService) (*entities.ClusterManagedService, error) {
	if err := d.canPerformActionInAccount(ctx, iamT.UpdateClusterManagedService); err != nil {
		return nil, errors.NewE(err)
	}

	service.EnsureGVK()
	if err := d.k8sClient.ValidateObject(ctx, &service); err != nil {
		return nil, errors.NewE(err)
	}

	cms, err := d.findClusterManagedService(ctx, clusterName, service.Name)
	if err != nil {
		return nil, errors.NewE(err)
	}

	if cms.IsMarkedForDeletion() {
		return nil, errors.Newf("cluster managed service %q (clusterName=%q) is marked for deletion", service.Name, clusterName)
	}

	cms.IncrementRecordVersion()
	cms.LastUpdatedBy = common.CreatedOrUpdatedBy{
		UserId:    ctx.UserId,
		UserName:  ctx.UserName,
		UserEmail: ctx.UserEmail,
	}

	cms.Labels = service.Labels
	cms.Annotations = service.Annotations

	cms.SyncStatus = t.GenSyncStatus(t.SyncActionApply, cms.RecordVersion)

	unp, err := d.clusterManagedServiceRepo.UpdateById(ctx, cms.Id, cms)
	if err != nil {
		return nil, errors.NewE(err)
	}

	d.resourceEventPublisher.PublishCMSEvent(unp, PublishUpdate)

	if err := d.resDispatcher.ApplyToTargetCluster(ctx, clusterName, &unp.ClusterManagedService, unp.RecordVersion); err != nil {
		return nil, errors.NewE(err)
	}

	return unp, nil
}

func (d *domain) DeleteClusterManagedService(ctx InfraContext, clusterName string, name string) error {
	if err := d.canPerformActionInAccount(ctx, iamT.DeleteClusterManagedService); err != nil {
		return errors.NewE(err)
	}

	svc, err := d.findClusterManagedService(ctx, clusterName, name)
	if err != nil {
		return errors.NewE(err)
	}

	if svc.IsMarkedForDeletion() {
		return errors.Newf("cluster managed service %q (clusterName=%q) is already marked for deletion", name, clusterName)
	}

	svc.MarkedForDeletion = fn.New(true)
	svc.SyncStatus = t.GetSyncStatusForDeletion(svc.Generation)
	upC, err := d.clusterManagedServiceRepo.UpdateById(ctx, svc.Id, svc)
	if err != nil {
		return errors.NewE(err)
	}

	d.resourceEventPublisher.PublishCMSEvent(upC, PublishUpdate)

	return d.resDispatcher.DeleteFromTargetCluster(ctx, clusterName, &upC.ClusterManagedService)
}

func (d *domain) OnClusterManagedServiceApplyError(ctx InfraContext, clusterName, name, errMsg string, opts UpdateAndDeleteOpts) error {
	svc, err := d.findClusterManagedService(ctx, clusterName, name)
	if err != nil {
		return errors.NewE(err)
	}

	svc.SyncStatus.State = t.SyncStateErroredAtAgent
	svc.SyncStatus.LastSyncedAt = opts.MessageTimestamp
	svc.SyncStatus.Error = &errMsg

	_, err = d.clusterManagedServiceRepo.UpdateById(ctx, svc.Id, svc)
	d.resourceEventPublisher.PublishCMSEvent(svc, PublishUpdate)
	return errors.NewE(err)
}

func (d *domain) OnClusterManagedServiceDeleteMessage(ctx InfraContext, clusterName string, service entities.ClusterManagedService) error {
	svc, err := d.findClusterManagedService(ctx, clusterName, service.Name)
	if err != nil {
		return err
	}

	if svc == nil {
		// does not exist, (maybe already deleted)
		return nil
	}

	if err := d.matchRecordVersion(service.Annotations, svc.RecordVersion); err != nil {
		return d.resyncToTargetCluster(ctx, svc.SyncStatus.Action, clusterName, svc, svc.RecordVersion)
	}

	if err := d.clusterManagedServiceRepo.DeleteById(ctx, svc.Id); err != nil {
		return err
	}
	d.resourceEventPublisher.PublishCMSEvent(svc, PublishDelete)
	return err
}

func (d *domain) OnClusterManagedServiceUpdateMessage(ctx InfraContext, clusterName string, service entities.ClusterManagedService, status types.ResourceStatus, opts UpdateAndDeleteOpts) error {
	svc, err := d.findClusterManagedService(ctx, clusterName, service.Name)
	if err != nil {
		return errors.NewE(err)
	}

	if err := d.matchRecordVersion(service.Annotations, svc.RecordVersion); err != nil {
		return d.resyncToTargetCluster(ctx, svc.SyncStatus.Action, clusterName, svc, svc.RecordVersion)
	}

	svc.Status = service.Status

	svc.SyncStatus.State = func() t.SyncState {
		if status == types.ResourceStatusDeleting {
			return t.SyncStateDeletingAtAgent
		}
		return t.SyncStateUpdatedAtAgent
	}()
	svc.SyncStatus.LastSyncedAt = opts.MessageTimestamp
	svc.SyncStatus.Error = nil
	svc.SyncStatus.RecordVersion = svc.RecordVersion

	if _, err := d.clusterManagedServiceRepo.UpdateById(ctx, svc.Id, svc); err != nil {
		return errors.NewE(err)
	}
	d.resourceEventPublisher.PublishCMSEvent(svc, PublishUpdate)
	return nil
}
