package redpanda

import (
	"context"
	"strings"

	"github.com/twmb/franz-go/pkg/kgo"
	"operators.kloudlite.io/lib/errors"
	"operators.kloudlite.io/lib/logging"
)

type consumer struct {
	client *kgo.Client
	logger logging.Logger
}

func (c *consumer) Ping(ctx context.Context) error {
	return c.client.Ping(ctx)
}

func (c *consumer) Close() {
	c.client.Close()
}

type ConsumerError interface {
	GetKey() string
	error
}

func (c *consumer) StartConsuming(onMessage ReaderFunc) error {
	for {
		fetches := c.client.PollFetches(context.Background())
		if fetches.IsClientClosed() {
			return errors.Newf("client is closed")
		}

		logger := c.logger

		fetches.EachError(
			func(topic string, partition int32, err error) {
				logger.Warnf("topic=%s, partition=%d read failed as %v", topic, partition, err)
			},
		)

		fetches.EachRecord(
			func(record *kgo.Record) {
				if err := onMessage(
					KafkaMessage{
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
	brokerHosts string, consumerGroup string, topicName string, options *ConsumerOpts,
) (Consumer, error) {
	cOpts := options.getWithDefaults()

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
	return &consumer{client: client, logger: cOpts.Logger}, nil
}
