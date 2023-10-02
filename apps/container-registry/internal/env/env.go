package env

import "github.com/codingconcepts/env"

type Env struct {
	// new
	RegistryUrl              string `env:"REGISTRY_URL" required:"true"`
	RegistrySecretKey        string `env:"REGISTRY_SECRET_KEY" required:"true"`
	RegistryAuthorizerPort   uint16 `env:"REGISTRY_AUTHORIZER_PORT" required:"true"`

	// old
	Port              uint16 `env:"PORT" required:"true"`
	CookieDomain      string `env:"COOKIE_DOMAIN" required:"true"`
	AccountCookieName string `env:"ACCOUNT_COOKIE_NAME" required:"true"`

	AuthRedisPrefix   string `env:"AUTH_REDIS_PREFIX" required:"true"`
	AuthRedisHosts    string `env:"AUTH_REDIS_HOSTS" required:"true"`
	AuthRedisUserName string `env:"AUTH_REDIS_USERNAME" required:"true"`
	AuthRedisPassword string `env:"AUTH_REDIS_PASSWORD" required:"true"`

	CRRedisPrefix   string `env:"REGISTRY_REDIS_PREFIX" required:"true"`
	CRRedisHosts    string `env:"REGISTRY_REDIS_HOSTS" required:"true"`
	CRRedisUserName string `env:"REGISTRY_REDIS_USERNAME" required:"true"`
	CRRedisPassword string `env:"REGISTRY_REDIS_PASSWORD" required:"true"`

	DBUri       string `env:"DB_URI" required:"true"`
	DBName      string `env:"DB_NAME" required:"true"`
	IAMGrpcAddr string `env:"IAM_GRPC_ADDR" required:"true"`
}

func LoadEnv() (*Env, error) {
	var e Env
	if err := env.Set(&e); err != nil {
		return nil, err
	}
	return &e, nil
}
