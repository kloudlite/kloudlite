package env

import "github.com/codingconcepts/env"

type Env struct {
	Port                uint16 `env:"PORT" required:"true"`
	CookieDomain        string `env:"COOKIE_DOMAIN" required:"true"`
	AccountCookieName   string `env:"ACCOUNT_COOKIE_NAME" required:"true"`
	HarborAdminPassword string `env:"HARBOR_ADMIN_PASSWORD" required:"true"`
	HarborRegistryHost  string `env:"HARBOR_REGISTRY_HOST" required:"true"`
	HarborAdminUsername string `env:"HARBOR_ADMIN_USERNAME" required:"true"`
	AuthRedisPrefix     string `env:"AUTH_REDIS_PREFIX" required:"true"`
	AuthRedisHosts      string `env:"AUTH_REDIS_HOSTS" required:"true"`
	AuthRedisUserName   string `env:"AUTH_REDIS_USERNAME" required:"true"`
	AuthRedisPassword   string `env:"AUTH_REDIS_PASSWORD" required:"true"`
	DBUri               string `env:"DB_URI" required:"true"`
	DBName              string `env:"DB_NAME" required:"true"`
	GRPCPort            uint16 `env:"GRPC_PORT" required:"true"`
}

func LoadEnv() (*Env, error) {
	var e Env
	if err := env.Set(&e); err != nil {
		return nil, err
	}
	return &e, nil
}
