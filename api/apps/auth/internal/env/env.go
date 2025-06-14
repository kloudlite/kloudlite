package env

import (
	"github.com/codingconcepts/env"
	"github.com/kloudlite/api/pkg/errors"
)

type authEnv struct {
	MongoUri    string `env:"MONGO_URI" required:"true"`
	MongoDbName string `env:"MONGO_DB_NAME" required:"true"`
	GrpcPort    uint16 `env:"GRPC_PORT" required:"true"`
	GrpcV2Port  uint16 `env:"GRPC_V2_PORT" required:"true"`

	UserEmailVerifactionEnabled bool `env:"USER_EMAIL_VERIFICATION_ENABLED" default:"true"`

	CommsService               string `env:"COMMS_SERVICE" required:"true"`
	NatsURL                    string `env:"NATS_URL" required:"true"`
	SessionKVBucket            string `env:"SESSION_KV_BUCKET" required:"true"`
	VerifyTokenKVBucket        string `env:"VERIFY_TOKEN_KV_BUCKET" required:"true"`
	ResetPasswordTokenKVBucket string `env:"RESET_PASSWORD_TOKEN_KV_BUCKET" required:"true"`

	IsDev bool
}

type Env struct {
	authEnv
}

func LoadEnv() (*Env, error) {
	var ev Env

	if err := env.Set(&ev.authEnv); err != nil {
		return nil, errors.NewE(err)
	}

	return &ev, nil
}
