package messaging

import (
	"context"

	"github.com/kloudlite/api/pkg/messaging/nats"
	"github.com/kloudlite/api/pkg/messaging/types"
)

type Producer interface {
	Produce(ctx context.Context, msg types.ProduceMsg) error
	ProduceAsync(ctx context.Context, msg types.ProduceMsg) error

	Stop(ctx context.Context) error
}

var _ Producer = (*nats.JetstreamProducer)(nil)
