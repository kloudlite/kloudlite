package redpanda

import (
	"context"
	"strings"
	"time"

	"github.com/kloudlite/api/pkg/errors"
	"github.com/kloudlite/api/pkg/logging"
	"github.com/twmb/franz-go/pkg/kgo"
	"go.uber.org/fx"
)

type Consumer interface {
	Close()
	Ping(ctx context.Context) error
	StartConsuming(onMessage ReaderFunc)
	StartConsumingSync(onMessage ReaderFunc)

	LifecycleOnStart(ctx context.Context) error
	LifecycleOnStop(ctx context.Context) error
}

type ConsumerImpl struct {
	client      *kgo.Client
	logger      logging.Logger
	isConnected bool
}

// LifecycleOnStart implements Consumer.
func (c *ConsumerImpl) LifecycleOnStart(ctx context.Context) error {
	c.logger.Debugf("consumer pinging to kafka brokers")
	if err := c.Ping(ctx); err != nil {
		return err
	}
	c.logger.Infof("consumer connected to kafka brokers")
	return nil
}

// LifecycleOnStop implements Consumer.
func (c *ConsumerImpl) LifecycleOnStop(ctx context.Context) error {
	c.Close()
	c.logger.Infof("consumer closed")
	return nil
}

// type Message struct {
//	Action  string         `json:"action"`
//	Payload map[string]any `json:"payload"`
//	record  *kgo.Record
// }

type ReaderFunc func(msg []byte, timeStamp time.Time, offset int64) error

func (c *ConsumerImpl) Close() {
	c.client.Close()
}

func (c *ConsumerImpl) Ping(ctx context.Context) error {
	if err := c.client.Ping(ctx); err != nil {
		return err
	}
	c.isConnected = true
	return nil
}

func (c *ConsumerImpl) StartConsuming(onMessage ReaderFunc) {
	go func() {
		for {
			fetches := c.client.PollFetches(context.Background())
			if fetches.IsClientClosed() {
				return
			}

			fetches.EachError(
				func(topic string, partition int32, err error) {
					c.logger.Warnf("topic=%s, partition=%d read failed as %v", topic, partition, err)
				},
			)

			fetches.EachRecord(
				func(record *kgo.Record) {
					if err := onMessage(record.Value, record.Timestamp, record.Offset); err != nil {
						c.logger.Errorf(err, "error in consumer ReaderFunc")
						return
					}
					if err := c.client.CommitRecords(context.TODO(), record); err != nil {
						c.logger.Errorf(err, "error while committing records")
						return
					}
				},
			)
		}
	}()
}

func (c *ConsumerImpl) StartConsumingSync(onMessage ReaderFunc) {
	for {
		fetches := c.client.PollFetches(context.Background())
		if fetches.IsClientClosed() {
			return
		}

		fetches.EachError(
			func(topic string, partition int32, err error) {
				c.logger.Warnf("topic=%s, partition=%d read failed as %v", topic, partition, err)
			},
		)

		fetches.EachRecord(
			func(record *kgo.Record) {
				if err := onMessage(record.Value, record.Timestamp, record.Offset); err != nil {
					c.logger.Errorf(err, "error in consumer ReaderFunc")
					return
				}
				if err := c.client.CommitRecords(context.TODO(), record); err != nil {
					c.logger.Errorf(err, "error while commiting records")
					return
				}
			},
		)
	}
}

// func NewRawConsumer(client Client, consumerGroupId string) (Consumer, error) {
// 	opts := []kgo.Opt{
// 		kgo.SeedBrokers(strings.Split(client.GetBrokerHosts(), ",")...),
// 		kgo.ConsumerGroup(consumerGroupId),
// 		kgo.DisableAutoCommit(),
// 	}
//
// 	saslOpt, err := parseSASLAuth(client.GetKafkaSASLAuth())
// 	if err != nil {
// 		return nil, err
// 	}
//
// 	if saslOpt != nil {
// 		opts = append(opts, saslOpt)
// 	}
//
// 	cli, err := kgo.NewClient(opts...)
// 	if err != nil {
// 		return nil, errors.NewEf(err, "unable to create client")
// 	}
// 	return &ConsumerImpl{client: cli}, nil
// }

func NewConsumer(
	brokerHosts string,
	consumerGroup string,
	options ConsumerOpts,
	topics []string,
) (Consumer, error) {
	cOpts := options.getWithDefaults()

	opts := []kgo.Opt{
		kgo.SeedBrokers(strings.Split(brokerHosts, ",")...),
		kgo.ConsumerGroup(consumerGroup),
		kgo.ConsumeTopics(topics...),
		kgo.DisableAutoCommit(),
	}

	saslOpt, err := parseSASLAuth(cOpts.SASLAuth)
	if err != nil {
		return nil, err
	}

	if saslOpt != nil {
		opts = append(opts, saslOpt)
	}

	client, err := kgo.NewClient(opts...)
	if err != nil {
		return nil, errors.NewEf(err, "unable to create client")
	}

	logger := options.Logger
	if logger == nil {
		logger, err = logging.New(&logging.Options{Name: "kafka-consumer"})
		if err != nil {
			return nil, err
		}
	}

	return &ConsumerImpl{client: client, logger: logger}, nil
}

type ConsumerConfig interface {
	GetSubscriptionTopics() []string
	GetConsumerGroupId() string
}

func NewConsumerFx[T ConsumerConfig]() fx.Option {
	return fx.Module(
		"consumer",
		fx.Provide(
			func(cfg T, client Client, lf fx.Lifecycle) (Consumer, error) {
				topics := cfg.GetSubscriptionTopics()
				consumerGroup := cfg.GetConsumerGroupId()
				consumer, err := NewConsumer(client.GetBrokerHosts(), consumerGroup, ConsumerOpts{
					SASLAuth: client.GetKafkaSASLAuth(),
				}, topics)
				if err != nil {
					return nil, err
				}

				lf.Append(
					fx.Hook{
						OnStart: func(ctx context.Context) error {
							return consumer.Ping(ctx)
						},
						OnStop: func(context.Context) error {
							consumer.Close()
							return nil
						},
					},
				)
				return consumer, nil
			},
		),
	)
}
