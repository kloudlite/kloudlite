package domain

import (
	"github.com/kloudlite/api/apps/infra/internal/entities"
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
	if err := d.pvRepo.DeleteOne(ctx, repos.Filter{
		"metadata.name":      pv.Name,
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
	pv.SyncStatus.LastSyncedAt = opts.MessageTimestamp
	pv.SyncStatus.State = func() t.SyncState {
		if status == types.ResourceStatusDeleting {
			return t.SyncStateDeletingAtAgent
		}
		return t.SyncStateUpdatedAtAgent
	}()
	pv.AccountName = ctx.AccountName
	pv.ClusterName = clusterName
	upsert, err := d.pvRepo.Upsert(ctx, repos.Filter{
		"accountName":   ctx.AccountName,
		"clusterName":   clusterName,
		"metadata.name": pv.Name,
	}, &pv)
	if err != nil {
		return errors.NewE(err)
	}
	d.resourceEventPublisher.PublishPvResEvent(upsert, PublishUpdate)
	return nil
}
