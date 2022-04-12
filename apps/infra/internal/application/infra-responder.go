package application

import (
	"kloudlite.io/apps/infra/internal/domain"
	"kloudlite.io/pkg/messaging"
)

type infraResponder struct {
	kProducer messaging.Producer[any]
}

// SendAddPeerResponse implements domain.InfraJobResponder
func (i *infraResponder) SendAddPeerResponse(action domain.AddPeerResponse) {
	panic("unimplemented")
}

// SendCreateClusterResponse implements domain.InfraJobResponder
func (i *infraResponder) SendCreateClusterResponse(action domain.SetupClusterResponse) {
	panic("unimplemented")
}

// SendDeleteClusterResponse implements domain.InfraJobResponder
func (i *infraResponder) SendDeleteClusterResponse(action domain.DeleteClusterResponse) {
	panic("unimplemented")
}

// SendDeletePeerResponse implements domain.InfraJobResponder
func (i *infraResponder) SendDeletePeerResponse(action domain.DeletePeerResponse) {
	panic("unimplemented")
}

// SendUpdateClusterResponse implements domain.InfraJobResponder
func (i *infraResponder) SendUpdateClusterResponse(action domain.UpdateClusterResponse) {
	panic("unimplemented")
}

func NewInfraResponder(k messaging.Producer[any]) domain.InfraJobResponder {
	return &infraResponder{kProducer: k}
}
