package redpanda

import (
	"context"
	"strings"
	"time"

	"kloudlite.io/pkg/logging"

	"github.com/twmb/franz-go/pkg/kgo"
	"go.uber.org/fx"
)

type Producer interface {
	Ping(ctx context.Context) error
	Close()
	Produce(ctx context.Context, topic string, key string, value []byte) (*ProducerOutput, error)
}

type ProducerImpl struct {
	client *kgo.Client
}

func (p *ProducerImpl) Ping(ctx context.Context) error {
	return p.client.Ping(ctx)
}

func (p *ProducerImpl) Close() {
	p.client.Close()
}

type ProducerOutput struct {
	Key        []byte    `json:"key,omitempty"`
	Timestamp  time.Time `json:"timestamp"`
	Topic      string    `json:"topic"`
	Partition  int32     `json:"partition,omitempty"`
	ProducerId int64     `json:"producerId,omitempty"`
	Offset     int64     `json:"offset"`
}

func (p *ProducerImpl) Produce(ctx context.Context, topic string, key string, value []byte) (*ProducerOutput, error) {
	record := kgo.KeySliceRecord(
		func() []byte {
			if key == "" {
				return nil
			}
			return []byte(key)
		}(), value,
	)
	record.Topic = topic
	sync, err := p.client.ProduceSync(ctx, record).First()
	if err != nil {
		return nil, err
	}
	return &ProducerOutput{
		Key:        sync.Key,
		Timestamp:  sync.Timestamp,
		Topic:      sync.Topic,
		Partition:  sync.Partition,
		ProducerId: sync.ProducerID,
		Offset:     sync.Offset,
	}, nil
}

func NewProducer(brokerHosts string, producerOpts *ProducerOpts) (Producer, error) {
	opts := []kgo.Opt{
		kgo.SeedBrokers(strings.Split(brokerHosts, ",")...),
	}
	saslOpt, err := parseSASLAuth(producerOpts.SASLAuth)
	if err != nil {
		return nil, err
	}

	if saslOpt != nil {
		opts = append(opts, saslOpt)
	}

	client, err := kgo.NewClient(opts...)

	if err != nil {
		return nil, err
	}
	if err := client.Ping(context.TODO()); err != nil {
		return nil, err
	}
	return &ProducerImpl{client: client}, nil
}

func NewProducerFx[T Client]() fx.Option {
	return fx.Module(
		"redpanda",
		fx.Provide(
			// func(client Client) (Producer, error) {
			func(client T) (Producer, error) {
				return NewProducer(
					client.GetBrokerHosts(), &ProducerOpts{
						SASLAuth: client.GetKafkaSASLAuth(),
						Logger:   logging.Logger{},
					},
				)
			},
		),
		fx.Invoke(
			func(lf fx.Lifecycle, producer Producer) {
				lf.Append(
					fx.Hook{
						OnStart: func(ctx context.Context) error {
							return producer.Ping(ctx)
						},
						OnStop: func(context.Context) error {
							producer.Close()
							return nil
						},
					},
				)
			},
		),
	)
}
