package domain

import (
	"github.com/kloudlite/api/apps/infra/internal/entities"
	"github.com/kloudlite/api/common"
	"github.com/kloudlite/api/pkg/errors"
	"github.com/kloudlite/api/pkg/repos"
	t "github.com/kloudlite/api/pkg/types"
	"github.com/kloudlite/operator/operators/resource-watcher/types"
)

func (d *domain) ListPVCs(ctx InfraContext, clusterName string, matchFilters map[string]repos.MatchFilter, pagination repos.CursorPagination) (*repos.PaginatedRecord[*entities.PersistentVolumeClaim], error) {
	filter := repos.Filter{
		"accountName": ctx.AccountName,
		"clusterName": clusterName,
	}
	return d.pvcRepo.FindPaginated(ctx, d.nodePoolRepo.MergeMatchFilters(filter, matchFilters), pagination)
}

func (d *domain) GetPVC(ctx InfraContext, clusterName string, buildRunName string) (*entities.PersistentVolumeClaim, error) {
	pvc, err := d.pvcRepo.FindOne(ctx, repos.Filter{
		"accountName":   ctx.AccountName,
		"clusterName":   clusterName,
		"metadata.name": buildRunName,
	})
	if err != nil {
		return nil, errors.NewE(err)
	}

	if pvc == nil {
		return nil, errors.Newf("persistent volume claim with name %q not found", buildRunName)
	}
	return pvc, nil
}

func (d *domain) OnPVCUpdateMessage(ctx InfraContext, clusterName string, pvc entities.PersistentVolumeClaim, status types.ResourceStatus, opts UpdateAndDeleteOpts) error {
	xpvc, err := d.pvcRepo.FindOne(ctx, repos.Filter{
		"accountName":   ctx.AccountName,
		"clusterName":   clusterName,
		"metadata.name": pvc.Name,
	})
	if err != nil {
		return err
	}

	if xpvc == nil {
		pvc.AccountName = ctx.AccountName
		pvc.ClusterName = clusterName

		pvc.CreatedBy = common.CreatedOrUpdatedBy{
			UserId:    repos.ID(common.CreatedByResourceSyncUserId),
			UserName:  common.CreatedByResourceSyncUsername,
			UserEmail: common.CreatedByResourceSyncUserEmail,
		}
		pvc.LastUpdatedBy = pvc.CreatedBy
		xpvc, err = d.pvcRepo.Create(ctx, &pvc)
		if err != nil {
			return errors.NewE(err)
		}
	}

	xpvc.SyncStatus.LastSyncedAt = opts.MessageTimestamp
	xpvc.SyncStatus.State = func() t.SyncState {
		if status == types.ResourceStatusDeleting {
			return t.SyncStateDeletingAtAgent
		}
		return t.SyncStateUpdatedAtAgent
	}()

	upvc, err := d.pvcRepo.UpdateById(ctx, xpvc.Id, xpvc)
	if err != nil {
		return errors.NewE(err)
	}
	d.resourceEventPublisher.PublishPvcResEvent(upvc, PublishUpdate)
	return nil
}

func (d *domain) OnPVCDeleteMessage(ctx InfraContext, clusterName string, pvc entities.PersistentVolumeClaim) error {
	if err := d.pvcRepo.DeleteOne(ctx, repos.Filter{
		"metadata.name":      pvc.Name,
		"metadata.namespace": pvc.Namespace,
		"accountName":        ctx.AccountName,
		"clusterName":        clusterName,
	}); err != nil {
		return errors.NewE(err)
	}
	d.resourceEventPublisher.PublishPvcResEvent(&pvc, PublishDelete)
	return nil
}
