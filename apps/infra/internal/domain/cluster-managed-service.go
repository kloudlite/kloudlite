package domain

import (
	"time"

	iamT "github.com/kloudlite/api/apps/iam/types"
	"github.com/kloudlite/api/apps/infra/internal/entities"
	"github.com/kloudlite/api/common"
	"github.com/kloudlite/api/pkg/errors"
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

func (d *domain) applyClusterManagedService(ctx InfraContext, cmsvc *entities.ClusterManagedService) error {
	addTrackingId(&cmsvc.ClusterManagedService, cmsvc.Id)
	return d.resDispatcher.ApplyToTargetCluster(ctx, cmsvc.ClusterName, &cmsvc.ClusterManagedService, cmsvc.RecordVersion)
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

	cms, err := d.clusterManagedServiceRepo.Create(ctx, &service)
	if err != nil {
		return nil, errors.NewE(err)
	}

	if err := d.applyClusterManagedService(ctx, &service); err != nil {
		return nil, errors.NewE(err)
	}

	d.resourceEventPublisher.PublishCMSEvent(&service, PublishAdd)

	return cms, nil
}

func (d *domain) UpdateClusterManagedService(ctx InfraContext, clusterName string, serviceIn entities.ClusterManagedService) (*entities.ClusterManagedService, error) {
	if err := d.canPerformActionInAccount(ctx, iamT.UpdateClusterManagedService); err != nil {
		return nil, errors.NewE(err)
	}

	serviceIn.EnsureGVK()
	if err := d.k8sClient.ValidateObject(ctx, &serviceIn); err != nil {
		return nil, errors.NewE(err)
	}

	cms, err := d.findClusterManagedService(ctx, clusterName, serviceIn.Name)
	if err != nil {
		return nil, errors.NewE(err)
	}

	if cms.IsMarkedForDeletion() {
		return nil, errors.Newf("cluster managed serviceIn %q (clusterName=%q) is marked for deletion", serviceIn.Name, clusterName)
	}

	unp, err := d.clusterManagedServiceRepo.PatchById(ctx, cms.Id, repos.Document{
		"metadata.labels":      serviceIn.Labels,
		"metadata.annotations": serviceIn.Annotations,
		"displayName":          serviceIn.DisplayName,
		"recordVersion":        cms.RecordVersion + 1,
		"spec":                 serviceIn.Spec,
		"lastUpdatedBy": common.CreatedOrUpdatedBy{
			UserId:    ctx.UserId,
			UserName:  ctx.UserName,
			UserEmail: ctx.UserEmail,
		},
		"syncStatus.lastSyncedAt": time.Now(),
		"syncStatus.action":       t.SyncActionApply,
		"syncStatus.state":        t.SyncStateInQueue,
	})
	if err != nil {
		return nil, errors.NewE(err)
	}

	d.resourceEventPublisher.PublishCMSEvent(unp, PublishUpdate)

	if err := d.applyClusterManagedService(ctx, unp); err != nil {
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
		return errors.Newf("cluster managed service with name %q is marked for deletion", name)
	}

	upC, err := d.clusterManagedServiceRepo.PatchById(ctx, svc.Id, repos.Document{
		"markedForDeletion": true,
		"lastUpdatedBy": common.CreatedOrUpdatedBy{
			UserId:    ctx.UserId,
			UserName:  ctx.UserName,
			UserEmail: ctx.UserEmail,
		},
		"syncStatus.lastSyncedAt": time.Now(),
		"syncStatus.action":       t.SyncActionDelete,
		"syncStatus.state":        t.SyncStateInQueue,
	})
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

	_, err = d.clusterManagedServiceRepo.PatchById(ctx, svc.Id, repos.Document{
		"syncStatus.state":        t.SyncStateErroredAtAgent,
		"syncStatus.lastSyncedAt": opts.MessageTimestamp,
		"syncStatus.error":        &errMsg,
	})
	d.resourceEventPublisher.PublishCMSEvent(svc, PublishUpdate)
	return errors.NewE(err)
}

func (d *domain) OnClusterManagedServiceDeleteMessage(ctx InfraContext, clusterName string, service entities.ClusterManagedService) error {
	xService, err := d.findClusterManagedService(ctx, clusterName, service.Name)
	if err != nil {
		return err
	}
	if xService == nil {
		// does not exist, (maybe already deleted)
		return nil
	}

	if err := d.clusterManagedServiceRepo.DeleteById(ctx, svc.Id); err != nil {
		return err
	}
	d.resourceEventPublisher.PublishCMSEvent(xService, PublishDelete)
	return err
}

func (d *domain) OnClusterManagedServiceUpdateMessage(ctx InfraContext, clusterName string, service entities.ClusterManagedService, status types.ResourceStatus, opts UpdateAndDeleteOpts) error {
	xService, err := d.findClusterManagedService(ctx, clusterName, service.Name)
	if err != nil {
		return errors.NewE(err)
	}

	if err := d.matchRecordVersion(service.Annotations, xService.RecordVersion); err != nil {
		return d.resyncToTargetCluster(ctx, xService.SyncStatus.Action, clusterName, xService, xService.RecordVersion)
	}

	// Ignore error if annotation don't have record version
	annVersion, _ := d.parseRecordVersionFromAnnotations(service.Annotations)

	if _, err := d.clusterManagedServiceRepo.PatchById(ctx, xService.Id, repos.Document{
		"metadata.labels":            service.Labels,
		"metadata.annotations":       service.Annotations,
		"metadata.generation":        service.Generation,
		"metadata.creationTimestamp": service.CreationTimestamp,
		"status":                     service.Status,
		"syncStatus": t.SyncStatus{
			LastSyncedAt:  opts.MessageTimestamp,
			Error:         nil,
			Action:        t.SyncActionApply,
			RecordVersion: annVersion,
			State: func() t.SyncState {
				if status == types.ResourceStatusDeleting {
					return t.SyncStateDeletingAtAgent
				}
				return t.SyncStateUpdatedAtAgent
			}(),
		},
	}); err != nil {
		return errors.NewE(err)
	}
	d.resourceEventPublisher.PublishCMSEvent(xService, PublishUpdate)
	return nil
}
