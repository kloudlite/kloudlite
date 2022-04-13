package application

import (
	"encoding/json"
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
	marshal, err := json.Marshal(map[string]any{
		"type":    "add-peer",
		"payload": action,
	})
	if err != nil {
		return err
	}
	return i.kProducer.SendMessage(i.responseTopic, "resp", marshal)
}

// SendCreateClusterResponse implements domain.InfraJobResponder
func (i *infraResponder) SendCreateClusterResponse(action domain.SetupClusterResponse) error {
	marshal, err := json.Marshal(map[string]any{
		"type":    "add-peer",
		"payload": action,
	})
	if err != nil {
		return err
	}
	return i.kProducer.SendMessage(i.responseTopic, "resp", marshal)
}

// SendDeleteClusterResponse implements domain.InfraJobResponder
func (i *infraResponder) SendDeleteClusterResponse(action domain.DeleteClusterResponse) error {
	marshal, err := json.Marshal(map[string]any{
		"type":    "add-peer",
		"payload": action,
	})
	if err != nil {
		return err
	}
	return i.kProducer.SendMessage(i.responseTopic, "resp", marshal)
}

// SendDeletePeerResponse implements domain.InfraJobResponder
func (i *infraResponder) SendDeletePeerResponse(action domain.DeletePeerResponse) error {
	marshal, err := json.Marshal(map[string]any{
		"type":    "add-peer",
		"payload": action,
	})
	if err != nil {
		return err
	}
	return i.kProducer.SendMessage(i.responseTopic, "resp", marshal)
}

// SendUpdateClusterResponse implements domain.InfraJobResponder
func (i *infraResponder) SendUpdateClusterResponse(action domain.UpdateClusterResponse) error {
	marshal, err := json.Marshal(map[string]any{
		"type":    "add-peer",
		"payload": action,
	})
	if err != nil {
		return err
	}
	return i.kProducer.SendMessage(i.responseTopic, "resp", marshal)
}

func NewInfraResponder(k messaging.Producer[any], responseTopic string) domain.InfraJobResponder {
	return &infraResponder{kProducer: k, responseTopic: responseTopic}
}
