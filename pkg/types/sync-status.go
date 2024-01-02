package types

import (
	"time"
)

type (
	SyncState  string
	SyncAction string
)

type SyncStatus struct {
	SyncScheduledAt time.Time  `json:"syncScheduledAt,omitempty"`
	LastSyncedAt    time.Time  `json:"lastSyncedAt,omitempty"`
  Action          SyncAction `json:"action"`
	RecordVersion   int        `json:"recordVersion"`
	State           SyncState  `json:"state"`
	Error           *string    `json:"error,omitempty"`
}

const (
	SyncActionApply  SyncAction = "APPLY"
	SyncActionDelete SyncAction = "DELETE"
)

const (
	SyncStateIdle                    SyncState = "IDLE"
	SyncStateInQueue                 SyncState = "IN_QUEUE"
	SyncStateAppliedAtAgent          SyncState = "APPLIED_AT_AGENT"
	SyncStateErroredAtAgent          SyncState = "ERRORED_AT_AGENT"
	SyncStateReceivedUpdateFromAgent SyncState = "RECEIVED_UPDATE_FROM_AGENT"
	SyncStateUpdatedAtAgent          SyncState = "UPDATED_AT_AGENT"
	SyncStateDeletingAtAgent         SyncState = "DELETING_AT_AGENT"
	SyncStateDeletedAtAgent          SyncState = "DELETED_AT_AGENT"
)

func GenSyncStatus(action SyncAction, recordVersion int) SyncStatus {
	return SyncStatus{
		SyncScheduledAt: time.Now(),
		Action:          action,
		RecordVersion:   recordVersion,
		State:           SyncStateIdle,
	}
}

func GetSyncStatusForCreation() SyncStatus {
	return SyncStatus{
		SyncScheduledAt: time.Now(),
		Action:          SyncActionApply,
		RecordVersion:   1,
		State:           SyncStateInQueue,
	}
}

func GetSyncStatusForDeletion(generation int64) SyncStatus {
	return SyncStatus{
		SyncScheduledAt: time.Now(),
		Action:          SyncActionDelete,
		RecordVersion:   int(generation),
		State:           SyncStateInQueue,
	}
}
