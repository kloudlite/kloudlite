package application

import (
	"context"
	"kloudlite.io/apps/infra/internal/domain"
	"kloudlite.io/pkg/messaging"
	//"kloudlite.io/pkg/messaging"
	//"kloudlite.io/pkg/messaging"
)

type infraResponder struct {
	kProducer     messaging.Producer[messaging.Json]
	responseTopic string
}

func (i *infraResponder) SendSetupAccountResponse(cxt context.Context, action domain.SetupAccountResponse) error {
	return i.kProducer.SendMessage(i.responseTopic, "resp", messaging.Json{
		"type":    "setup-cluster-account",
		"payload": action,
	})
}

// SendAddPeerResponse implements domain.InfraJobResponder
func (i *infraResponder) SendAddPeerResponse(cxt context.Context, action domain.AddPeerResponse) error {
	return i.kProducer.SendMessage(i.responseTopic, "resp", messaging.Json{
		"type":    "add-peer",
		"payload": action,
	})
}

// SendCreateClusterResponse implements domain.InfraJobResponder
func (i *infraResponder) SendCreateClusterResponse(cxt context.Context, action domain.SetupClusterResponse) error {
	return i.kProducer.SendMessage(i.responseTopic, "resp", messaging.Json{
		"type":    "create-cluster",
		"payload": action,
	})
}

// SendDeleteClusterResponse implements domain.InfraJobResponder
func (i *infraResponder) SendDeleteClusterResponse(cxt context.Context, action domain.DeleteClusterResponse) error {
	return i.kProducer.SendMessage(i.responseTopic, "resp", messaging.Json{
		"type":    "delete-cluster",
		"payload": action,
	})
}

// SendDeletePeerResponse implements domain.InfraJobResponder
func (i *infraResponder) SendDeletePeerResponse(cxt context.Context, action domain.DeletePeerResponse) error {

	return i.kProducer.SendMessage(i.responseTopic, "resp", messaging.Json{
		"type":    "delete-peer",
		"payload": action,
	})
}

// SendUpdateClusterResponse implements domain.InfraJobResponder
func (i *infraResponder) SendUpdateClusterResponse(cxt context.Context, action domain.UpdateClusterResponse) error {
	return i.kProducer.SendMessage(i.responseTopic, "resp", messaging.Json{
		"type":    "update-cluster",
		"payload": action,
	})
}

func NewInfraResponder(k messaging.Producer[messaging.Json], responseTopic string) domain.InfraJobResponder {
	return &infraResponder{
		kProducer:     k,
		responseTopic: responseTopic,
	}
}

func fxJobResponder(p messaging.Producer[messaging.Json], env *InfraEnv) domain.InfraJobResponder {
	return NewInfraResponder(p, env.KafkaInfraResponseTopic)
}
