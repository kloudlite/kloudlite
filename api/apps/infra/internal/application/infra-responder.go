package application

import (
	"kloudlite.io/apps/infra/internal/domain"
	"kloudlite.io/pkg/messaging"
	//"kloudlite.io/pkg/messaging"
)

type infraResponder struct {
	kProducer     messaging.Producer[any]
	responseTopic string
}

// SendAddPeerResponse implements domain.InfraJobResponder
func (i *infraResponder) SendAddPeerResponse(action domain.AddPeerResponse) error {
	return i.kProducer.SendMessage(i.responseTopic, "resp", map[string]any{
		"type":    "add-peer",
		"payload": action,
	})
}

// SendCreateClusterResponse implements domain.InfraJobResponder
func (i *infraResponder) SendCreateClusterResponse(action domain.SetupClusterResponse) error {
	return i.kProducer.SendMessage(i.responseTopic, "resp", map[string]any{
		"type":    "create-cluster",
		"payload": action,
	})
}

// SendDeleteClusterResponse implements domain.InfraJobResponder
func (i *infraResponder) SendDeleteClusterResponse(action domain.DeleteClusterResponse) error {
	return i.kProducer.SendMessage(i.responseTopic, "resp", map[string]any{
		"type":    "delete-cluster",
		"payload": action,
	})
}

// SendDeletePeerResponse implements domain.InfraJobResponder
func (i *infraResponder) SendDeletePeerResponse(action domain.DeletePeerResponse) error {

	return i.kProducer.SendMessage(i.responseTopic, "resp", map[string]any{
		"type":    "delete-peer",
		"payload": action,
	})
}

// SendUpdateClusterResponse implements domain.InfraJobResponder
func (i *infraResponder) SendUpdateClusterResponse(action domain.UpdateClusterResponse) error {
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
