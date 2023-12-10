package messaging

import (
	"context"

	"kloudlite.io/pkg/messaging/nats"
	"kloudlite.io/pkg/messaging/types"
)

type Consumer interface {
	Consume(consumeFn func(msg *types.ConsumeMsg) error, opts types.ConsumeOpts) error
	Stop(ctx context.Context) error
}

var _ Consumer = (*nats.JetstreamConsumer)(nil)
