package framework

import (
	"context"
	"fmt"
	"github.com/kloudlite/api/apps/iam/internal/app"
	"github.com/kloudlite/api/apps/iam/internal/env"
	"github.com/kloudlite/api/pkg/errors"
	"github.com/kloudlite/api/pkg/grpc"
	"github.com/kloudlite/api/pkg/logging"
	"github.com/kloudlite/api/pkg/repos"
	"go.uber.org/fx"
	"time"
)

type fm struct {
	*env.Env
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
					return errors.NewE(err)
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
