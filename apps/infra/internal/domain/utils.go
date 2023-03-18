package domain

import (
	"time"

	t "kloudlite.io/pkg/types"
)

func getSyncStatusForCreation() t.SyncStatus {
	return t.SyncStatus{
		SyncScheduledAt: time.Now(),
		Action:          t.SyncActionApply,
		Generation:      1,
		State:           t.SyncStateIdle,
	}
}

func getSyncStatusForUpdation(generation int64) t.SyncStatus {
	return t.SyncStatus{
		SyncScheduledAt: time.Now(),
		Action:          t.SyncActionApply,
		Generation:      generation,
		State:           t.SyncStateIdle,
	}
}

func getSyncStatusForDeletion(generation int64) t.SyncStatus {
	return t.SyncStatus{
		SyncScheduledAt: time.Now(),
		Action:          t.SyncActionDelete,
		Generation:      generation,
		State:           t.SyncStateIdle,
	}
}

func parseSyncState(isReady bool) t.SyncState {
	if isReady {
		return t.SyncStateReady
	}
	return t.SyncStateNotReady
}
