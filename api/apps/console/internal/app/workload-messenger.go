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
	fmt.Println(res)
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
	i.producer.Produce(context.TODO(), i.topic, resId, marshal)
	return err
}

func fxWorkloadMessenger(env *WorkloadConsumerEnv, p redpanda.Producer) domain.WorkloadMessenger {
	return &workloadMessengerImpl{
		topic:    env.Topic,
		producer: p,
	}
}
