package framework

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/kloudlite/api/common"
	"github.com/kloudlite/api/pkg/nats"

	app "github.com/kloudlite/api/apps/container-registry/internal/app"
	"github.com/kloudlite/api/apps/container-registry/internal/env"
	"github.com/kloudlite/api/pkg/grpc"
	httpServer "github.com/kloudlite/api/pkg/http-server"
	"github.com/kloudlite/api/pkg/kv"
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

var Module = fx.Module("framework",
	fx.Provide(func(ev *env.Env) *fm {
		return &fm{ev}
	}),

	fx.Provide(func(ev *env.Env) (app.IAMGrpcClient, error) {
		return grpc.NewGrpcClient(ev.IAMGrpcAddr)
	}),

	fx.Provide(func(ev *env.Env) (app.AuthGrpcClient, error) {
		return grpc.NewGrpcClient(ev.AuthGrpcAddr)
	}),

	mongoDb.NewMongoClientFx[*fm](),

	fx.Provide(func(logger *slog.Logger, ev *env.Env) (*nats.Client, error) {
		name := "cr"
		return nats.NewClient(ev.NatsURL, nats.ClientOpts{
			Name:   name,
			Logger: logger,
		})
	}),

	fx.Provide(func(client *nats.Client) (*nats.JetstreamClient, error) {
		return nats.NewJetstreamClient(client)
	}),

	fx.Provide(
		func(ev *env.Env, jc *nats.JetstreamClient) (kv.Repo[*common.AuthSession], error) {
			cxt := context.TODO()
			return kv.NewNatsKVRepo[*common.AuthSession](cxt, ev.SessionKVBucket, jc)
		},
		func(ev *env.Env, jc *nats.JetstreamClient) (kv.BinaryDataRepo, error) {
			cxt := context.TODO()
			return kv.NewNatsKVBinaryRepo(cxt, ev.SessionKVBucket, jc)
		},
	),

	app.Module,

	fx.Provide(func(logger logging.Logger, e *env.Env) httpServer.Server {
		corsOrigins := "https://studio.apollographql.com"
		return httpServer.NewServer(httpServer.ServerArgs{Logger: logger, CorsAllowOrigins: &corsOrigins, IsDev: e.IsDev})
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

	// creates New GRPC server
	fx.Provide(func(logger *slog.Logger) (app.ContainerRegistryGRPCServer, error) {
		return grpc.NewGrpcServer(grpc.ServerOpts{Logger: logger})
	}),

	// handles GRPC server lifecycle
	fx.Invoke(func(lf fx.Lifecycle, server app.ContainerRegistryGRPCServer, ev *env.Env, logger logging.Logger) {
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
)
