package framework

import (
	"go.uber.org/fx"
	"kloudlite.io/apps/iam/internal/application"
	"kloudlite.io/pkg/cache"
	"kloudlite.io/pkg/config"
	rpc "kloudlite.io/pkg/grpc"
	"kloudlite.io/pkg/repos"
)

type Env struct {
	Port        uint16 `env:"GRPC_PORT" required:"true"`
	MongoDbUri  string `env:"MONGO_DB_URI" required:"true"`
	MongoDbName string `env:"MONGO_DB_NAME" required:"true"`
	RedisHosts  string `env:"REDIS_HOSTS" required:"true"`
}

func (env *Env) RedisOptions() (hosts string, username string, password string) {
	return env.RedisHosts, "", ""
}

func (env *Env) GetMongoConfig() (url, dbName string) {
	return env.MongoDbUri, env.MongoDbName
}

func (env *Env) GetGRPCPort() uint16 {
	return env.Port
}

var Module = fx.Module("framework",
	config.EnvFx[*Env](),
	repos.NewMongoClientFx[*Env](),
	cache.NewRedisFx[*Env](),
	rpc.NewGrpcServerFx[*Env](),
	application.Module,
)
