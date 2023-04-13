package redpanda

import (
	"context"
	"github.com/twmb/franz-go/pkg/kgo"
	"github.com/twmb/franz-go/pkg/sasl/scram"
	"kloudlite.io/pkg/errors"
	fn "kloudlite.io/pkg/functions"
	"kloudlite.io/pkg/logging"
)

type SASLMechanism string

const (
	ScramSHA256 SASLMechanism = "SCRAM-SHA-256"
	ScramSHA512 SASLMechanism = "SCRAM-SHA-512"
)

type KafkaSASLAuth struct {
	SASLMechanism SASLMechanism
	User          string
	Password      string
}

type ProducerOpts struct {
	SASLAuth *KafkaSASLAuth
	Logger   logging.Logger
}

type ConsumerOpts struct {
	SASLAuth       *KafkaSASLAuth
	Logger         logging.Logger
	MaxRetries     *int
	MaxPollRecords *int
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
