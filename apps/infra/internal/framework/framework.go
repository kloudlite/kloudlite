package framework

import (
	"go.uber.org/fx"
	"kloudlite.io/apps/infra/internal/app"
	"kloudlite.io/apps/infra/internal/env"
	"kloudlite.io/pkg/cache"
	rpc "kloudlite.io/pkg/grpc"
	httpServer "kloudlite.io/pkg/http-server"
	mongoRepo "kloudlite.io/pkg/repos"
)

type fEnv struct {
	*env.Env
}


func (f *fEnv) GetGRPCServerURL() string {
	return f.FinanceGrpcAddr
}

func (f *fEnv) GetHttpCors() string {
	return "https://studio.apollographql.com"
}

func (f *fEnv) GetHttpPort() uint16 {
	return f.HttpPort
}

func (f *fEnv) GetMongoConfig() (url string, dbName string) {
	return f.InfraDbUri, f.InfraDbName
}

var Module = fx.Module("framework",
	fx.Provide(func(ev *env.Env) *fEnv {
		return &fEnv{Env: ev}
	}),

	mongoRepo.NewMongoClientFx[*fEnv](),
	httpServer.NewHttpServerFx[*fEnv](),

	fx.Provide(
		func(f *fEnv) app.AuthCacheClient {
			return cache.NewRedisClient(f.AuthRedisHosts, f.AuthRedisUserName, f.AuthRedisPassword, f.AuthRedisPrefix)
		},
	),
	rpc.NewGrpcClientFx[*fEnv, app.FinanceClientConnection](),
	cache.FxLifeCycle[app.AuthCacheClient](),
	app.Module,
)
