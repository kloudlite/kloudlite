package nats

import (
	"context"
	"fmt"
	"time"

	"github.com/nats-io/nats.go/jetstream"
	"kloudlite.io/pkg/messaging/types"
)

type JetstreamProducer struct {
	js jetstream.JetStream
}

// Stop implements messaging.Producer.
func (c *JetstreamProducer) Stop(ctx context.Context) error {
	sctx, cf := context.WithTimeout(ctx, 5*time.Second)
	defer cf()

	select {
	case <-c.js.PublishAsyncComplete():
		fmt.Println("All Messages Acknowledged")
	case <-sctx.Done():
		fmt.Println("server is dying, cannot wait more, Message Acknowledgement Timeout")
	}
	return nil
}

// ProduceAsync implements messaging.Producer.
func (c *JetstreamProducer) ProduceAsync(ctx context.Context, msg types.ProduceMsg) error {
	pa, err := c.js.PublishAsync(msg.Subject, msg.Payload)
	if err != nil {
		return err
	}

	go func() {
		fmt.Println("waiting for acknowledgement")
		select {
		case ack := <-pa.Ok():
			fmt.Println("Message Acknowledged, at stream: ", ack.Stream, " seq: ", ack.Sequence)
		case <-pa.Err():
			fmt.Println("Message Failed to be Acknowledged")
		}
	}()
	return nil
}

// Produce implements messaging.Producer.
func (c *JetstreamProducer) Produce(ctx context.Context, msg types.ProduceMsg) error {
	_, err := c.js.Publish(ctx, msg.Subject, msg.Payload)
	return err
}
