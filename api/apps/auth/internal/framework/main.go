package framework

import (
	"go.uber.org/fx"

	"kloudlite.io/apps/auth/internal/app"
	"kloudlite.io/apps/auth/internal/env"
	"kloudlite.io/pkg/cache"
	rpc "kloudlite.io/pkg/grpc"
	httpServer "kloudlite.io/pkg/http-server"
	"kloudlite.io/pkg/repos"
)

type CommsGrpcEnv struct {
	*env.Env
}

func (c CommsGrpcEnv) GetGRPCServerURL() string {
	return c.CommsService
}

type fm struct {
	*env.Env
}

func (e *fm) GetHttpPort() uint16 {
	return e.Port
}

func (e *fm) GetHttpCors() string {
	return e.CorsOrigins
}

func (e *fm) RedisOptions() (hosts, username, password, basePrefix string) {
	return e.RedisHosts, e.RedisUserName, e.RedisPassword, e.RedisPrefix
}

func (e *fm) GetMongoConfig() (url string, dbName string) {
	return e.MongoUri, e.MongoDbName
}

func (e *fm) GetGRPCPort() uint16 {
	return e.GrpcPort
}

var Module fx.Option = fx.Module(
	"framework",
	fx.Provide(func(ev *env.Env) *fm {
		return &fm{ev}
	}),
	fx.Provide(func(ev *env.Env) *CommsGrpcEnv {
		return &CommsGrpcEnv{ev}
	}),
	repos.NewMongoClientFx[*fm](),
	cache.NewRedisFx[*fm](),
	httpServer.NewHttpServerFx[*fm](),
	rpc.NewGrpcServerFx[*fm](),
	rpc.NewGrpcClientFx[*CommsGrpcEnv, app.CommsClientConnection](),
	app.Module,
)
