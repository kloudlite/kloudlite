package framework

import (
	"go.uber.org/fx"
	app "kloudlite.io/apps/container-registry/internal/app"
	"kloudlite.io/apps/container-registry/internal/env"
	"kloudlite.io/pkg/cache"
	rpc "kloudlite.io/pkg/grpc"
	httpServer "kloudlite.io/pkg/http-server"
	mongoDb "kloudlite.io/pkg/repos"
)

type fm struct {
	ev *env.Env
}

func (fm *fm) GetHttpPort() uint16 {
	return fm.ev.Port
}

func (fm *fm) GetHttpCors() string {
	return "*"
}

func (fm *fm) GetMongoConfig() (url string, dbName string) {
	return fm.ev.DBUri, fm.ev.DBName
}

func (fm *fm) GetGRPCPort() uint16 {
	return fm.ev.GRPCPort
}

var Module = fx.Module("framework",
	fx.Provide(func(ev *env.Env) *fm {
		return &fm{ev}
	}),

	mongoDb.NewMongoClientFx[*fm](),

	fx.Provide(func(ev *env.Env) app.AuthCacheClient {
		return cache.NewRedisClient(ev.AuthRedisHosts, ev.AuthRedisUserName, ev.AuthRedisPassword, ev.AuthRedisPrefix)
	}),
	cache.FxLifeCycle[app.AuthCacheClient](),

	rpc.NewGrpcServerFx[*fm](),

	app.Module,
	httpServer.NewHttpServerFx[*fm](),
)
