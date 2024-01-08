package domain

import (
	"github.com/kloudlite/api/apps/infra/internal/entities"
	"github.com/kloudlite/api/common"
	"github.com/kloudlite/api/pkg/errors"
	"github.com/kloudlite/api/pkg/repos"
	t "github.com/kloudlite/api/pkg/types"
	"github.com/kloudlite/operator/operators/resource-watcher/types"
)

// GetVolumeAttachment implements Domain.
func (d *domain) GetVolumeAttachment(ctx InfraContext, clusterName string, volAttachmentName string) (*entities.VolumeAttachment, error) {
	volatt, err := d.volumeAttachmentRepo.FindOne(ctx, repos.Filter{
		"accountName":   ctx.AccountName,
		"clusterName":   clusterName,
		"metadata.name": volAttachmentName,
	})
	if err != nil {
		return nil, errors.NewE(err)
	}

	if volatt == nil {
		return nil, errors.Newf("persistent volume claim with name %q not found", volAttachmentName)
	}
	return volatt, nil
}

// ListVolumeAttachments implements Domain.
func (d *domain) ListVolumeAttachments(ctx InfraContext, clusterName string, search map[string]repos.MatchFilter, pagination repos.CursorPagination) (*repos.PaginatedRecord[*entities.VolumeAttachment], error) {
	filter := repos.Filter{
		"accountName": ctx.AccountName,
		"clusterName": clusterName,
	}
	return d.volumeAttachmentRepo.FindPaginated(ctx, d.nodePoolRepo.MergeMatchFilters(filter, search), pagination)
}

// OnVolumeAttachmentDeleteMessage implements Domain.
func (d *domain) OnVolumeAttachmentDeleteMessage(ctx InfraContext, clusterName string, volumeAttachment entities.VolumeAttachment) error {
	if err := d.pvcRepo.DeleteOne(ctx, repos.Filter{
		"metadata.name":      volumeAttachment.Name,
		"metadata.namespace": volumeAttachment.Namespace,
		"accountName":        ctx.AccountName,
		"clusterName":        clusterName,
	}); err != nil {
		return errors.NewE(err)
	}
	d.resourceEventPublisher.PublishVolumeAttachmentEvent(&volumeAttachment, PublishDelete)
	return nil
}

// OnVolumeAttachmentUpdateMessage implements Domain.
func (d *domain) OnVolumeAttachmentUpdateMessage(ctx InfraContext, clusterName string, volumeAttachment entities.VolumeAttachment, status types.ResourceStatus, opts UpdateAndDeleteOpts) error {
	vatt, err := d.volumeAttachmentRepo.FindOne(ctx, repos.Filter{
		"accountName":   ctx.AccountName,
		"clusterName":   clusterName,
		"metadata.name": volumeAttachment.Name,
	})
	if err != nil {
		return err
	}

	if vatt == nil {
		volumeAttachment.AccountName = ctx.AccountName
		volumeAttachment.ClusterName = clusterName

		volumeAttachment.CreatedBy = common.CreatedOrUpdatedBy{
			UserId:    repos.ID(common.CreatedByResourceSyncUserId),
			UserName:  common.CreatedByResourceSyncUsername,
			UserEmail: common.CreatedByResourceSyncUserEmail,
		}
		volumeAttachment.LastUpdatedBy = volumeAttachment.CreatedBy
		vatt, err = d.volumeAttachmentRepo.Create(ctx, &volumeAttachment)
		if err != nil {
			return errors.NewE(err)
		}
	}

	vatt.SyncStatus.LastSyncedAt = opts.MessageTimestamp
	vatt.SyncStatus.State = func() t.SyncState {
		if status == types.ResourceStatusDeleting {
			return t.SyncStateDeletingAtAgent
		}
		return t.SyncStateUpdatedAtAgent
	}()

	vatt, err = d.volumeAttachmentRepo.UpdateById(ctx, vatt.Id, vatt)
	if err != nil {
		return errors.NewE(err)
	}
	d.resourceEventPublisher.PublishVolumeAttachmentEvent(vatt, PublishUpdate)
	return nil
}
