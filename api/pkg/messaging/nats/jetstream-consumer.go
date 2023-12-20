package nats

import (
	"context"
	"github.com/kloudlite/api/pkg/errors"
	"os"
	"os/signal"

	"github.com/kloudlite/api/pkg/messaging/types"
	"github.com/kloudlite/api/pkg/nats"
	"github.com/nats-io/nats.go/jetstream"
)

type JetstreamConsumer struct {
	name       string
	client     *nats.JetstreamClient
	consumer   jetstream.Consumer
	consumeCtx jetstream.ConsumeContext
}

// Consume implements messaging.Consumer.
func (jc *JetstreamConsumer) Consume(consumeFn func(msg *types.ConsumeMsg) error, opts types.ConsumeOpts) error {
	cctx, err := jc.consumer.Consume(func(msg jetstream.Msg) {
		mm, err := msg.Metadata()
		if err != nil {
			msg.Nak()
			return
		}

		msg.InProgress()

		if err := consumeFn(&types.ConsumeMsg{
			Subject:   msg.Subject(),
			Timestamp: mm.Timestamp,
			Payload:   msg.Data(),
		}); err != nil {
			if opts.OnError == nil {
				jc.client.Logger.Errorf(err, "while consuming message from subject: %s, sending NACK", msg.Subject())
				msg.Nak()
				return
			}

			if opts.OnError != nil {
				if err := opts.OnError(err); err != nil {
					msg.Nak()
					return
				}
			}
		}

		msg.Ack()
		jc.client.Logger.Infof("acknowledged message, stream: %s, consumer: %s", mm.Stream, mm.Consumer)
	})
	if err != nil {
		return err
	}

	defer cctx.Stop()

	jc.consumeCtx = cctx

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt)
	signal := <-quit
	return errors.Newf("os signal: %s received, stopped consuming messages", signal)
}

// Stop implements Consumer.
func (nc *JetstreamConsumer) Stop(context.Context) error {
	nc.consumeCtx.Stop()
	return nil
}

type ConsumerConfig jetstream.ConsumerConfig

type JetstreamConsumerArgs struct {
	Stream         string
	ConsumerConfig ConsumerConfig
}

func NewJetstreamConsumer(ctx context.Context, jc *nats.JetstreamClient, args JetstreamConsumerArgs) (*JetstreamConsumer, error) {
	s, err := jc.Jetstream.Stream(ctx, args.Stream)
	if err != nil {
		return nil, err
	}

	c, err := s.CreateOrUpdateConsumer(ctx, jetstream.ConsumerConfig(args.ConsumerConfig))
	if err != nil {
		return nil, err
	}

	return &JetstreamConsumer{
		name:     args.ConsumerConfig.Name,
		client:   jc,
		consumer: c,
	}, nil
}
