package redpanda

import (
	"context"
	"fmt"
	"github.com/twmb/franz-go/pkg/kgo"
	"go.uber.org/fx"
	"go.uber.org/zap"
	"kloudlite.io/pkg/errors"
	"strings"
)

type Consumer interface {
	SetupLogger(logger *zap.SugaredLogger)
	Close()
	StartConsuming(onMessage ReaderFunc)
}

type ConsumerImpl struct {
	client *kgo.Client
	logger *zap.SugaredLogger
}

//type Message struct {
//	Action  string         `json:"action"`
//	Payload map[string]any `json:"payload"`
//	record  *kgo.Record
//}

type ReaderFunc func(msg []byte) error

func (c *ConsumerImpl) SetupLogger(logger *zap.SugaredLogger) {
	c.logger = logger
}

func (c *ConsumerImpl) Close() {
	c.client.Close()
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
					fmt.Println(record.Timestamp)
					if err := onMessage(record.Value); err != nil {
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

func NewConsumer(brokerHosts string, consumerGroup string, topics ...string) (Consumer, error) {
	opts := []kgo.Opt{
		kgo.SeedBrokers(strings.Split(brokerHosts, ",")...),
		kgo.ConsumerGroup(consumerGroup),
		kgo.ConsumeTopics(topics...),
		kgo.DisableAutoCommit(),
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
		"redis",
		fx.Provide(
			func(env T, client Client) (Consumer, error) {
				topics := env.GetSubscriptionTopics()
				consumerGroup := env.GetConsumerGroupId()
				return NewConsumer(client.GetBrokerHosts(), consumerGroup, topics...)
			},
		),
		fx.Invoke(
			func(lf fx.Lifecycle, r Consumer) {
				lf.Append(
					fx.Hook{
						OnStart: func(ctx context.Context) error {
							return nil
						},
						OnStop: func(ctx context.Context) error {
							r.Close()
							return nil
						},
					},
				)
			},
		),
	)
}
