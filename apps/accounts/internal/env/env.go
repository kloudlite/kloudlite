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

	// RedisHosts    string `env:"REDIS_HOSTS" required:"true"`
	// RedisUsername string `env:"REDIS_USERNAME" required:"true"`
	// RedisPassword string `env:"REDIS_PASSWORD" required:"true"`
	// RedisPrefix   string `env:"REDIS_PREFIX" required:"true"`

	AuthRedisHosts    string `env:"REDIS_AUTH_HOSTS" required:"true"`
	AuthRedisUserName string `env:"REDIS_AUTH_USERNAME" required:"true"`
	AuthRedisPassword string `env:"REDIS_AUTH_PASSWORD" required:"true"`
	AuthRedisPrefix   string `env:"REDIS_AUTH_PREFIX" required:"true"`

	CookieDomain string `env:"COOKIE_DOMAIN" required:"true"`
	// StripePublicKey string `env:"STRIPE_PUBLIC_KEY" required:"true"`
	// StripeSecretKey string `env:"STRIPE_SECRET_KEY" required:"true"`

	IamGrpcAddr               string `env:"IAM_GRPC_ADDR" required:"true"`
	CommsGrpcAddr             string `env:"COMMS_GRPC_ADDR" required:"true"`
	ContainerRegistryGrpcAddr string `env:"CONTAINER_REGISTRY_GRPC_ADDR" required:"true"`
	ConsoleGrpcAddr           string `env:"CONSOLE_GRPC_ADDR" required:"true"`
	AuthGrpcAddr              string `env:"AUTH_GRPC_ADDR" required:"true"`
}

func LoadEnv() (*Env, error) {
	var ev Env
	if err := env.Set(&ev); err != nil {
		return nil, err
	}
	return &ev, nil
}
