package framework

import (
	"context"
	"fmt"

	"go.uber.org/fx"
	"kloudlite.io/apps/infra/internal/app"
	"kloudlite.io/apps/infra/internal/env"
	"kloudlite.io/pkg/cache"
	"kloudlite.io/pkg/grpc"
	rpc "kloudlite.io/pkg/grpc"
	httpServer "kloudlite.io/pkg/http-server"
	"kloudlite.io/pkg/logging"
	"kloudlite.io/pkg/redpanda"
	mongoRepo "kloudlite.io/pkg/repos"
)

type framework struct {
	*env.Env
}

func (f *framework) GetBrokers() (brokers string) {
	return f.Env.KafkaBrokers
}

func (f *framework) GetKafkaSASLAuth() *redpanda.KafkaSASLAuth {
	return nil
	// return &redpanda.KafkaSASLAuth{
	// 	SASLMechanism: redpanda.ScramSHA256,
	// 	User:          f.Env.KafkaUsername,
	// 	Password:      f.Env.KafkaPassword,
	// }
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

	redpanda.NewClientFx[*framework](),

	fx.Provide(
		func(f *framework) app.AuthCacheClient {
			return cache.NewRedisClient(f.AuthRedisHosts, f.AuthRedisUserName, f.AuthRedisPassword, f.AuthRedisPrefix)
		},
	),

	fx.Provide(func(ev *env.Env) (app.IAMGrpcClient, error) {
		return rpc.NewGrpcClient(ev.IAMGrpcAddr)
	}),

	fx.Provide(func(ev *env.Env) (app.AccountGrpcClient, error) {
		return rpc.NewGrpcClient(ev.AccountsGrpcAddr)
	}),

	fx.Provide(func(ev *env.Env) (app.MessageOfficeInternalGrpcClient, error) {
		return rpc.NewGrpcClient(ev.MessageOfficeInternalGrpcAddr)
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

	httpServer.NewHttpServerFx[*framework](),
)
