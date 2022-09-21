package redpanda

import (
	"context"
	"time"

	fn "operators.kloudlite.io/lib/functions"
	"operators.kloudlite.io/lib/logging"
)

type Consumer interface {
	Ping(ctx context.Context) error
	Close()
	StartConsuming(reader ReaderFunc) error
}

type ConsumerOpts struct {
	Logger         logging.Logger
	MaxRetries     *int
	MaxPollRecords *int
}

func (c *ConsumerOpts) getWithDefaults() *ConsumerOpts {
	opts := func() *ConsumerOpts {
		if c == nil {
			return &ConsumerOpts{}
		}
		return c
	}()

	if opts.MaxRetries == nil {
		opts.MaxRetries = fn.New(3)
	}
	if opts.MaxPollRecords == nil {
		opts.MaxPollRecords = fn.New(1)
	}

	return opts
}

type ReaderFunc func(msg KafkaMessage) error

type KafkaMessage struct {
	Key        []byte
	Value      []byte
	Timestamp  time.Time
	Topic      string
	Partition  int32
	ProducerId int64
	Offset     int64
}
