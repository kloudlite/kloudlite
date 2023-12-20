package env

import (
	"github.com/codingconcepts/env"
)

type Env struct {
	HttpPort uint16 `env:"HTTP_PORT" required:"true"`
	HttpCors string `env:"CORS_ORIGINS" required:"false"`
	GrpcPort uint16 `env:"GRPC_PORT" required:"true"`

	DBName string `env:"MONGO_DB_NAME" required:"true"`
	DBUrl  string `env:"MONGO_URI" required:"true"`

	CookieDomain string `env:"COOKIE_DOMAIN" required:"true"`

	IamGrpcAddr               string `env:"IAM_GRPC_ADDR" required:"true"`
	CommsGrpcAddr             string `env:"COMMS_GRPC_ADDR" required:"true"`
	ContainerRegistryGrpcAddr string `env:"CONTAINER_REGISTRY_GRPC_ADDR" required:"true"`
	ConsoleGrpcAddr           string `env:"CONSOLE_GRPC_ADDR" required:"true"`
	AuthGrpcAddr              string `env:"AUTH_GRPC_ADDR" required:"true"`
	SessionKVBucket           string `env:"SESSION_KV_BUCKET" required:"true"`
	NatsURL                   string `env:"NATS_URL" required:"true"`
	IsDev                     bool
}

func LoadEnv() (*Env, error) {
	var ev Env
	if err := env.Set(&ev); err != nil {
		return nil, err
	}
	return &ev, nil
}
