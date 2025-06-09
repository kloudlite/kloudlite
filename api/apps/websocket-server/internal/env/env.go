package env

import (
	"github.com/codingconcepts/env"
	"github.com/kloudlite/api/pkg/errors"
)

type Env struct {
	SocketPort uint16 `env:"SOCKET_PORT" required:"true"`

	IamGrpcAddr  string `env:"IAM_GRPC_ADDR" required:"true"`
	CookieDomain string `env:"COOKIE_DOMAIN" required:"true"`

	SessionKVBucket string `env:"SESSION_KV_BUCKET" required:"true"`
	NatsURL         string `env:"NATS_URL" required:"true"`
	IsDev           bool

	LogsStreamName string `env:"LOGS_STREAM_NAME" default:"logs"`

	ObservabilityApiAddr string `env:"OBSERVABILITY_API_ADDR" required:"true"`
}

func LoadEnv() (*Env, error) {
	var ev Env
	if err := env.Set(&ev); err != nil {
		return nil, errors.NewE(err)
	}
	return &ev, nil
}
