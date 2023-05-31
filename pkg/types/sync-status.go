package types

import (
	"time"
)

type SyncState string
type SyncAction string

type SyncStatus struct {
	SyncScheduledAt time.Time  `json:"syncScheduledAt,omitempty"`
	LastSyncedAt    time.Time  `json:"lastSyncedAt,omitempty"`
	Action          SyncAction `json:"action"`
	Generation      int64      `json:"generation"`
	State           SyncState  `json:"state,omitempty"`
	Error           *string    `json:"error,omitempty"`
}

const (
	SyncActionApply  SyncAction = "APPLY"
	SyncActionDelete SyncAction = "DELETE"
)

const (
	SyncStateIdle       SyncState = "IDLE"
	SyncStateInProgress SyncState = "IN_PROGRESS"
	SyncStateReady      SyncState = "READY"
	SyncStateNotReady   SyncState = "NOT_READY"
)

func GenSyncStatus(action SyncAction, generation int64) SyncStatus {
	return SyncStatus{
		SyncScheduledAt: time.Now(),
		Action:          action,
		Generation:      generation,
		State:           SyncStateIdle,
	}
}

func GetSyncStatusForCreation() SyncStatus {
	return SyncStatus{
		SyncScheduledAt: time.Now(),
		Action:          SyncActionApply,
		Generation:      1,
		State:           SyncStateIdle,
	}
}

func GetSyncStatusForUpdation(generation int64) SyncStatus {
	return SyncStatus{
		SyncScheduledAt: time.Now(),
		Action:          SyncActionApply,
		Generation:      generation,
		State:           SyncStateIdle,
	}
}

func GetSyncStatusForDeletion(generation int64) SyncStatus {
	return SyncStatus{
		SyncScheduledAt: time.Now(),
		Action:          SyncActionDelete,
		Generation:      generation,
		State:           SyncStateIdle,
	}
}

func ParseSyncState(isReady bool) SyncState {
	if isReady {
		return SyncStateReady
	}
	return SyncStateNotReady
}
