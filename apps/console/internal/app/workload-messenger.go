package app

import (
	"context"
	"encoding/json"
	"fmt"

	"kloudlite.io/apps/console/internal/domain"
	"kloudlite.io/pkg/redpanda"
)

type workloadMessengerImpl struct {
	topic    string
	producer redpanda.Producer
}

func (i *workloadMessengerImpl) SendAction(action string, resId string, res any) error {
	marshal, err := json.Marshal(
		map[string]any{
			"action":  action,
			"payload": res,
		},
	)
	if err != nil {
		return err
	}
	fmt.Println(i.topic, resId, string(marshal))
	if _, err := i.producer.Produce(context.TODO(), i.topic, resId, marshal); err != nil {
		return err
	}
	return nil
}

func fxWorkloadMessenger(env *WorkloadConsumerEnv, p redpanda.Producer) domain.WorkloadMessenger {
	return &workloadMessengerImpl{
		topic:    env.Topic,
		producer: p,
	}
}
