package nats

import (
	"context"
	"os"
	"os/signal"

	"github.com/kloudlite/api/pkg/errors"

	"github.com/kloudlite/api/pkg/messaging/types"
	"github.com/kloudlite/api/pkg/nats"
	"github.com/nats-io/nats.go/jetstream"
)

type JetstreamConsumer struct {
	name       string
	stream     string
	client     *nats.JetstreamClient
	consumer   jetstream.Consumer
	consumeCtx jetstream.ConsumeContext
}

// Consume implements messaging.Consumer.
func (jc *JetstreamConsumer) Consume(consumeFn func(msg *types.ConsumeMsg) error, opts types.ConsumeOpts) error {
	cctx, err := jc.consumer.Consume(func(msg jetstream.Msg) {
		logger := jc.client.Logger.With("subject", msg.Subject())

		mm, err := msg.Metadata()
		if err != nil {
			if err := msg.Nak(); err != nil {
				logger.Error("failed to send NAK", "err", err)
				return
			}
			return
		}

		logger = logger.With("consumer", mm.Consumer, "stream", mm.Stream)

		if err = msg.InProgress(); err != nil {
			if err := msg.Nak(); err != nil {
				logger.Error("failed to send NAK", "err", err)
				return
			}
			return
		}

		if err := consumeFn(&types.ConsumeMsg{
			Subject:   msg.Subject(),
			Timestamp: mm.Timestamp,
			Payload:   msg.Data(),
		}); err != nil {
			if opts.OnError == nil {
				if err := msg.Nak(); err != nil {
					logger.Error("failed to send NAK", "err", err)
					return
				}
				return
			}

			if opts.OnError != nil {
				if err := opts.OnError(err); err != nil {
					if err := msg.Nak(); err != nil {
						logger.Error("failed to send NAK", "err", err)
						return
					}
					return
				}
			}
		}

		if err := msg.Ack(); err != nil {
			logger.Error("failed to send ACK, got", "err", err)
			return
		}
		logger.Debug("CONSUMED message", "stream", mm.Stream, "consumer", mm.Consumer)
	})
	if err != nil {
		return errors.NewE(err)
	}

	defer cctx.Stop()

	jc.consumeCtx = cctx

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt)
	s := <-quit
	return errors.Newf("os signal: %s received, stopped consuming messages", s)
}

// Stop implements Consumer.
func (nc *JetstreamConsumer) Stop(context.Context) error {
	if nc.consumeCtx != nil {
		nc.consumeCtx.Stop()
	}
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
		return nil, errors.NewE(err)
	}

	c, err := s.CreateOrUpdateConsumer(ctx, jetstream.ConsumerConfig(args.ConsumerConfig))
	if err != nil {
		return nil, errors.NewE(err)
	}

	return &JetstreamConsumer{
		name:     args.ConsumerConfig.Name,
		client:   jc,
		consumer: c,
		stream:   args.Stream,
	}, nil
}

func DeleteConsumer(ctx context.Context, jc *nats.JetstreamClient, consumer *JetstreamConsumer) error {
	return jc.Jetstream.DeleteConsumer(ctx, consumer.stream, consumer.name)
}
