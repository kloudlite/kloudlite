package app

import (
	"context"
	"fmt"
	"kloudlite.io/apps/finance/internal/domain/entities"
	"kloudlite.io/pkg/messaging"
)

type infraMessengerImpl struct {
	env                     *Env
	producer                messaging.Producer[messaging.Json]
	onAddClusterResponse    func(ctx context.Context, m entities.SetupClusterResponse)
	onDeleteClusterResponse func(ctx context.Context, m entities.DeleteClusterResponse)
	onUpdateClusterResponse func(ctx context.Context, m entities.UpdateClusterResponse)
	onAddDeviceResponse     func(ctx context.Context, m entities.AddPeerResponse)
	onRemoveDeviceResponse  func(ctx context.Context, m entities.DeletePeerResponse)
}

func (i *infraMessengerImpl) SendAddClusterAction(action entities.SetupClusterAction) error {
	return i.producer.SendMessage(i.env.KafkaInfraTopic, action.ClusterID, messaging.Json{
		"type":    "setup-cluster",
		"payload": action,
	})
}

func (i *infraMessengerImpl) SendDeleteClusterAction(action entities.DeleteClusterAction) error {
	return i.producer.SendMessage(i.env.KafkaInfraTopic, action.ClusterID, messaging.Json{
		"type":    "delete-cluster",
		"payload": action,
	})
}

func (i *infraMessengerImpl) SendUpdateClusterAction(action entities.UpdateClusterAction) error {
	fmt.Println(i.env, i.producer, action)
	return i.producer.SendMessage(i.env.KafkaInfraTopic, action.ClusterID, messaging.Json{
		"type":    "update-cluster",
		"payload": action,
	})
}

func (i *infraMessengerImpl) SendAddDeviceAction(action entities.AddPeerAction) error {
	return i.producer.SendMessage(i.env.KafkaInfraTopic, action.PublicKey, messaging.Json{
		"type":    "add-peer",
		"payload": action,
	})
}

func (i *infraMessengerImpl) SendRemoveDeviceAction(action entities.DeletePeerAction) error {
	return i.producer.SendMessage(i.env.KafkaInfraTopic, action.PublicKey, messaging.Json{
		"type":    "delete-peer",
		"payload": action,
	})
}
