package nats

import (
	"context"
	"fmt"
	"os"
	"os/signal"

	"github.com/nats-io/nats.go/jetstream"
	"kloudlite.io/pkg/logging"
	"kloudlite.io/pkg/messaging/types"
)

type JetstreamConsumer struct {
	name       string
	js         jetstream.JetStream
	logger     logging.Logger
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
				jc.logger.Errorf(err, "while consuming message from subject: %s, sending NACK", msg.Subject())
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
		jc.logger.Infof("acknowledged message, stream: %s, consumer: %s", mm.Stream, mm.Consumer)
	})
	if err != nil {
		return err
	}

	defer cctx.Stop()

	jc.consumeCtx = cctx

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt)
	signal := <-quit
	return fmt.Errorf("os signal: %s received, stopped consuming messages", signal)
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
