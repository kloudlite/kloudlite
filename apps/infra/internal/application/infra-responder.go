package application

import (
	"context"
	"kloudlite.io/apps/infra/internal/domain"
	"kloudlite.io/pkg/messaging"
	//"kloudlite.io/pkg/messaging"
)

type infraResponder struct {
	kProducer     messaging.Producer[any]
	responseTopic string
}

// SendAddPeerResponse implements domain.InfraJobResponder
func (i *infraResponder) SendAddPeerResponse(cxt context.Context, action domain.AddPeerResponse) error {
	return i.kProducer.SendMessage(i.responseTopic, "resp", map[string]any{
		"type":    "add-peer",
		"payload": action,
	})
}

// SendCreateClusterResponse implements domain.InfraJobResponder
func (i *infraResponder) SendCreateClusterResponse(cxt context.Context, action domain.SetupClusterResponse) error {
	return i.kProducer.SendMessage(i.responseTopic, "resp", map[string]any{
		"type":    "create-cluster",
		"payload": action,
	})
}

// SendDeleteClusterResponse implements domain.InfraJobResponder
func (i *infraResponder) SendDeleteClusterResponse(cxt context.Context, action domain.DeleteClusterResponse) error {
	return i.kProducer.SendMessage(i.responseTopic, "resp", map[string]any{
		"type":    "delete-cluster",
		"payload": action,
	})
}

// SendDeletePeerResponse implements domain.InfraJobResponder
func (i *infraResponder) SendDeletePeerResponse(cxt context.Context, action domain.DeletePeerResponse) error {

	return i.kProducer.SendMessage(i.responseTopic, "resp", map[string]any{
		"type":    "delete-peer",
		"payload": action,
	})
}

// SendUpdateClusterResponse implements domain.InfraJobResponder
func (i *infraResponder) SendUpdateClusterResponse(cxt context.Context, action domain.UpdateClusterResponse) error {
	return i.kProducer.SendMessage(i.responseTopic, "resp", map[string]any{
		"type":    "update-cluster",
		"payload": action,
	})
}

func NewInfraResponder(k messaging.Producer[any], responseTopic string) domain.InfraJobResponder {
	return &infraResponder{
		kProducer:     k,
		responseTopic: responseTopic,
	}
}
