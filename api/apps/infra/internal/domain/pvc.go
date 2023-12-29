package domain

import (
	"github.com/kloudlite/api/apps/infra/internal/entities"
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
	pvc.SyncStatus = t.SyncStatus{
		LastSyncedAt: opts.MessageTimestamp,
		State: func() t.SyncState {
			if status == types.ResourceStatusDeleting {
				return t.SyncStateDeletingAtAgent
			}
			return t.SyncStateUpdatedAtAgent
		}(),
	}

	if _, err := d.pvcRepo.Upsert(ctx, repos.Filter{
		"metadata.name":      pvc.Name,
		"metadata.namespace": pvc.Namespace,
		"accountName":        ctx.AccountName,
		"clusterName":        clusterName,
	}, &pvc); err != nil {
		return errors.NewE(err)
	}
	d.resourceEventPublisher.PublishPvcResEvent(&pvc, PublishUpdate)
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
