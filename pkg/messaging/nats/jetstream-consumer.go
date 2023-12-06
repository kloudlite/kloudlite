package nats

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"

	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"
	"kloudlite.io/pkg/messaging/types"
)

type JetstreamConsumer struct {
	consumer   jetstream.Consumer
	consumeCtx jetstream.ConsumeContext
}

// Consume implements messaging.Consumer.
func (jc *JetstreamConsumer) Consume(ctx context.Context, consumeFn func(msg *types.ConsumeMsg) error) error {
	var sendAck bool
	flag.BoolVar(&sendAck, "ack", false, "--ack")
	flag.Parse()
	cctx, err := jc.consumer.Consume(func(msg jetstream.Msg) {
		if err := consumeFn(&types.ConsumeMsg{
			NatsJetstreamMsg: &types.NatsJetstreamConsumeMsg{
				Payload: msg.Data(),
			},
		}); err != nil {
			fmt.Println(err)
			msg.Nak()
		}
		if sendAck {
			msg.DoubleAck(ctx)
		}
		mm, err := msg.Metadata()
		if err != nil {
			msg.Nak()
		}
		fmt.Printf("message sequence, stream: %d, consumer: %d\n", mm.Sequence.Stream, mm.Sequence.Consumer)
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
func (nc *JetstreamConsumer) Stop() error {
	nc.consumeCtx.Stop()
	return nil
}

type JetstreamConsumerArgs struct {
	Stream string
	jetstream.ConsumerConfig
}

func NewJetstreamConsumer(ctx context.Context, jsc *JetstreamClient, args JetstreamConsumerArgs) (*JetstreamConsumer, error) {
	s, err := jsc.js.Stream(ctx, args.Stream)
	if err != nil {
		return nil, err
	}

	c, err := s.CreateOrUpdateConsumer(ctx, args.ConsumerConfig)
	if err != nil {
		return nil, err
	}

	return &JetstreamConsumer{
		consumer: c,
	}, nil
}
