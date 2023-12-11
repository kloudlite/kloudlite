package framework

import (
	"context"
	"fmt"

	"go.uber.org/fx"

	"kloudlite.io/apps/message-office/internal/app"
	"kloudlite.io/apps/message-office/internal/env"
	"kloudlite.io/pkg/grpc"
	httpServer "kloudlite.io/pkg/http-server"
	"kloudlite.io/pkg/logging"
	"kloudlite.io/pkg/messaging/nats"
	mongoDb "kloudlite.io/pkg/repos"
)

type fm struct {
	*env.Env
}

func (f *fm) GetMongoConfig() (url string, dbName string) {
	return f.DbUri, f.DbName
}

func (f *fm) GetHttpCors() string {
	return ""
}

var Module = fx.Module("framework",
	fx.Provide(func(ev *env.Env) *fm {
		return &fm{Env: ev}
	}),
	mongoDb.NewMongoClientFx[*fm](),

	fx.Provide(func(f *fm) (app.RealVectorGrpcClient, error) {
		return grpc.NewGrpcClient(f.VectorGrpcAddr)
	}),

	fx.Provide(func(ev *env.Env, logger logging.Logger) (*nats.JetstreamClient, error) {
		nc, err := nats.NewClient(ev.NatsUrl, nats.ClientOpts{
			Name:   "message-offfice",
			Logger: logger,
		})
		if err != nil {
			return nil, err
		}

		return nats.NewJetstreamClient(nc)
	}),

	app.Module,

	fx.Provide(func(logr logging.Logger) (app.InternalGrpcServer, error) {
		return grpc.NewGrpcServer(grpc.ServerOpts{
			Logger: logr.WithName("internal-grpc-server"),
		})
	}),

	fx.Invoke(func(lf fx.Lifecycle, server app.InternalGrpcServer, ev *env.Env) {
		lf.Append(fx.Hook{
			OnStart: func(context.Context) error {
				go server.Listen(fmt.Sprintf(":%d", ev.InternalGrpcPort))
				return nil
			},
			OnStop: func(context.Context) error {
				server.Stop()
				return nil
			},
		})
	}),

	fx.Provide(func(logr logging.Logger) (app.ExternalGrpcServer, error) {
		return grpc.NewGrpcServer(grpc.ServerOpts{
			Logger: logr.WithName("external-grpc-server"),
		})
	}),

	fx.Invoke(func(lf fx.Lifecycle, server app.ExternalGrpcServer, ev *env.Env) {
		lf.Append(fx.Hook{
			OnStart: func(context.Context) error {
				go server.Listen(fmt.Sprintf(":%d", ev.ExternalGrpcPort))
				return nil
			},
			OnStop: func(context.Context) error {
				server.Stop()
				return nil
			},
		})
	}),

	fx.Provide(func(logger logging.Logger) httpServer.Server {
		return httpServer.NewServer(httpServer.ServerArgs{Logger: logger})
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
