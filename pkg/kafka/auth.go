package kafka

import (
	"context"
	"github.com/kloudlite/api/pkg/errors"
	"github.com/twmb/franz-go/pkg/kgo"
	"github.com/twmb/franz-go/pkg/sasl/scram"
)

type SASLMechanism string

const (
	ScramSHA256 SASLMechanism = "SCRAM-SHA-256"
	ScramSHA512 SASLMechanism = "SCRAM-SHA-512"
)

type SASLAuth struct {
	SASLMechanism SASLMechanism
	User          string
	Password      string
}

func parseSASLAuth(sasl *SASLAuth) (kgo.Opt, error) {
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
	return nil, errors.Newf("unknown SASL mechanism: %s", sasl.SASLMechanism)
}
