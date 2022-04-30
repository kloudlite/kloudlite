package app

import (
	"kloudlite.io/apps/console/internal/domain"
	"kloudlite.io/apps/console/internal/domain/entities"
	"kloudlite.io/pkg/errors"
	"kloudlite.io/pkg/messaging"
)

type infraMessengerImpl struct {
	topic    string
	producer messaging.Producer[messaging.Json]
}

func (i *infraMessengerImpl) SendAction(action any) error {
	switch a := action.(type) {
	case entities.SetupClusterAction:
		{
			return i.producer.SendMessage(i.topic, string(a.ClusterID), messaging.Json{
				"type":    "create-cluster",
				"payload": action,
			})
		}
	case entities.DeleteClusterAction:
		{
			return i.producer.SendMessage(i.topic, string(a.ClusterID), messaging.Json{
				"type":    "delete-cluster",
				"payload": action,
			})
		}
	case entities.UpdateClusterAction:
		{
			return i.producer.SendMessage(i.topic, string(a.ClusterID), messaging.Json{
				"type":    "update-cluster",
				"payload": action,
			})
		}
	case entities.AddPeerAction:
		{
			return i.producer.SendMessage(i.topic, a.PublicKey, messaging.Json{
				"type":    "add-peer",
				"payload": action,
			})
		}
	case entities.DeletePeerAction:
		{
			return i.producer.SendMessage(i.topic, a.PublicKey, messaging.Json{
				"type":    "delete-peer",
				"payload": action,
			})
		}

	}
	return errors.New("no matching message type")
}

func fxInfraMessenger(env *InfraConsumerEnv, p messaging.Producer[messaging.Json]) domain.InfraMessenger {
	return &infraMessengerImpl{
		topic:    env.Topic,
		producer: p,
	}
}
