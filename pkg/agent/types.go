package agent

import (
	"context"

	"kloudlite.io/pkg/redpanda"
)

type Action string

const (
	Apply  Action = "apply"
	Delete Action = "delete"
)

type Message struct {
	Action Action
	Yamls  []byte
}

type Sender interface {
	Dispatch(ctx context.Context, topic string, key string, msg Message) (*redpanda.ProducerOutput, error)
}
