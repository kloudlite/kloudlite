package nats

import (
	"context"

	"github.com/nats-io/nats.go/jetstream"
	"kloudlite.io/pkg/logging"
)

type JetstreamClient struct {
	js     jetstream.JetStream
	logger logging.Logger
}

type ConsumerManager interface {
	GetConsumerInfo(ctx context.Context, stream string, consumer string) (*jetstream.ConsumerInfo, error)
	ListConsumers(ctx context.Context, stream string) ([]*jetstream.ConsumerInfo, error)
	DeleteConsumer(ctx context.Context, stream string, consumer string) error
}

var _ ConsumerManager = (*JetstreamClient)(nil)

func (jsc *JetstreamClient) CreateProducer() *JetstreamProducer {
	return &JetstreamProducer{
		js: jsc.js,
	}
}

func (jsc *JetstreamClient) CreateConsumer(ctx context.Context, args JetstreamConsumerArgs) (*JetstreamConsumer, error) {
	s, err := jsc.js.Stream(ctx, args.Stream)
	if err != nil {
		return nil, err
	}

	c, err := s.CreateOrUpdateConsumer(ctx, jetstream.ConsumerConfig(args.ConsumerConfig))
	if err != nil {
		return nil, err
	}

	return &JetstreamConsumer{
		name:       args.ConsumerConfig.Name,
		js:         jsc.js,
		logger:     jsc.logger.WithName(args.ConsumerConfig.Name),
		consumer:   c,
		consumeCtx: nil,
	}, nil
}

// DeleteConsumer implements ConsumerManager.
func (jc *JetstreamClient) DeleteConsumer(ctx context.Context, stream string, consumer string) error {
	err := jc.js.DeleteConsumer(ctx, stream, consumer)
	return err
}

// ListConsumers implements ConsumerManager.
func (jc *JetstreamClient) ListConsumers(ctx context.Context, stream string) ([]*jetstream.ConsumerInfo, error) {
	s, err := jc.js.Stream(ctx, stream)
	if err != nil {
		return nil, err
	}

	consumers := make([]*jetstream.ConsumerInfo, 0, 5)

	cil := s.ListConsumers(ctx)
	for ci := range cil.Info() {
		consumers = append(consumers, ci)
	}

	return consumers, nil
}

func (jc *JetstreamClient) GetConsumerInfo(ctx context.Context, stream string, consumer string) (*jetstream.ConsumerInfo, error) {
	s, err := jc.js.Stream(ctx, stream)
	if err != nil {
		return nil, err
	}

	c, err := s.Consumer(ctx, consumer)
	if err != nil {
		return nil, err
	}

	return c.Info(ctx)
}

func NewJetstreamClient(nc *Client) (*JetstreamClient, error) {
	js, err := jetstream.New(nc.Conn)
	if err != nil {
		return nil, err
	}

	return &JetstreamClient{
		js:     js,
		logger: nc.logger,
	}, nil
}
