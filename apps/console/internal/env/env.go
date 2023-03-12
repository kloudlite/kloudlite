package env

import "github.com/codingconcepts/env"

type Env struct {
	Port         uint16 `env:"PORT" required:"true"`
	CookieDomain string `env:"COOKIE_DOMAIN" required:"true"`

	ConsoleDBUri  string `env:"CONSOLE_DB_URI" required:"true"`
	ConsoleDBName string `env:"CONSOLE_DB_NAME" required:"true"`

	AuthRedisHosts    string `env:"AUTH_REDIS_HOSTS" required:"true"`
	AuthRedisUserName string `env:"AUTH_REDIS_USERNAME" required:"true"`
	AuthRedisPassword string `env:"AUTH_REDIS_PASSWORD" required:"true"`
	AuthRedisPrefix   string `env:"AUTH_REDIS_PREFIX" required:"true"`
}

func LoadEnv() (*Env, error) {
	var e Env
	if err := env.Set(&e); err != nil {
		return nil, err
	}
	return &e, nil
}
