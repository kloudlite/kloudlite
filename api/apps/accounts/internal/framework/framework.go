package framework

import (
	"context"
	"fmt"
	"time"

	"github.com/kloudlite/api/pkg/cache"
	"github.com/kloudlite/api/pkg/repos"

	"github.com/kloudlite/api/pkg/logging"

	"github.com/kloudlite/api/apps/accounts/internal/app"
	"github.com/kloudlite/api/apps/accounts/internal/env"
	"github.com/kloudlite/api/pkg/grpc"
	httpServer "github.com/kloudlite/api/pkg/http-server"
	"go.uber.org/fx"
)

type fm struct {
	env *env.Env
}

func (f *fm) GetMongoConfig() (url string, dbName string) {
	return f.env.DBUrl, f.env.DBName
}

var Module = fx.Module("framework",
	fx.Provide(func(ev *env.Env) *fm {
		return &fm{env: ev}
	}),

	fx.Provide(
		func(ev *env.Env) app.AuthCacheClient {
			return cache.NewRedisClient(
				ev.AuthRedisHosts,
				ev.AuthRedisUsername,
				ev.AuthRedisPassword,
				ev.AuthRedisPrefix,
			)
		},
	),
	cache.FxLifeCycle[app.AuthCacheClient](),
	repos.NewMongoClientFx[*fm](),

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

	app.Module,

	fx.Invoke(func(c1 app.AuthClient, c2 app.IAMClient, c3 app.CommsClient, c4 app.ContainerRegistryClient, c5 app.ConsoleClient, lf fx.Lifecycle) {
		lf.Append(fx.Hook{
			OnStop: func(context.Context) error {
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

	fx.Provide(func(logger logging.Logger) (app.AccountsGrpcServer, error) {
		return grpc.NewGrpcServer(grpc.ServerOpts{
			Logger: logger.WithKV("component", "grpc-server"),
		})
	}),

	fx.Invoke(func(lf fx.Lifecycle, server app.AccountsGrpcServer, ev *env.Env, logger logging.Logger) {
		lf.Append(fx.Hook{
			OnStart: func(ctx context.Context) error {
				errCh := make(chan error, 1)

				tctx, cf := context.WithTimeout(ctx, 2*time.Second)
				defer cf()

				go func() {
					err := server.Listen(fmt.Sprintf(":%d", ev.GrpcPort))
					if err != nil {
						errCh <- err
						logger.Errorf(err, "failed to start grpc server")
					}
				}()

				select {
				case <-tctx.Done():
				case err := <-errCh:
					return err
				}

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
