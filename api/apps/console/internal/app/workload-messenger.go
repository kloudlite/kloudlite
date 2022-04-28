package app

import (
	"kloudlite.io/apps/console/internal/domain"
	"kloudlite.io/pkg/messaging"
)

type workloadMessengerImpl struct {
	topic    string
	producer messaging.Producer[messaging.Json]
}

func (i *workloadMessengerImpl) SendAction(action string, resId string, res any) error {
	err := i.producer.SendMessage(i.topic, resId, messaging.Json{
		"action":  action,
		"payload": res,
	})
	return err
}

func fxWorkloadMessenger(env *WorkloadConsumerEnv, p messaging.Producer[messaging.Json]) domain.WorkloadMessenger {
	return &workloadMessengerImpl{
		topic:    env.ResponseTopic,
		producer: p,
	}
}
