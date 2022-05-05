package framework

import (
	"go.uber.org/fx"
	"kloudlite.io/apps/auth/internal/app"
	"kloudlite.io/pkg/cache"
	"kloudlite.io/pkg/config"
	rpc "kloudlite.io/pkg/grpc"
	httpServer "kloudlite.io/pkg/http-server"
	"kloudlite.io/pkg/logger"
	"kloudlite.io/pkg/repos"
)

type CommsGrpcEnv struct {
	CommsHost string `env:"COMMS_HOST" required:"true"`
	CommsPort string `env:"COMMS_PORT" required:"true"`
}

func (c CommsGrpcEnv) GetGCPServerURL() string {
	return c.CommsHost + ":" + c.CommsPort
}

type Env struct {
	MongoUri      string `env:"MONGO_URI" required:"true"`
	RedisHosts    string `env:"REDIS_HOSTS" required:"true"`
	RedisUserName string `env:"REDIS_USERNAME"`
	RedisPassword string `env:"REDIS_PASSWORD"`
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

func (e *Env) RedisOptions() (hosts, username, password string) {
	return e.RedisHosts, e.RedisUserName, e.RedisPassword
}

func (e *Env) GetMongoConfig() (url string, dbName string) {
	return e.MongoUri, e.MongoDbName
}

func (e *Env) GetGRPCPort() uint16 {
	return e.GrpcPort
}

var Module = fx.Module("framework",
	config.EnvFx[Env](),
	config.EnvFx[CommsGrpcEnv](),
	fx.Provide(logger.NewLogger),
	repos.NewMongoClientFx[*Env](),
	cache.NewRedisFx[*Env](),
	httpServer.NewHttpServerFx[*Env](),
	rpc.NewGrpcServerFx[*Env](),
	rpc.NewGrpcClientFx[*CommsGrpcEnv, app.CommsClientConnection](),
	app.Module,
)
