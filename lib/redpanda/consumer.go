package redpanda

import (
	"context"
	"encoding/json"
	"github.com/twmb/franz-go/pkg/kgo"
	"go.uber.org/zap"
	"operators.kloudlite.io/lib/errors"
	"strings"
)

type Consumer struct {
	client *kgo.Client
	logger *zap.SugaredLogger
}

type Message struct {
	Action  string                 `json:"action"`
	Payload map[string]interface{} `json:"payload"`

	record *kgo.Record
}

type ReaderFunc func(m *Message) error

func (c *Consumer) SetupLogger(logger *zap.SugaredLogger) {
	c.logger = logger
}

func (c *Consumer) Close() {
	c.client.Close()
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
				var j Message
				if err := json.Unmarshal(record.Value, &j); err != nil {
					if c.logger != nil {
						c.logger.Error("could not unmarshal message []byte into type Message")
					}
					return
				}

				j.record = record

				if err := onMessage(&j); err != nil {
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
}

func NewConsumer(brokerHosts string, consumerGroup string, topicName string) (*Consumer, error) {
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
	return &Consumer{client: client}, nil
}
