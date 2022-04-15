package framework

import (
	"context"
	"go.uber.org/fx"
	"google.golang.org/grpc"
	"kloudlite.io/apps/iam/internal/application"
	"kloudlite.io/pkg/config"
	rpc "kloudlite.io/pkg/grpc"
)

type Env struct {
	Port int `env:"GRPC_PORT"`
}

var Module = fx.Module("framework",
	fx.Provide(config.LoadEnv[Env]()),
	fx.Provide(grpc.NewServer),
	application.Module,
	fx.Invoke(func(lifecycle fx.Lifecycle, env *Env, server *grpc.Server) {
		lifecycle.Append(fx.Hook{
			OnStart: func(ctx context.Context) error {
				return rpc.GRPCStartServer(ctx, server, env.Port)
			},
			OnStop: func(ctx context.Context) error {
				server.Stop()
				return nil
			},
		})
	}),
)
