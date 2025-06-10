package framework

import (
	"context"
	"fmt"
	"log/slog"
	"net"

	"github.com/kloudlite/api/common"
	"github.com/kloudlite/api/pkg/errors"
	"github.com/kloudlite/api/pkg/kv"
	"github.com/kloudlite/api/pkg/logging"
	"github.com/kloudlite/api/pkg/nats"
	"google.golang.org/grpc"

	"go.uber.org/fx"

	"github.com/kloudlite/api/apps/auth/internal/app"
	"github.com/kloudlite/api/apps/auth/internal/env"
	rpc "github.com/kloudlite/api/pkg/grpc"
	"github.com/kloudlite/api/pkg/repos"
)

type CommsGrpcEnv struct {
	*env.Env
}

func (c CommsGrpcEnv) GetGRPCServerURL() string {
	return c.CommsService
}

type fm struct {
	*env.Env
}

func (e *fm) GetHttpPort() uint16 {
	return e.Port
}

func (e *fm) GetHttpCors() string {
	return e.CorsOrigins
}

func (e *fm) GetMongoConfig() (url string, dbName string) {
	return e.MongoUri, e.MongoDbName
}

func (e *fm) GetGRPCPort() uint16 {
	return e.GrpcPort
}

var Module fx.Option = fx.Module(
	"framework",
	fx.Provide(func(ev *env.Env) *fm {
		return &fm{ev}
	}),
	fx.Provide(func(ev *env.Env) *CommsGrpcEnv {
		return &CommsGrpcEnv{ev}
	}),

	repos.NewMongoClientFx[*fm](),

	fx.Provide(func(ev *env.Env, logger *slog.Logger) (*nats.JetstreamClient, error) {
		name := "auth:jetstream-client"
		nc, err := nats.NewClient(ev.NatsURL, nats.ClientOpts{
			Name:   name,
			Logger: logger,
		})
		if err != nil {
			return nil, errors.NewE(err)
		}

		return nats.NewJetstreamClient(nc)
	}),

	fx.Provide(
		func(ev *env.Env, jc *nats.JetstreamClient) (kv.Repo[*common.AuthSession], error) {
			cxt := context.TODO()
			return kv.NewNatsKVRepo[*common.AuthSession](cxt, ev.SessionKVBucket, jc)
		},
	),

	rpc.NewGrpcClientFx[*CommsGrpcEnv, app.CommsClientConnection](),

	// rpc.NewGrpcServerFx[*fm](),

	fx.Module(
		"grpc-servers",

		fx.Provide(func(logger logging.Logger) app.AuthGrpcServer {
			return grpc.NewServer()
		}),

		fx.Invoke(
			func(lf fx.Lifecycle, server app.AuthGrpcServer, env *env.Env, logger *slog.Logger) error {
				listener, err := net.Listen("tcp", fmt.Sprintf(":%d", env.GrpcPort))
				if err != nil {
					return errors.NewEf(err, "could not listen on port %d", env.GrpcPort)
				}
				lf.Append(
					fx.Hook{
						OnStart: func(ctx context.Context) error {
							logger.Info(fmt.Sprintf("Starting Auth gRPC server on port %d", env.GrpcPort))
							go ((*grpc.Server)(server)).Serve(listener)
							return nil
						},
						OnStop: func(ctx context.Context) error {
							logger.Info(fmt.Sprintf("Stopping Auth gRPC server on port %d", env.GrpcPort))
							go ((*grpc.Server)(server)).Stop()
							return nil
						},
					},
				)
				return nil
			},
		),

		fx.Provide(func(logger logging.Logger) app.AuthGrpcServerV2 {
			return grpc.NewServer()
		}),

		fx.Invoke(
			func(lf fx.Lifecycle, server app.AuthGrpcServerV2, env *env.Env, logger *slog.Logger) error {
				listener, err := net.Listen("tcp", fmt.Sprintf(":%d", env.GrpcV2Port))
				if err != nil {
					return errors.NewEf(err, "could not listen on port %d", env.GrpcV2Port)
				}
				lf.Append(
					fx.Hook{
						OnStart: func(ctx context.Context) error {
							logger.Info(fmt.Sprintf("Starting Auth V2 gRPC server on port %d", env.GrpcV2Port))
							go ((*grpc.Server)(server)).Serve(listener)
							return nil
						},
						OnStop: func(ctx context.Context) error {
							logger.Info(fmt.Sprintf("Stopping Auth V2 gRPC server on port %d", env.GrpcV2Port))
							go ((*grpc.Server)(server)).Stop()
							return nil
						},
					},
				)
				return nil
			},
		),
	),

	app.Module,
)
