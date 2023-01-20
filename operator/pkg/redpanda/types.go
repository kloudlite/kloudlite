package redpanda

import (
	"context"
	"time"

	"github.com/twmb/franz-go/pkg/kgo"
	"github.com/twmb/franz-go/pkg/sasl/scram"
	"github.com/kloudlite/operator/pkg/errors"
	fn "github.com/kloudlite/operator/pkg/functions"
	"github.com/kloudlite/operator/pkg/logging"
)

type Consumer interface {
	Ping(ctx context.Context) error
	Close()
	StartConsuming(reader ReaderFunc)
}

type KafkaSASLAuth struct {
	SASLMechanism SASLMechanism
	User          string
	Password      string
}

func parseSASLAuth(sasl *KafkaSASLAuth) (kgo.Opt, error) {
	if sasl == nil {
		return nil, nil
	}
	if sasl.SASLMechanism == ScramSHA256 {
		return kgo.SASL(
			scram.Sha256(
				func(context.Context) (scram.Auth, error) {
					return scram.Auth{
						User: sasl.User,
						Pass: sasl.Password,
					}, nil
				},
			),
		), nil
	}

	if sasl.SASLMechanism == ScramSHA512 {
		return kgo.SASL(
			scram.Sha512(
				func(ctx context.Context) (scram.Auth, error) {
					return scram.Auth{
						User: sasl.User,
						Pass: sasl.Password,
					}, nil
				},
			),
		), nil
	}
	return nil, errors.Newf("unknown SASL mechanism")
}

type ConsumerOpts struct {
	SASLAuth       *KafkaSASLAuth
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

type SASLMechanism string

const (
	ScramSHA256 SASLMechanism = "SCRAM-SHA-256"
	ScramSHA512 SASLMechanism = "SCRAM-SHA-512"
)

type ProducerOutput struct {
	Key        []byte    `json:"key,omitempty"`
	Timestamp  time.Time `json:"timestamp"`
	Topic      string    `json:"topic"`
	Partition  int32     `json:"partition,omitempty"`
	ProducerId int64     `json:"producerId,omitempty"`
	Offset     int64     `json:"offset"`
}

type ProducerOpts struct {
	SASLAuth *KafkaSASLAuth
	Logger   logging.Logger
}
