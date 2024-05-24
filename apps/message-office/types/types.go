package types

import (
	"github.com/shamaton/msgpack/v2"
)

type ErrMessage struct {
	AccountName string
	ClusterName string

	// this must be unmarshalled into github.com/kloudlite/api/apps/tenant-agent/types.AgentErrMessage
	Error []byte
}

func MarshalErrMessage(ru ErrMessage) ([]byte, error) {
	return msgpack.Marshal(ru)
}

func UnmarshalErrMessage(b []byte) (ErrMessage, error) {
	var errM ErrMessage
	if err := msgpack.Unmarshal(b, &errM); err != nil {
		return ErrMessage{}, err
	}
	return errM, nil
}

type ResourceUpdate struct {
	AccountName string
	ClusterName string

	// this should be json unmarshalled into github.com/kloudlite/operator/operators/resource-watcher/types.ResourceUpdate
	WatcherUpdate []byte
}

func MarshalResourceUpdate(ru ResourceUpdate) ([]byte, error) {
	return msgpack.Marshal(ru)
}

func UnmarshalResourceUpdate(b []byte) (ResourceUpdate, error) {
	var ru ResourceUpdate
	if err := msgpack.Unmarshal(b, &ru); err != nil {
		return ResourceUpdate{}, err
	}
	return ru, nil
}
