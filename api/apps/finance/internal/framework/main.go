package framework

import (
	"fmt"
	"go.uber.org/fx"
	"kloudlite.io/apps/finance/internal/app"
	"kloudlite.io/pkg/cache"
	"kloudlite.io/pkg/config"
	rpc "kloudlite.io/pkg/grpc"
	httpServer "kloudlite.io/pkg/http-server"
	"kloudlite.io/pkg/logger"
	"kloudlite.io/pkg/repos"
)

type AuthGRPCEnv struct {
	AuthServerHost string `env:"AUTH_SERVER_HOST"`
	AuthServerPort uint16 `env:"AUTH_SERVER_PORT"`
}

func (e *AuthGRPCEnv) GetGCPServerURL() string {
	return fmt.Sprintf("%v:%v", e.AuthServerHost, e.AuthServerPort)
}

type ConsoleGRPCEnv struct {
	ConsoleGrpcService string `env:"CONSOLE_GRPC_SERVICE"`
}

func (e *ConsoleGRPCEnv) GetGCPServerURL() string {
	return e.ConsoleGrpcService
}

type CiGrpcEnv struct {
	CiService string `env:"CI_GRPC_SERVICE" required:"true"`
}

func (e *CiGrpcEnv) GetGCPServerURL() string {
	return e.CiService
}

type IAMGRPCEnv struct {
	IAMService string `env:"IAM_GRPC_SERVICE"`
}

func (e *IAMGRPCEnv) GetGCPServerURL() string {
	return e.IAMService
}

type Env struct {
	DBName        string `env:"MONGO_DB_NAME"`
	DBUrl         string `env:"MONGO_URI"`
	RedisHosts    string `env:"REDIS_HOSTS"`
	RedisUsername string `env:"REDIS_USERNAME"`
	RedisPassword string `env:"REDIS_PASSWORD"`

	AuthRedisHosts    string `env:"REDIS_AUTH_HOSTS"`
	AuthRedisUserName string `env:"REDIS_AUTH_USERNAME"`
	AuthRedisPassword string `env:"REDIS_AUTH_PASSWORD"`
	AuthRedisPrefix   string `env:"REDIS_AUTH_PREFIX"`

	HttpPort uint16 `env:"PORT"`
	HttpCors string `env:"ORIGINS"`
}

func (e *Env) GetMongoConfig() (url string, dbName string) {
	return e.DBUrl, e.DBName
}

func (e *Env) RedisOptions() (hosts, username, password, basePrefix string) {
	return e.RedisHosts, e.RedisUsername, e.RedisPassword, basePrefix
}

func (e *Env) GetHttpPort() uint16 {
	return e.HttpPort
}

func (e *Env) GetHttpCors() string {
	return e.HttpCors
}

var Module = fx.Module(
	"framework",
	logger.FxProvider(),
	config.EnvFx[Env](),
	config.EnvFx[ConsoleGRPCEnv](),
	config.EnvFx[IAMGRPCEnv](),
	config.EnvFx[CiGrpcEnv](),
	config.EnvFx[AuthGRPCEnv](),
	rpc.NewGrpcClientFx[*ConsoleGRPCEnv, app.ConsoleClientConnection](),
	rpc.NewGrpcClientFx[*IAMGRPCEnv, app.IAMClientConnection](),
	rpc.NewGrpcClientFx[*CiGrpcEnv, app.CIGrpcClientConn](),
	rpc.NewGrpcClientFx[*AuthGRPCEnv, app.AuthGrpcClientConn](),
	repos.NewMongoClientFx[*Env](),
	fx.Provide(
		func(env *Env) app.AuthCacheClient {
			return cache.NewRedisClient(
				env.AuthRedisHosts,
				env.AuthRedisUserName,
				env.AuthRedisPassword,
				env.AuthRedisPrefix,
			)
		},
	),
	cache.FxLifeCycle[app.AuthCacheClient](),

	fx.Provide(
		func(env *Env) cache.Client {
			return cache.NewRedisClient(env.RedisOptions())
		},
	),
	cache.FxLifeCycle[cache.Client](),
	httpServer.NewHttpServerFx[*Env](),
	app.Module,
)
