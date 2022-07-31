package redpanda

import (
	"context"
	"github.com/twmb/franz-go/pkg/kgo"
	"operators.kloudlite.io/lib/errors"
	"operators.kloudlite.io/lib/logging"
	"strings"
)

type Consumer struct {
	client  *kgo.Client
	logger  logging.Logger
	options *ConsumerOptions
}

type ReaderFunc func(msg []byte, key []byte) error

func (c *Consumer) SetupLogger(logger logging.Logger) {
	c.logger = logger
}

func (c *Consumer) Close() {
	c.client.Close()
}

type ConsumerOptions struct {
	ErrProducer *Producer
	ErrTopic    string
}

type ConsumerError interface {
	GetKey() string
	error
}

func (c *Consumer) StartConsuming(onMessage ReaderFunc) {
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
				if err := onMessage(record.Value, record.Key); err != nil {
					if c.logger != nil {
						c.logger.Errorf(err, "in onMessage()")
					}

					if err := c.client.CommitRecords(context.TODO(), record); err != nil {
						return
					}
					return
				}
				if err := c.client.CommitRecords(context.TODO(), record); err != nil {
					if c.logger != nil {
						c.logger.Errorf(err, "commiting records")
					}
					return
				}
			},
		)
	}
}

func NewConsumer(brokerHosts string, consumerGroup string, topicName string, options *ConsumerOptions) (*Consumer,
	error) {
	opts := []kgo.Opt{
		kgo.SeedBrokers(strings.Split(brokerHosts, ",")...),
		kgo.ConsumerGroup(consumerGroup),
		kgo.ConsumeTopics(topicName),
		kgo.DisableAutoCommit(),
	}

	client, err := kgo.NewClient(opts...)
	if err != nil {
		return nil, errors.NewEf(err, "unable to create client")
	}
	return &Consumer{client: client, options: options}, nil
}
