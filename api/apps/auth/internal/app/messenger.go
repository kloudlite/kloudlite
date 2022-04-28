package app

import (
	"context"
	"kloudlite.io/apps/auth/internal/domain"
	"kloudlite.io/pkg/messaging"
)

type messengerI struct {
}

func (m *messengerI) SendEmail(ctx context.Context, template string, payload messaging.Json) error {
	return nil
}

func fxMessenger() domain.Messenger {
	return &messengerI{}
}
