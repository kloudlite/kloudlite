package domain

import (
	"github.com/kloudlite/api/apps/infra/internal/entities"
	"github.com/kloudlite/api/common"
	"github.com/kloudlite/api/pkg/errors"
	"github.com/kloudlite/api/pkg/repos"
	t "github.com/kloudlite/api/pkg/types"
	"github.com/kloudlite/operator/operators/resource-watcher/types"
)

// GetPV implements Domain.
func (d *domain) GetPV(ctx InfraContext, clusterName string, pvName string) (*entities.PersistentVolume, error) {
	pv, err := d.pvRepo.FindOne(ctx, repos.Filter{
		"accountName":   ctx.AccountName,
		"clusterName":   clusterName,
		"metadata.name": pvName,
	})
	if err != nil {
		return nil, errors.NewE(err)
	}

	if pv == nil {
		return nil, errors.Newf("persistent volume with name %q not found", pvName)
	}
	return pv, nil
}

// ListPVs implements Domain.
func (d *domain) ListPVs(ctx InfraContext, clusterName string, search map[string]repos.MatchFilter, pagination repos.CursorPagination) (*repos.PaginatedRecord[*entities.PersistentVolume], error) {
	filter := repos.Filter{
		"accountName": ctx.AccountName,
		"clusterName": clusterName,
	}
	return d.pvRepo.FindPaginated(ctx, d.nodePoolRepo.MergeMatchFilters(filter, search), pagination)
}

// OnPVDeleteMessage implements Domain.
func (d *domain) OnPVDeleteMessage(ctx InfraContext, clusterName string, pv entities.PersistentVolume) error {
	if err := d.pvcRepo.DeleteOne(ctx, repos.Filter{
		"metadata.name":      pv.Name,
		"metadata.namespace": pv.Namespace,
		"accountName":        ctx.AccountName,
		"clusterName":        clusterName,
	}); err != nil {
		return errors.NewE(err)
	}
	d.resourceEventPublisher.PublishPvResEvent(&pv, PublishDelete)
	return nil
}

// OnPVUpdateMessage implements Domain.
func (d *domain) OnPVUpdateMessage(ctx InfraContext, clusterName string, pv entities.PersistentVolume, status types.ResourceStatus, opts UpdateAndDeleteOpts) error {
	pvol, err := d.pvRepo.FindOne(ctx, repos.Filter{
		"accountName":   ctx.AccountName,
		"clusterName":   clusterName,
		"metadata.name": pv.Name,
	})
	if err != nil {
		return err
	}

	if pvol == nil {
		pv.CreatedBy = common.CreatedOrUpdatedBy{
			UserId:    repos.ID(common.CreatedByResourceSyncUserId),
			UserName:  common.CreatedByResourceSyncUsername,
			UserEmail: common.CreatedByResourceSyncUserEmail,
		}
		pv.LastUpdatedBy = pv.CreatedBy
		pvol, err = d.pvRepo.Create(ctx, &pv)
		if err != nil {
			return errors.NewE(err)
		}
	}

	pvol.SyncStatus.LastSyncedAt = opts.MessageTimestamp
	pvol.SyncStatus.State = func() t.SyncState {
		if status == types.ResourceStatusDeleting {
			return t.SyncStateDeletingAtAgent
		}
		return t.SyncStateUpdatedAtAgent
	}()

	pvol, err = d.pvRepo.UpdateById(ctx, pvol.Id, pvol)
	if err != nil {
		return errors.NewE(err)
	}
	d.resourceEventPublisher.PublishPvResEvent(pvol, PublishUpdate)
	return nil
}
