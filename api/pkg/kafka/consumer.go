package kafka

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/twmb/franz-go/pkg/kgo"
	"kloudlite.io/pkg/errors"
	fn "kloudlite.io/pkg/functions"
	"kloudlite.io/pkg/logging"
)

type RecordMetadata struct {
	Key       []byte
	Headers   map[string][]byte
	Timestamp time.Time
}

type ConsumerContext struct {
	context.Context
	Logger logging.Logger
}

type ReaderFunc func(ctx ConsumerContext, topic string, value []byte, metadata RecordMetadata) error

type Consumer interface {
	Ping(ctx context.Context) error
	Close() error

	StartConsuming(readerMsg ReaderFunc)
	StopConsuming()
	LifecycleOnStart(ctx context.Context) error
	LifecycleOnStop(ctx context.Context) error
}

type consumer struct {
	ctx           context.Context
	cancelContext context.CancelFunc

	client *kgo.Client
	logger logging.Logger
}

func (c *consumer) Ping(ctx context.Context) error {
	return c.client.Ping(ctx)
}

func (c *consumer) Close() error {
	if c.client == nil {
		return fmt.Errorf("client is nil")
	}
	c.client.Close()
	return nil
}

func (c *consumer) StartConsuming(readMessage ReaderFunc) {
	for {
		fetches := c.client.PollFetches(c.ctx)
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
				headers := make(map[string][]byte, len(record.Headers))
				for i := range record.Headers {
					headers[record.Headers[i].Key] = record.Headers[i].Value
				}

				ctx := ConsumerContext{Context: c.ctx, Logger: c.logger}
				if err := readMessage(ctx, record.Topic, record.Value, RecordMetadata{
					Key:       record.Key,
					Headers:   headers,
					Timestamp: record.Timestamp,
				}); err != nil {
					c.logger.Errorf(err, "error in consumer ReaderFunc")
					return
				}

				if err := c.client.CommitRecords(record.Context, record); err != nil {
					c.logger.Errorf(err, "error while committing records")
					return
				}
			},
		)
	}
}

func (c *consumer) StopConsuming() {
	c.cancelContext()
}

func (c *consumer) LifecycleOnStart(ctx context.Context) error {
	c.logger.Debugf("consumer is about to ping kafka brokers")
	if err := c.Ping(ctx); err != nil {
		return err
	}
	c.logger.Infof("consumer is connected to kafka brokers")
	return nil
}

func (c *consumer) LifecycleOnStop(context.Context) error {
	c.cancelContext()
	c.logger.Debugf("consumer is about to be closed")
	if err := c.Close(); err != nil {
		return err
	}
	c.logger.Infof("consumer is closed")
	return nil
}

type ConsumerOpts struct {
	Logger         logging.Logger
	MaxRetries     *int
	MaxPollRecords *int
}

func NewConsumer(conn Conn, consumerGroup string, topics []string, opts ConsumerOpts) (Consumer, error) {
	if opts.MaxRetries == nil {
		opts.MaxRetries = fn.New(3)
	}

	if opts.MaxPollRecords == nil {
		opts.MaxPollRecords = fn.New(1)
	}

	if opts.Logger == nil {
		var err error
		opts.Logger, err = logging.New(&logging.Options{Name: "kafka-consumer"})
		if err != nil {
			return nil, err
		}
	}

	kgoOpts := []kgo.Opt{
		kgo.SeedBrokers(strings.Split(conn.GetBrokers(), ",")...),
		kgo.ConsumerGroup(consumerGroup),
		kgo.ConsumeTopics(topics...),
		kgo.DisableAutoCommit(),
	}

	saslOpt, err := parseSASLAuth(conn.GetSASLAuth())
	if err != nil {
		return nil, err
	}

	if saslOpt != nil {
		kgoOpts = append(kgoOpts, saslOpt)
	}

	client, err := kgo.NewClient(kgoOpts...)
	if err != nil {
		return nil, errors.NewEf(err, "unable to create client")
	}

	ctx, cancelFunc := context.WithCancel(context.TODO())
	return &consumer{
		ctx:           ctx,
		cancelContext: cancelFunc,
		client:        client,
		logger:        opts.Logger,
	}, nil
}
