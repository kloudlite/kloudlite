package domain

import (
	"github.com/kloudlite/api/apps/infra/internal/entities"
	"github.com/kloudlite/api/common"
	"github.com/kloudlite/api/common/fields"
	"github.com/kloudlite/api/pkg/errors"
	"github.com/kloudlite/api/pkg/repos"
	"github.com/kloudlite/operator/operators/resource-watcher/types"
)

func (d *domain) ListPVCs(ctx InfraContext, clusterName string, matchFilters map[string]repos.MatchFilter, pagination repos.CursorPagination) (*repos.PaginatedRecord[*entities.PersistentVolumeClaim], error) {
	filter := repos.Filter{
		fields.AccountName: ctx.AccountName,
		fields.ClusterName: clusterName,
	}
	return d.pvcRepo.FindPaginated(ctx, d.nodePoolRepo.MergeMatchFilters(filter, matchFilters), pagination)
}

func (d *domain) GetPVC(ctx InfraContext, clusterName string, buildRunName string) (*entities.PersistentVolumeClaim, error) {
	pvc, err := d.pvcRepo.FindOne(ctx, repos.Filter{
		fields.AccountName:  ctx.AccountName,
		fields.ClusterName:  clusterName,
		fields.MetadataName: buildRunName,
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
		fields.AccountName:  ctx.AccountName,
		fields.ClusterName:  clusterName,
		fields.MetadataName: pvc.Name,
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

	upvc, err := d.pvcRepo.PatchById(
		ctx,
		xpvc.Id,
		common.PatchForSyncFromAgent(&pvc, pvc.RecordVersion, status, common.PatchOpts{
			MessageTimestamp: opts.MessageTimestamp,
		}))
	if err != nil {
		return errors.NewE(err)
	}
	d.resourceEventPublisher.PublishResourceEvent(ctx, clusterName, ResourceTypePVC, upvc.Name, PublishUpdate)
	return nil
}

func (d *domain) OnPVCDeleteMessage(ctx InfraContext, clusterName string, pvc entities.PersistentVolumeClaim) error {
	if err := d.pvcRepo.DeleteOne(ctx, repos.Filter{
		fields.MetadataName:      pvc.Name,
		fields.MetadataNamespace: pvc.Namespace,
		fields.AccountName:       ctx.AccountName,
		fields.ClusterName:       clusterName,
	}); err != nil {
		return errors.NewE(err)
	}
	d.resourceEventPublisher.PublishResourceEvent(ctx, clusterName, ResourceTypePVC, pvc.Name, PublishDelete)
	return nil
}
