package redpanda

import (
	"context"
	"strings"

	"github.com/kloudlite/operator/pkg/errors"
	"github.com/kloudlite/operator/pkg/logging"
	"github.com/twmb/franz-go/pkg/kgo"
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

func (c *consumer) StartConsuming(reader ReaderFunc) {
	for {
		fetches := c.client.PollFetches(context.Background())
		if fetches.IsClientClosed() {
			return
		}

		logger := c.logger

		fetches.EachError(
			func(topic string, partition int32, err error) {
				logger.Warnf("topic=%s, partition=%d read failed as %v", topic, partition, err)
			},
		)

		fetches.EachRecord(
			func(record *kgo.Record) {
				if err := reader(
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
					logger.Errorf(err, "in readerFunc()")

					// if err := c.client.CommitRecords(context.TODO(), record); err != nil {
					// 	return
					// }
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

func NewConsumer(brokerHosts string, consumerGroup string, topicName string, options ConsumerOpts) (Consumer, error) {
	cOpts := options.getWithDefaults()

	opts := []kgo.Opt{
		kgo.SeedBrokers(strings.Split(brokerHosts, ",")...),
		kgo.ConsumerGroup(consumerGroup),
		kgo.ConsumeTopics(topicName),
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
	return &consumer{client: client, logger: cOpts.Logger}, nil
}
