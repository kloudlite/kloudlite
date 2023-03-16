package agent

import (
	"context"
	"encoding/json"

	"kloudlite.io/pkg/redpanda"
)

type sender struct {
	producer redpanda.Producer
}

func (s *sender) Dispatch(ctx context.Context, topic string, key string, msg Message) (*redpanda.ProducerOutput, error) {
	m, err := json.Marshal(msg)
	if err != nil {
		return nil, err
	}
	return s.producer.Produce(ctx, topic, key, m)
}

func NewSender(producer redpanda.Producer) Sender {
	return &sender{
		producer: producer,
	}
}
