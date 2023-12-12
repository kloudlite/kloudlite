package messaging

import (
	"context"

	"github.com/kloudlite/api/pkg/messaging/nats"
	"github.com/kloudlite/api/pkg/messaging/types"
)

type Consumer interface {
	Consume(consumeFn func(msg *types.ConsumeMsg) error, opts types.ConsumeOpts) error
	Stop(ctx context.Context) error
}

var _ Consumer = (*nats.JetstreamConsumer)(nil)
