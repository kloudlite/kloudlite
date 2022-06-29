package app

import (
	"context"
	"encoding/json"
	"kloudlite.io/apps/console/internal/domain"
	"kloudlite.io/pkg/messaging"
	"kloudlite.io/pkg/redpanda"
)

type workloadMessengerImpl struct {
	topic    string
	producer redpanda.Producer
}

func (i *workloadMessengerImpl) SendAction(action string, resId string, res any) error {
	marshal, err := json.Marshal(messaging.Json{
		"action":  action,
		"payload": res,
	})
	if err != nil {
		return err
	}
	i.producer.Produce(context.TODO(), i.topic, resId, marshal)
	return err
}

func fxWorkloadMessenger(env *WorkloadConsumerEnv, p redpanda.Producer) domain.WorkloadMessenger {
	return &workloadMessengerImpl{
		topic:    env.Topic,
		producer: p,
	}
}
