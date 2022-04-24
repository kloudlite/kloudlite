package app

import (
	"kloudlite.io/pkg/messaging"
)

type workloadMessengerImpl struct {
	env      *Env
	producer messaging.Producer[messaging.Json]
}

func (i *workloadMessengerImpl) SendAction(action string, resId string, res any) error {
	err := i.producer.SendMessage(i.env.KafkaWorkloadTopic, resId, messaging.Json{
		"action":  action,
		"payload": res,
	})
	return err
}
