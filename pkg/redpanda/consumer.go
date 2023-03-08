package redpanda

import (
	"context"
	"strings"
	"time"

	"github.com/twmb/franz-go/pkg/kgo"
	"go.uber.org/fx"
	"go.uber.org/zap"
	"kloudlite.io/pkg/errors"
)

type Consumer interface {
	SetupLogger(logger *zap.SugaredLogger)
	Close()
	Ping(ctx context.Context) error
	StartConsuming(onMessage ReaderFunc)
}

type ConsumerImpl struct {
	client      *kgo.Client
	logger      *zap.SugaredLogger
	isConnected bool
}

// type Message struct {
//	Action  string         `json:"action"`
//	Payload map[string]any `json:"payload"`
//	record  *kgo.Record
// }

type ReaderFunc func(msg []byte, timeStamp time.Time, offset int64) error

func (c *ConsumerImpl) SetupLogger(logger *zap.SugaredLogger) {
	c.logger = logger
}

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
					if c.logger != nil {
						c.logger.Warnf("topic=%s, partition=%d read failed as %v", topic, partition, err)
					}
				},
			)

			fetches.EachRecord(
				func(record *kgo.Record) {
					if err := onMessage(record.Value, record.Timestamp, record.Offset); err != nil {
						if c.logger != nil {
							c.logger.Error("error in onMessage(): %+v\n", err)
						}
						return
					}
					if err := c.client.CommitRecords(context.TODO(), record); err != nil {
						if c.logger != nil {
							c.logger.Error("error while commiting records: %+v\n", err)
						}
						return
					}
				},
			)
		}
	}()
}

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
	return &ConsumerImpl{client: client}, nil
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

				// lf.Append(
				// 	fx.Hook{
				// 		OnStart: func(ctx context.Context) error {
				// 			return consumer.Ping(ctx)
				// 		},
				// 		OnStop: func(context.Context) error {
				// 			consumer.Close()
				// 			return nil
				// 		},
				// 	},
				// )
				return consumer, nil
			},
		),
	)
}
