package framework

import (
	"go.uber.org/fx"
	"kloudlite.io/apps/ci/internal/app"
	"kloudlite.io/pkg/cache"
	"kloudlite.io/pkg/config"
	rpc "kloudlite.io/pkg/grpc"
	httpServer "kloudlite.io/pkg/http-server"
	"kloudlite.io/pkg/redpanda"
	"kloudlite.io/pkg/repos"
)

type GrpcAuthConfig struct {
	Addr string `env:"AUTH_ADDR" required:"true"`
}

func (e *GrpcAuthConfig) GetGRPCServerURL() string {
	return e.Addr
}

type GrpcConsoleConfig struct {
	Addr string `env:"CONSOLE_ADDR" required:"true"`
}

func (e *GrpcConsoleConfig) GetGRPCServerURL() string {
	return e.Addr
}

type Env struct {
	DBName        string `env:"MONGO_DB_NAME" required:"true"`
	DBUrl         string `env:"MONGO_URI" required:"true"`
	RedisHost     string `env:"REDIS_HOSTS" required:"true"`
	RedisUserName string `env:"REDIS_USERNAME" required:"true"`
	RedisPassword string `env:"REDIS_PASSWORD" required:"true"`
	RedisPrefix   string `env:"REDIS_PREFIX" required:"true"`

	HttpPort uint16 `env:"PORT" required:"true"`
	HttpCors string `env:"ORIGINS" required:"true"`
	GrpcPort uint16 `env:"GRPC_PORT" required:"true"`

	AuthRedisHost     string `env:"AUTH_REDIS_HOSTS" required:"true"`
	AuthRedisUsername string `env:"AUTH_REDIS_USERNAME" required:"true"`
	AuthRedisPassword string `env:"AUTH_REDIS_PASSWORD" required:"true"`
	AuthRedisPrefix   string `env:"AUTH_REDIS_PREFIX" required:"true"`

	KafkaBrokers string `env:"KAFKA_BROKERS" required:"true"`
}

func (e *Env) GetBrokers() string {
	return e.KafkaBrokers
}

func (e *Env) GetGRPCPort() uint16 {
	return e.GrpcPort
}

func (e *Env) RedisOptions() (hosts, username, password, basePrefix string) {
	return e.RedisHost, e.RedisUserName, e.RedisPassword, basePrefix
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

var Module fx.Option = fx.Module(
	"framework",
	config.EnvFx[Env](),
	config.EnvFx[GrpcAuthConfig](),
	config.EnvFx[GrpcConsoleConfig](),
	fx.Provide(
		func(env *Env) app.AuthCacheClient {
			return cache.NewRedisClient(env.AuthRedisHost, env.AuthRedisUsername, env.AuthRedisPassword, env.AuthRedisPrefix)
		},
	),
	cache.FxLifeCycle[app.AuthCacheClient](),

	fx.Provide(
		func(env *Env) app.CacheClient {
			return cache.NewRedisClient(env.RedisOptions())
		},
	),
	cache.FxLifeCycle[app.CacheClient](),

	httpServer.NewHttpServerFx[*Env](),
	rpc.NewGrpcServerFx[*Env](),
	repos.NewMongoClientFx[*Env](),
	rpc.NewGrpcClientFx[*GrpcAuthConfig, app.AuthGRPCClient](),
	rpc.NewGrpcClientFx[*GrpcConsoleConfig, app.ConsoleGRPCClient](),
	redpanda.NewClientFx[*Env](),
	app.Module,
)
