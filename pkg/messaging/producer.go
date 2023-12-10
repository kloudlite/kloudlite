package messaging

import (
	"context"

	"kloudlite.io/pkg/messaging/nats"
	"kloudlite.io/pkg/messaging/types"
)

type Producer interface {
	Produce(ctx context.Context, msg types.ProduceMsg) error
	ProduceAsync(ctx context.Context, msg types.ProduceMsg) error

	Stop(ctx context.Context) error
}

var _ Producer = (*nats.JetstreamProducer)(nil)
