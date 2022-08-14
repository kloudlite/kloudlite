package redpanda

import (
	"context"
	"strings"

	"github.com/twmb/franz-go/pkg/kgo"
	"operators.kloudlite.io/lib/errors"
	"operators.kloudlite.io/lib/logging"
)

type Consumer struct {
	client  *kgo.Client
	logger  logging.Logger
	options *ConsumerOptions
}

func (c *Consumer) Close() {
	c.client.Close()
}

type ConsumerOptions struct {
	ErrProducer *Producer
	ErrTopic    string
	logger      logging.Logger
}

type ConsumerError interface {
	GetKey() string
	error
}

func (c *Consumer) StartConsuming(onMessage ReaderFunc) {
	logger := func() logging.Logger {
		if c.logger != nil {
			return c.logger
		}
		if c.options.logger != nil {
			c.logger = c.options.logger
			return c.logger
		}
		c.logger = logging.NewOrDie(&logging.Options{Name: "default-logger"})
		return c.logger
	}()

	for {
		fetches := c.client.PollFetches(context.Background())
		if fetches.IsClientClosed() {
			return
		}

		fetches.EachError(
			func(topic string, partition int32, err error) {
				logger.Warnf("topic=%s, partition=%d read failed as %v", topic, partition, err)
			},
		)

		fetches.EachRecord(
			func(record *kgo.Record) {
				if err := onMessage(
					&KafkaMessage{
						Key:        record.Key,
						Value:      record.Value,
						Timestamp:  record.Timestamp,
						Topic:      record.Topic,
						Partition:  record.Partition,
						ProducerId: record.ProducerID,
						Offset:     record.Offset,
					},
				); err != nil {
					logger.Errorf(err, "in onMessage()")

					if err := c.client.CommitRecords(context.TODO(), record); err != nil {
						return
					}
					return
				}
				if err := c.client.CommitRecords(context.TODO(), record); err != nil {
					if c.logger != nil {
						logger.Errorf(err, "commiting records")
					}
					return
				}
			},
		)
	}
}

func NewConsumer(
	brokerHosts string, consumerGroup string, topicName string, options *ConsumerOptions,
) (*Consumer, error) {
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
