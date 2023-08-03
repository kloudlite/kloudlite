package framework

import (
	"context"

	"go.uber.org/fx"
	"kloudlite.io/apps/accounts/internal/app"
	"kloudlite.io/apps/accounts/internal/env"
	"kloudlite.io/pkg/grpc"
)

var Module = fx.Module("framework",
	fx.Provide(func(ev *env.Env) (app.AuthClient, error) {
		return grpc.NewGrpcClient(ev.AuthGrpcAddr)
	}),

	fx.Provide(func(ev *env.Env) (app.IAMClient, error) {
		return grpc.NewGrpcClient(ev.IamGrpcAddr)
	}),

	fx.Provide(func(ev *env.Env) (app.CommsClient, error) {
		return grpc.NewGrpcClient(ev.CommsGrpcAddr)
	}),

	fx.Provide(func(ev *env.Env) (app.ContainerRegistryClient, error) {
		return grpc.NewGrpcClient(ev.ContainerRegistryGrpcAddr)
	}),

	fx.Provide(func(ev *env.Env) (app.ConsoleClient, error) {
		return grpc.NewGrpcClient(ev.ConsoleGrpcAddr)
	}),

	fx.Invoke(func(c1 app.AuthClient, c2 app.IAMClient, c3 app.CommsClient, c4 app.ContainerRegistryClient, c5 app.ConsoleClient, lf fx.Lifecycle) {
		lf.Append(fx.Hook{
			OnStop: func(ctx context.Context) error {
				if err := c1.Close(); err != nil {
					return err
				}
				if err := c2.Close(); err != nil {
					return err
				}
				if err := c3.Close(); err != nil {
					return err
				}
				if err := c4.Close(); err != nil {
					return err
				}
				if err := c5.Close(); err != nil {
					return err
				}
				return nil
			},
		})
	}),

	fx.Provide(func() (grpc.GrpcServer, error) {
		return grpc.NewGrpcServer(grpc.GrpcServerOpts{})
	}),

	fx.Invoke(func() error {
	}),

	app.Module,
)
