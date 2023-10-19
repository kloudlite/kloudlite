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

	LifecycleOnStart(ctx context.Context) error
	LifecycleOnStop(ctx context.Context) error
}

type ProducerImpl struct {
	client *kgo.Client
	logger logging.Logger
}

// LifecycleOnStart implements Producer.
func (p *ProducerImpl) LifecycleOnStart(ctx context.Context) error {
	p.logger.Debugf("producer pinging to kafka brokers")
	if err := p.Ping(ctx); err != nil {
		return err
	}
	p.logger.Infof("producer connected to kafka brokers")
	return nil
}

// LifecycleOnStop implements Producer.
func (p *ProducerImpl) LifecycleOnStop(ctx context.Context) error {
	p.Close()
	p.logger.Infof("producer closed")
	return nil
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

	logger := producerOpts.Logger
	if logger == nil {
		logger, err = logging.New(&logging.Options{Name: "redpanda-logger"})
		if err != nil {
			return nil, err
		}
	}

	return &ProducerImpl{client: client, logger: logger}, nil
}

func NewProducerFx[T Client]() fx.Option {
	logger, _ := logging.New(&logging.Options{Name: "redpanda-logger", Dev: false})
	return fx.Module(
		"redpanda",
		fx.Provide(
			// func(client Client) (Producer, error) {
			func(client T) (Producer, error) {
				return NewProducer(
					client.GetBrokerHosts(), &ProducerOpts{
						SASLAuth: client.GetKafkaSASLAuth(),
						Logger:   logger,
					},
				)
			},
		),

		fx.Invoke(
			func(lf fx.Lifecycle, producer Producer, logger logging.Logger) {
				lf.Append(
					fx.Hook{
						OnStart: func(ctx context.Context) error {
							if err := producer.Ping(ctx); err != nil {
								return err
							}
							logger.Infof("successfully connected to kafka brokers")
							return nil
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
