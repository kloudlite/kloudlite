package framework

import (
	"go.uber.org/fx"
	"kloudlite.io/apps/finance_deprecated/internal/app"
	"kloudlite.io/apps/finance_deprecated/internal/env"
	"kloudlite.io/pkg/cache"
	"kloudlite.io/pkg/config"
	rpc "kloudlite.io/pkg/grpc"
	httpServer "kloudlite.io/pkg/http-server"
	"kloudlite.io/pkg/repos"
)

type AuthGRPCEnv struct {
	AuthService string `env:"AUTH_SERVICE"`
}

func (e *AuthGRPCEnv) GetGRPCServerURL() string {
	return e.AuthService
}

type CommsGRPCEnv struct {
	CommsGrpcService string `env:"COMMS_SERVICE"`
}

func (e *CommsGRPCEnv) GetGRPCServerURL() string {
	return e.CommsGrpcService
}

type ContainerRegistryGRPCEnv struct {
	ContainerRegistryGrpcService string `env:"CONTAINER_REGISTRY_SERVICE"`
}

func (e *ContainerRegistryGRPCEnv) GetGRPCServerURL() string {
	return e.ContainerRegistryGrpcService
}

type ConsoleGRPCEnv struct {
	ConsoleGrpcService string `env:"CONSOLE_SERVICE"`
}

func (e *ConsoleGRPCEnv) GetGRPCServerURL() string {
	return e.ConsoleGrpcService
}

type IAMGRPCEnv struct {
	IAMService string `env:"IAM_SERVICE"`
}

func (e *IAMGRPCEnv) GetGRPCServerURL() string {
	return e.IAMService
}

type fm struct {
	*env.Env
}

func (f *fm) GetGRPCPort() uint16 {
	return f.GrpcPort
}

func (f *fm) GetMongoConfig() (url string, dbName string) {
	return f.DBUrl, f.DBName
}

func (f *fm) RedisOptions() (hosts, username, password, basePrefix string) {
	return f.RedisHosts, f.RedisUsername, f.RedisPassword, f.RedisPrefix
}

func (f *fm) GetHttpPort() uint16 {
	return f.HttpPort
}

func (f *fm) GetHttpCors() string {
	return f.HttpCors
}

var Module fx.Option = fx.Module(
	"framework",
	fx.Provide(func(ev *env.Env) *fm {
		return &fm{Env: ev}
	}),

	// config.EnvFx[Env](),
	config.EnvFx[ConsoleGRPCEnv](),
	config.EnvFx[CommsGRPCEnv](),
	config.EnvFx[IAMGRPCEnv](),
	config.EnvFx[AuthGRPCEnv](),
	config.EnvFx[ContainerRegistryGRPCEnv](),
	rpc.NewGrpcServerFx[*fm](),
	rpc.NewGrpcClientFx[*ConsoleGRPCEnv, app.ConsoleClientConnection](),
	rpc.NewGrpcClientFx[*ContainerRegistryGRPCEnv, app.ContainerRegistryClientConnection](),
	rpc.NewGrpcClientFx[*CommsGRPCEnv, app.CommsClientConnection](),
	rpc.NewGrpcClientFx[*IAMGRPCEnv, app.IAMClientConnection](),
	rpc.NewGrpcClientFx[*AuthGRPCEnv, app.AuthGrpcClientConn](),
	repos.NewMongoClientFx[*fm](),
	fx.Provide(
		func(env *fm) app.AuthCacheClient {
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
		func(f *fm) cache.Client {
			return cache.NewRedisClient(f.RedisOptions())
		},
	),

	cache.FxLifeCycle[cache.Client](),
	httpServer.NewHttpServerFx[*fm](),
	app.Module,
)
