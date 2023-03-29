package env

import "github.com/codingconcepts/env"

type Env struct {
	Port          uint16 `env:"GRPC_PORT" required:"true"`
	MongoDbUri    string `env:"MONGO_DB_URI" required:"true"`
	MongoDbName   string `env:"MONGO_DB_NAME" required:"true"`
	RedisHosts    string `env:"REDIS_HOSTS" required:"true"`
	RedisUsername string `env:"REDIS_USERNAME" required:"true"`
	RedisPassword string `env:"REDIS_PASSWORD" required:"true"`
	RedisPrefix   string `env:"REDIS_PREFIX" required:"true"`
}

func LoadEnv() (*Env, error) {
	var e Env
	if err := env.Set(&e); err != nil {
		return nil, err
	}
	return &e, nil
}
