package common

import (
	"maps"
	"time"

	"github.com/kloudlite/api/common/fields"
	"github.com/kloudlite/api/pkg/repos"
	t "github.com/kloudlite/api/pkg/types"
	"github.com/kloudlite/operator/operators/resource-watcher/types"
	rApi "github.com/kloudlite/operator/pkg/operator"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type PatchOpts struct {
	MessageTimestamp time.Time
	XPatch           repos.Document
}

type ResourceForSync interface {
	GetName() string
	GetNamespace() string
	GetCreationTimestamp() metav1.Time
	GetLabels() map[string]string
	GetDisplayName() string
	GetAnnotations() map[string]string
	GetGeneration() int64
	GetStatus() rApi.Status
	GetRecordVersion() int
}

type ResourceUpdateContext interface {
	GetUserId() repos.ID
	GetUserEmail() string
	GetUserName() string
}

func PatchForSyncFromAgent(
	res ResourceForSync,
	recordVersion int,
	status types.ResourceStatus,
	opts PatchOpts,
) repos.Document {
	res.GetCreationTimestamp()
	generatedPatch := repos.Document{
		fields.MetadataCreationTimestamp: res.GetCreationTimestamp(),
		fields.MetadataLabels:            res.GetLabels(),
		fields.MetadataAnnotations:       res.GetAnnotations(),
		fields.MetadataGeneration:        res.GetGeneration(),
		fields.Status:                    res.GetStatus(),
		fields.SyncStatusState: func() t.SyncState {
			if status == types.ResourceStatusDeleting {
				return t.SyncStateDeletingAtAgent
			}
			return t.SyncStateUpdatedAtAgent
		}(),
		fields.SyncStatusRecordVersion: recordVersion,
		fields.SyncStatusLastSyncedAt:  opts.MessageTimestamp,
		fields.SyncStatusError:         nil,
	}
	var patch repos.Document = nil
	patch = opts.XPatch
	if patch == nil {
		return generatedPatch
	}
	maps.Copy(patch, generatedPatch)
	return patch
}

func PatchForErrorFromAgent(errMsg string, opts PatchOpts) repos.Document {
	return repos.Document{
		fields.SyncStatusState:        t.SyncStateErroredAtAgent,
		fields.SyncStatusLastSyncedAt: opts.MessageTimestamp,
		fields.SyncStatusError:        errMsg,
	}
}

func PatchForMarkDeletion(opts ...PatchOpts) repos.Document {
	generatedPatch := repos.Document{
		fields.MarkedForDeletion:         true,
		fields.SyncStatusSyncScheduledAt: time.Now(),
		fields.SyncStatusAction:          t.SyncActionDelete,
		fields.SyncStatusState:           t.SyncStateIdle,
	}
	var patch repos.Document = nil
	if len(opts) > 0 {
		patch = opts[0].XPatch
	}
	if patch == nil {
		return generatedPatch
	}
	maps.Copy(patch, generatedPatch)
	return patch
}

func PatchForUpdate(
	ctx ResourceUpdateContext,
	res ResourceForSync,
	opts ...PatchOpts,
) repos.Document {
	generatedPatch := repos.Document{
		fields.MetadataLabels:      res.GetLabels(),
		fields.MetadataAnnotations: res.GetAnnotations(),
		fields.DisplayName:         res.GetDisplayName(),
		fields.LastUpdatedBy: CreatedOrUpdatedBy{
			UserId:    ctx.GetUserId(),
			UserName:  ctx.GetUserName(),
			UserEmail: ctx.GetUserEmail(),
		},
		fields.SyncStatusSyncScheduledAt: time.Now(),
		fields.SyncStatusState:           t.SyncStateInQueue,
		fields.SyncStatusAction:          t.SyncActionApply,
		"$inc": repos.Document{
			fields.RecordVersion: 1,
		},
	}
	var patch repos.Document
	if len(opts) > 0 {
		patch = opts[0].XPatch
	}
	if patch == nil {
		return generatedPatch
	}
	maps.Copy(patch, generatedPatch)
	return patch
}
