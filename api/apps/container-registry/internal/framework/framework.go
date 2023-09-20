package framework

import (
	"context"

	"go.uber.org/fx"
	app "kloudlite.io/apps/container-registry/internal/app"
	"kloudlite.io/apps/container-registry/internal/env"
	"kloudlite.io/pkg/cache"
	rpc "kloudlite.io/pkg/grpc"
	httpServer "kloudlite.io/pkg/http-server"
	"kloudlite.io/pkg/logging"
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

var Module = fx.Module("framework",
	fx.Provide(func(ev *env.Env) *fm {
		return &fm{ev}
	}),

	fx.Provide(func(ev *env.Env) (app.IAMGrpcClient, error) {
		return rpc.NewGrpcClient(ev.IAMGrpcAddr)
	}),

	mongoDb.NewMongoClientFx[*fm](),

	fx.Provide(func(ev *env.Env) app.AuthCacheClient {
		return cache.NewRedisClient(ev.AuthRedisHosts, ev.AuthRedisUserName, ev.AuthRedisPassword, ev.AuthRedisPrefix)
	}),

	cache.FxLifeCycle[app.AuthCacheClient](),

	app.Module,
	httpServer.NewHttpServerFx[*fm](),
	fx.Provide(func() app.EventListnerHttpServer {
		return httpServer.NewHttpServerV2[app.EventListnerHttpServer](httpServer.HttpServerV2Opts{})
	}),
	fx.Invoke(func(lf fx.Lifecycle, ev *env.Env, server app.EventListnerHttpServer, logger logging.Logger) {
		lf.Append(fx.Hook{
			OnStart: func(ctx context.Context) error {
				return httpServer.StartHttpServerV2(ctx, server, ev.RegistryEventListnerPort, logger.WithKV("server-name", "registry-event-listner"))
			},
			OnStop: func(context.Context) error {
				return httpServer.StopHttpServerV2(server)
			},
		})
	}),
)
