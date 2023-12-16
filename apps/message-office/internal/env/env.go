package env

import (
	"github.com/codingconcepts/env"
)

type Env struct {
	NatsUrl    string `env:"NATS_URL" required:"true"`
	NatsStream string `env:"NATS_STREAM" required:"true"`

	PlatformAccessToken string `env:"PLATFORM_ACCESS_TOKEN" required:"true"`

	DbName string `env:"DB_NAME" required:"true"`
	DbUri  string `env:"DB_URI"  required:"true"`

	ExternalGrpcPort uint16 `env:"EXTERNAL_GRPC_PORT" required:"true"`
	InternalGrpcPort uint16 `env:"INTERNAL_GRPC_PORT" required:"true"`
	HttpPort         uint16 `env:"HTTP_PORT" required:"true"`

	VectorGrpcAddr string `env:"VECTOR_GRPC_ADDR" required:"true"`

	TokenHashingSecret string `env:"TOKEN_HASHING_SECRET" required:"true"`
}

func LoadEnvOrDie() *Env {
	var ev Env
	if err := env.Set(&ev); err != nil {
		panic(err)
	}
	return &ev
}
