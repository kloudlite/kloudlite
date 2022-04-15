package framework

import (
	"context"
	"fmt"

	"go.uber.org/fx"
	"google.golang.org/grpc"
	"kloudlite.io/apps/iam/internal/application"
	"kloudlite.io/pkg/cache"
	"kloudlite.io/pkg/config"
	rpc "kloudlite.io/pkg/grpc"
)

type Env struct {
	Port       int    `env:"GRPC_PORT" required:"true"`
	RedisHosts string `env:"REDIS_HOSTS" required:"true"`
}

var Module = fx.Module("framework",
	fx.Provide(config.LoadEnv[Env]()),
	fx.Provide(grpc.NewServer),
	application.Module,

	fx.Provide(func(env *Env) cache.Client {
		return cache.NewRedisClient(cache.RedisConnectOptions{
			Addr: env.RedisHosts,
		})
	}),

	fx.Invoke(func(lf fx.Lifecycle, cli cache.Client) {
		lf.Append(fx.Hook{
			OnStart: func(ctx context.Context) error {
				return cli.Connect(ctx)
			},
			OnStop: func(ctx context.Context) error {
				return cli.Close(ctx)
			},
		})
	}),

	fx.Invoke(func(lf fx.Lifecycle, env *Env, server *grpc.Server) {
		lf.Append(fx.Hook{
			OnStart: func(ctx context.Context) error {
				return rpc.GRPCStartServer(ctx, server, env.Port)
			},
			OnStop: func(context.Context) error {
				server.Stop()
				return nil
			},
		})
	}),
)
