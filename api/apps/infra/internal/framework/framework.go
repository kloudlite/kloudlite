package framework

import (
	"context"
	"fmt"

	"go.uber.org/fx"
	"kloudlite.io/apps/infra/internal/app"
	"kloudlite.io/apps/infra/internal/env"
	"kloudlite.io/pkg/cache"
	"kloudlite.io/pkg/grpc"
	httpServer "kloudlite.io/pkg/http-server"
	"kloudlite.io/pkg/kafka"
	"kloudlite.io/pkg/logging"
	mongoRepo "kloudlite.io/pkg/repos"
)

type framework struct {
	*env.Env
}

func (f *framework) GetHttpCors() string {
	return "https://studio.apollographql.com"
}

func (f *framework) GetHttpPort() uint16 {
	return f.HttpPort
}

func (f *framework) GetMongoConfig() (url string, dbName string) {
	return f.InfraDbUri, f.InfraDbName
}

var Module = fx.Module("framework",
	fx.Provide(func(ev *env.Env) *framework {
		return &framework{Env: ev}
	}),

	mongoRepo.NewMongoClientFx[*framework](),

	fx.Provide(func(ev *env.Env) (kafka.Conn, error) {
		return kafka.Connect(ev.KafkaBrokers, kafka.ConnectOpts{})
	}),

	fx.Provide(
		func(f *framework) app.AuthCacheClient {
			return cache.NewRedisClient(f.AuthRedisHosts, f.AuthRedisUserName, f.AuthRedisPassword, f.AuthRedisPrefix)
		},
	),

	fx.Provide(func(ev *env.Env) (app.IAMGrpcClient, error) {
		return grpc.NewGrpcClient(ev.IAMGrpcAddr)
	}),

	fx.Provide(func(ev *env.Env) (app.AccountGrpcClient, error) {
		return grpc.NewGrpcClient(ev.AccountsGrpcAddr)
	}),

	fx.Provide(func(ev *env.Env) (app.MessageOfficeInternalGrpcClient, error) {
		return grpc.NewGrpcClient(ev.MessageOfficeInternalGrpcAddr)
	}),

	fx.Invoke(func(lf fx.Lifecycle, c1 app.IAMGrpcClient) {
		lf.Append(fx.Hook{
			OnStop: func(context.Context) error {
				if err := c1.Close(); err != nil {
					return err
				}
				return nil
			},
		})
	}),

	cache.FxLifeCycle[app.AuthCacheClient](),
	app.Module,

	fx.Provide(func(logr logging.Logger) (app.InfraGrpcServer, error) {
		return grpc.NewGrpcServer(grpc.ServerOpts{
			Logger: logr,
		})
	}),

	fx.Invoke(func(ev *env.Env, server app.InfraGrpcServer, lf fx.Lifecycle) {
		lf.Append(fx.Hook{
			OnStart: func(context.Context) error {
				go server.Listen(fmt.Sprintf(":%d", ev.GrpcPort))
				return nil
			},
			OnStop: func(context.Context) error {
				server.Stop()
				return nil
			},
		})
	}),

	fx.Provide(func(logger logging.Logger) httpServer.Server {
		corsOrigins := "https://studio.apollographql.com"
		return httpServer.NewServer(httpServer.ServerArgs{Logger: logger, CorsAllowOrigins: &corsOrigins})
	}),

	fx.Invoke(func(lf fx.Lifecycle, server httpServer.Server, ev *env.Env) {
		lf.Append(fx.Hook{
			OnStart: func(context.Context) error {
				return server.Listen(fmt.Sprintf(":%d", ev.HttpPort))
			},
			OnStop: func(context.Context) error {
				return server.Close()
			},
		})
	}),
)
