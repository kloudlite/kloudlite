package types

import "time"

type SyncState string
type SyncAction string

type SyncStatus struct {
	SyncScheduledAt time.Time  `json:"syncScheduledAt,omitempty"`
	LastSyncedAt    time.Time  `json:"lastSyncedAt,omitempty"`
	Action          SyncAction `json:"action,omitempty"`
	Generation      int64      `json:"generation,omitempty"`
	State           SyncState  `json:"state,omitempty"`
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
