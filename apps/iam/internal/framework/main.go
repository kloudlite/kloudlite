package framework

import (
	"context"
	"fmt"
	"go.uber.org/fx"
	"kloudlite.io/apps/iam/internal/app"
	"kloudlite.io/apps/iam/internal/env"
	"kloudlite.io/pkg/cache"
	"kloudlite.io/pkg/grpc"
	"kloudlite.io/pkg/logging"
	"kloudlite.io/pkg/repos"
	"time"
)

type fm struct {
	*env.Env
}

func (f *fm) RedisOptions() (hosts, username, password, basePrefix string) {
	return f.RedisHosts, f.RedisUsername, f.RedisPassword, f.RedisPrefix
}

func (f *fm) GetMongoConfig() (url, dbName string) {
	return f.MongoDbUri, f.MongoDbName
}

func (f *fm) GetGRPCPort() uint16 {
	return f.GrpcPort
}

var Module fx.Option = fx.Module(
	"framework",
	fx.Provide(func(ev *env.Env) *fm {
		return &fm{Env: ev}
	}),
	repos.NewMongoClientFx[*fm](),
	cache.NewRedisFx[*fm](),

	fx.Provide(func(logger logging.Logger) (app.IAMGrpcServer, error) {
		return grpc.NewGrpcServer(grpc.ServerOpts{
			Logger: logger,
		})
	}),

	app.Module,

	fx.Invoke(func(lf fx.Lifecycle, server app.IAMGrpcServer, ev *env.Env) {
		lf.Append(fx.Hook{
			OnStart: func(ctx context.Context) error {
				timeout, cf := context.WithTimeout(ctx, 2*time.Second)
				defer cf()
				errCh := make(chan error, 1)
				go func() {
					if err := server.Listen(fmt.Sprintf(":%d", ev.GrpcPort)); err != nil {
						errCh <- err
					}
				}()

				select {
				case <-timeout.Done():
				case err := <-errCh:
					return err
				}
				return nil
			},

			OnStop: func(ctx context.Context) error {
				server.Stop()
				return nil
			},
		})
	}),
)
