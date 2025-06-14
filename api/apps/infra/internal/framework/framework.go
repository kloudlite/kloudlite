package framework

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/kloudlite/api/apps/infra/internal/app"
	"github.com/kloudlite/api/apps/infra/internal/app/adapters"
	"github.com/kloudlite/api/apps/infra/internal/env"
	"github.com/kloudlite/api/common"
	"github.com/kloudlite/api/pkg/errors"
	"github.com/kloudlite/api/pkg/grpc"
	httpServer "github.com/kloudlite/api/pkg/http-server"
	"github.com/kloudlite/api/pkg/kv"
	"github.com/kloudlite/api/pkg/logging"
	"github.com/kloudlite/api/pkg/nats"

	mongoRepo "github.com/kloudlite/api/pkg/repos"
	"go.uber.org/fx"
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

	fx.Provide(func(ev *env.Env, logger *slog.Logger) (*nats.Client, error) {
		return nats.NewClient(ev.NatsURL, nats.ClientOpts{
			Name:   "infra",
			Logger: logger,
		})
	}),

	fx.Provide(func(c *nats.Client) (*nats.JetstreamClient, error) {
		return nats.NewJetstreamClient(c)
	}),

	fx.Provide(
		func(ev *env.Env, jc *nats.JetstreamClient) (kv.Repo[*common.AuthSession], error) {
			cxt := context.TODO()
			return kv.NewNatsKVRepo[*common.AuthSession](cxt, ev.SessionKVBucket, jc)
		},
	),

	fx.Provide(func(ev *env.Env) (app.IAMGrpcClient, error) {
		return grpc.NewGrpcClient(ev.IAMGrpcAddr)
	}),

	fx.Provide(func(ev *env.Env) (app.AccountGrpcClient, error) {
		return grpc.NewGrpcClient(ev.AccountsGrpcAddr)
	}),

	fx.Provide(func(ev *env.Env) (adapters.MesasageOfficeGRPCClient, error) {
		return grpc.NewGrpcClient(ev.MessageOfficeInternalGrpcAddr)
	}),

	fx.Provide(func(ev *env.Env) (app.ConsoleGrpcClient, error) {
		return grpc.NewGrpcClient(ev.ConsoleGrpcAddr)
	}),

	fx.Provide(func(ev *env.Env) (app.AuthGrpcClient, error) {
		return grpc.NewGrpcClient(ev.AuthGrpcAddr)
	}),

	fx.Invoke(func(lf fx.Lifecycle, c1 app.IAMGrpcClient) {
		lf.Append(fx.Hook{
			OnStop: func(context.Context) error {
				if err := c1.Close(); err != nil {
					return errors.NewE(err)
				}
				return nil
			},
		})
	}),

	app.Module,

	fx.Provide(func(logger *slog.Logger) (app.InfraGrpcServer, error) {
		return grpc.NewGrpcServer(grpc.ServerOpts{Logger: logger})
	}),

	fx.Invoke(func(ev *env.Env, server app.InfraGrpcServer, lf fx.Lifecycle, logger logging.Logger) {
		lf.Append(fx.Hook{
			OnStart: func(context.Context) error {
				go func() {
					if err := server.Listen(fmt.Sprintf(":%d", ev.GrpcPort)); err != nil {
						logger.Errorf(err, "while starting grpc server")
					}
				}()
				return nil
			},
			OnStop: func(context.Context) error {
				server.Stop()
				return nil
			},
		})
	}),

	fx.Provide(func(logger logging.Logger, e *env.Env) httpServer.Server {
		corsOrigins := "https://studio.apollographql.com"
		return httpServer.NewServer(httpServer.ServerArgs{Logger: logger, CorsAllowOrigins: &corsOrigins, IsDev: e.IsDev})
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
