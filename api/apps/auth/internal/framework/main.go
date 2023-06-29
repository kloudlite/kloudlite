package framework

import (
	"go.uber.org/fx"
	"kloudlite.io/apps/auth/internal/app"
	"kloudlite.io/pkg/cache"
	"kloudlite.io/pkg/config"
	rpc "kloudlite.io/pkg/grpc"
	httpServer "kloudlite.io/pkg/http-server"
	"kloudlite.io/pkg/repos"
)

type CommsGrpcEnv struct {
	CommsService string `env:"COMMS_SERVICE" required:"true"`
}

func (c CommsGrpcEnv) GetGRPCServerURL() string {
	return c.CommsService
}

type Env struct {
	MongoUri      string `env:"MONGO_URI" required:"true"`
	RedisHosts    string `env:"REDIS_HOSTS" required:"true"`
	RedisUserName string `env:"REDIS_USERNAME" required:"true"`
	RedisPassword string `env:"REDIS_PASSWORD" required:"true"`
	RedisPrefix   string `env:"REDIS_PREFIX" required:"true"`
	MongoDbName   string `env:"MONGO_DB_NAME" required:"true"`
	Port          uint16 `env:"PORT" required:"true"`
	GrpcPort      uint16 `env:"GRPC_PORT" required:"true"`
	CorsOrigins   string `env:"ORIGINS"`
}

func (e *Env) GetHttpPort() uint16 {
	return e.Port
}

func (e *Env) GetHttpCors() string {
	return e.CorsOrigins
}

func (e *Env) RedisOptions() (hosts, username, password, basePrefix string) {
	return e.RedisHosts, e.RedisUserName, e.RedisPassword, e.RedisPrefix
}

func (e *Env) GetMongoConfig() (url string, dbName string) {
	return e.MongoUri, e.MongoDbName
}

func (e *Env) GetGRPCPort() uint16 {
	return e.GrpcPort
}

var Module fx.Option = fx.Module(
	"framework",
	config.EnvFx[Env](),
	config.EnvFx[CommsGrpcEnv](),
	repos.NewMongoClientFx[*Env](),
	cache.NewRedisFx[*Env](),
	httpServer.NewHttpServerFx[*Env](),
	rpc.NewGrpcServerFx[*Env](),
	rpc.NewGrpcClientFx[*CommsGrpcEnv, app.CommsClientConnection](),
	app.Module,
)
