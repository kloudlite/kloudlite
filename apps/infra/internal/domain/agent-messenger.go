package domain

import (
	"context"
	"encoding/json"
	"kloudlite.io/pkg/redpanda"
)

type Action string

const (
	ActionApply  Action = "apply"
	ActionDelete Action = "delete"
)

type AgentMessenger interface {
	SendAction(ctx context.Context, action Action, topic string, key string, res any) error
}

type aMessenger struct {
	producer redpanda.Producer
}

func (am aMessenger) SendAction(ctx context.Context, action Action, topic string, key string, res any) error {
	marshal, err := json.Marshal(
		map[string]any{
			"action":  action,
			"payload": res,
		},
	)
	if err != nil {
		return err
	}
	if _, err := am.producer.Produce(ctx, topic, key, marshal); err != nil {
		return err
	}
	return nil
}

func NewAgentMessenger(producer redpanda.Producer) AgentMessenger {
	return &aMessenger{producer: producer}
}
