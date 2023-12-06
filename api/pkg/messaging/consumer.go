package messaging

import (
	"context"

	"kloudlite.io/pkg/messaging/nats"
	"kloudlite.io/pkg/messaging/types"
)

type Consumer interface {
	Consume(ctx context.Context, consumeFn func(msg *types.ConsumeMsg) error) error
	Stop() error
}

var _ Consumer = (*nats.JetstreamConsumer)(nil)
