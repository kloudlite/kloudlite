package framework

import (
	"context"
	"fmt"
	"go.uber.org/fx"
	"google.golang.org/grpc"
	"kloudlite.io/apps/iam/internal/application"
	"kloudlite.io/pkg/config"
	"net"
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
				listen, portFormatError := net.Listen("tcp", fmt.Sprintf(":%d", env.Port))
				if portFormatError != nil {
					return portFormatError
				}
				go func() error {
					serverStartError := server.Serve(listen)
					if serverStartError != nil {
						return serverStartError
					}
					return nil
				}()
				return nil
			},
			OnStop: func(ctx context.Context) error {
				server.Stop()
				return nil
			},
		})
	}),
)
