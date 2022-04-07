package domain

import (
	"errors"

	"go.uber.org/fx"
	"kloudlite.io/pkg/shared"
)

type domain struct{}

func (domain *domain) ProcessMessage(msg Message) error {
	switch msg.ResourceType {
	case shared.RESOURCE_PROJECT:
		{
			if spec, ok := msg.Spec.(Project); ok {
			}
			return errors.New("malformed spec not of type(Project)")
		}
	case shared.RESOURCE_CONFIG:
		{
		}
	case shared.RESOURCE_SECRET:
		{
		}
	case shared.RESOURCE_GIT_PIPELINE:
		{
		}
	case shared.RESOURCE_APP:
		{
		}
	case shared.RESOURCE_MANAGED_SERVICE:
		{
		}
	case shared.RESOURCE_MANAGED_RESOURCE:
		{
		}
	}
	return nil
}

type Domain interface {
	ProcessMessage(msg Message) error
}

var fxDomain = func() Domain {
	return &domain{}
}

var Module = fx.Module("domain")
