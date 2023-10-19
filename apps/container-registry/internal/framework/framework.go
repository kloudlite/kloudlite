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
	"kloudlite.io/pkg/redpanda"
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

func (fm *fm) GetBrokers() string {
	return fm.ev.KafkaBrokers
}

func (fm *fm) GetKafkaSASLAuth() *redpanda.KafkaSASLAuth {
	return nil
}

var Module = fx.Module("framework",
	fx.Provide(func(ev *env.Env) *fm {
		return &fm{ev}
	}),

	fx.Provide(func(ev *env.Env) (app.IAMGrpcClient, error) {
		return rpc.NewGrpcClient(ev.IAMGrpcAddr)
	}),

	fx.Provide(func(ev *env.Env) (app.AuthGrpcClient, error) {
		return rpc.NewGrpcClient(ev.AuthGrpcAddr)
	}),

	mongoDb.NewMongoClientFx[*fm](),

	// fx.Provide(func() kafka.Conn {
	//   return kafk
	// }),
	redpanda.NewClientFx[*fm](),

	fx.Provide(func(ev *env.Env) app.AuthCacheClient {
		return cache.NewRedisClient(ev.AuthRedisHosts, ev.AuthRedisUserName, ev.AuthRedisPassword, ev.AuthRedisPrefix)
	}),

	cache.FxLifeCycle[app.AuthCacheClient](),

	fx.Provide(func(ev *env.Env) cache.Client {
		return cache.NewRedisClient(ev.CRRedisHosts, ev.CRRedisUserName, ev.CRRedisPassword, ev.CRRedisPrefix)
	}),

	cache.FxLifeCycle[cache.Client](),

	app.Module,
	httpServer.NewHttpServerFx[*fm](),

	fx.Provide(func() app.AuthorizerHttpServer {
		return httpServer.NewHttpServerV2[app.AuthorizerHttpServer](httpServer.HttpServerV2Opts{})
	}),

	fx.Invoke(func(lf fx.Lifecycle, ev *env.Env, server app.AuthorizerHttpServer, logger logging.Logger) {
		lf.Append(fx.Hook{
			OnStart: func(ctx context.Context) error {
				return httpServer.StartHttpServerV2(ctx, server, ev.RegistryAuthorizerPort, logger.WithKV("server-name", "registry-authorizer"))
			},
			OnStop: func(context.Context) error {
				return httpServer.StopHttpServerV2(server)
			},
		})
	}),
)
