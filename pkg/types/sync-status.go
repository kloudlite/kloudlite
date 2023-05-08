package types

import (
	"time"
)

type SyncState string
type SyncAction string

type SyncStatus struct {
	SyncScheduledAt time.Time  `json:"syncScheduledAt,omitempty"`
	LastSyncedAt    time.Time  `json:"lastSyncedAt,omitempty"`
	Action          SyncAction `json:"action,omitempty"`
	Generation      int64      `json:"generation,omitempty"`
	State           SyncState  `json:"state,omitempty"`
	Error           string     `json:"error,omitempty"`
}

// func (obj *SyncStatus) UnmarshalGQL(v interface{}) error {
//   switch t := v.(type) {
//     case map[string]any:
//       b, err := json.Marshal(t)
//       if err != nil {
//         return err
//       }
//
//       if err := json.Unmarshal(b, obj); err != nil {
//         return err
//       }
//
//     case string:
//       if err := json.Unmarshal([]byte(t), obj); err != nil {
//         return err
//       }
//   }
//
// 	return nil
// }
//
// func (obj SyncStatus) MarshalGQL(w io.Writer) {
// 	b, err := json.Marshal(obj)
// 	if err != nil {
// 		w.Write([]byte("{}"))
// 	}
// 	w.Write(b)
// }

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
