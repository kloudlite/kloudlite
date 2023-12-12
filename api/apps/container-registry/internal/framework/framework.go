package framework

import (
	"context"
	"fmt"

	app "github.com/kloudlite/api/apps/container-registry/internal/app"
	"github.com/kloudlite/api/apps/container-registry/internal/env"
	"github.com/kloudlite/api/pkg/cache"
	rpc "github.com/kloudlite/api/pkg/grpc"
	httpServer "github.com/kloudlite/api/pkg/http-server"
	"github.com/kloudlite/api/pkg/kafka"
	"github.com/kloudlite/api/pkg/logging"
	mongoDb "github.com/kloudlite/api/pkg/repos"
	"go.uber.org/fx"
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

	fx.Provide(func(ev *env.Env) (kafka.Conn, error) {
		return kafka.Connect(ev.KafkaBrokers, kafka.ConnectOpts{})
	}),

	fx.Provide(func(ev *env.Env) app.AuthCacheClient {
		return cache.NewRedisClient(ev.AuthRedisHosts, ev.AuthRedisUserName, ev.AuthRedisPassword, ev.AuthRedisPrefix)
	}),

	// cache.FxLifeCycle[app.AuthCacheClient](),
	fx.Invoke(
		func(c app.AuthCacheClient, lf fx.Lifecycle, logger logging.Logger) {
			lf.Append(
				fx.Hook{
					OnStart: func(ctx context.Context) error {
						logger.Infof("connecting to redis")
						if err := c.Connect(ctx); err != nil {
							return err
						}
						logger.Infof("connected to redis")
						return nil
					},
					OnStop: func(ctx context.Context) error {
						return c.Disconnect(ctx)
					},
				},
			)
		},
	),

	fx.Provide(func(ev *env.Env) cache.Client {
		return cache.NewRedisClient(ev.CRRedisHosts, ev.CRRedisUserName, ev.CRRedisPassword, ev.CRRedisPrefix)
	}),

	// cache.FxLifeCycle[cache.Client](),

	fx.Invoke(
		func(c cache.Client, lf fx.Lifecycle, logger logging.Logger) {
			lf.Append(
				fx.Hook{
					OnStart: func(ctx context.Context) error {
						logger.Infof("connecting to redis")
						if err := c.Connect(ctx); err != nil {
							return err
						}
						logger.Infof("connected to redis")
						return nil
					},
					OnStop: func(ctx context.Context) error {
						return c.Disconnect(ctx)
					},
				},
			)
		},
	),

	app.Module,

	fx.Provide(func(logger logging.Logger) httpServer.Server {
		corsOrigins := "https://studio.apollographql.com"
		return httpServer.NewServer(httpServer.ServerArgs{Logger: logger, CorsAllowOrigins: &corsOrigins})
	}),

	fx.Invoke(func(lf fx.Lifecycle, server httpServer.Server, ev *env.Env) {
		lf.Append(fx.Hook{
			OnStart: func(context.Context) error {
				return server.Listen(fmt.Sprintf(":%d", ev.Port))
			},
			OnStop: func(context.Context) error {
				return server.Close()
			},
		})
	}),

	fx.Provide(func(logger logging.Logger) app.AuthorizerHttpServer {
		return httpServer.NewServer(httpServer.ServerArgs{Logger: logger})
	}),

	fx.Invoke(func(lf fx.Lifecycle, server app.AuthorizerHttpServer, ev *env.Env) {
		lf.Append(fx.Hook{
			OnStart: func(context.Context) error {
				return server.Listen(fmt.Sprintf(":%d", ev.RegistryAuthorizerPort))
			},
			OnStop: func(context.Context) error {
				return server.Close()
			},
		})
	}),
)
