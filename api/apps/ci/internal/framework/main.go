package framework

import (
	"go.uber.org/fx"
	"kloudlite.io/apps/ci/internal/app"
	"kloudlite.io/pkg/cache"
	"kloudlite.io/pkg/config"
	rpc "kloudlite.io/pkg/grpc"
	httpServer "kloudlite.io/pkg/http-server"
	"kloudlite.io/pkg/logger"
	"kloudlite.io/pkg/repos"
)

type Env struct {
	DBName        string `env:"MONGO_DB_NAME" required:"true"`
	DBUrl         string `env:"MONGO_URI" required:"true"`
	RedisHost     string `env:"REDIS_HOSTS" required:"true"`
	RedisUserName string `env:"REDIS_USERNAME"`
	RedisPassword string `env:"REDIS_PASSWORD"`
	HttpPort      uint16 `env:"PORT" required:"true"`
	HttpCors      string `env:"ORIGINS" required:"true"`
	GrpcPort      uint16 `env:"GRPC_PORT" required:"true"`
}

type GrpcAuthConfig struct {
	InfraGrpcHost string `env:"AUTH_HOST" required:"true"`
	InfraGrpcPort string `env:"AUTH_PORT" required:"true"`
}

func (e *GrpcAuthConfig) GetGCPServerURL() string {
	return e.InfraGrpcHost + ":" + e.InfraGrpcPort
}

func (e *Env) GetGRPCPort() uint16 {
	return e.GrpcPort
}

func (e *Env) RedisOptions() (hosts, username, password string) {
	return e.RedisHost, e.RedisUserName, e.RedisPassword
}

func (e *Env) GetMongoConfig() (url string, dbName string) {
	return e.DBUrl, e.DBName
}

func (e *Env) GetHttpPort() uint16 {
	return e.HttpPort
}

func (e *Env) GetHttpCors() string {
	return e.HttpCors
}

var Module = fx.Module("framework",
	fx.Provide(logger.NewLogger),
	config.EnvFx[Env](),
	config.EnvFx[GrpcAuthConfig](),
	cache.NewRedisFx[*Env](),
	httpServer.NewHttpServerFx[*Env](),
	rpc.NewGrpcServerFx[*Env](),
	repos.NewMongoClientFx[*Env](),
	rpc.NewGrpcClientFx[*GrpcAuthConfig, app.AuthClientConnection](),
	app.Module,
)
