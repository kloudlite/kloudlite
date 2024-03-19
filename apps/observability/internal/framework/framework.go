package framework

import (
	"context"
	"fmt"

	"github.com/kloudlite/api/apps/observability/internal/app"
	"github.com/kloudlite/api/apps/observability/internal/env"
	"github.com/kloudlite/api/common"
	"github.com/kloudlite/api/pkg/grpc"
	httpServer "github.com/kloudlite/api/pkg/http-server"
	"github.com/kloudlite/api/pkg/kv"
	"github.com/kloudlite/api/pkg/logging"
	"github.com/kloudlite/api/pkg/nats"
	"go.uber.org/fx"
)

var Module = fx.Module("framework",
	fx.Provide(func(ev *env.Env) (app.IAMGrpcClient, error) {
		return grpc.NewGrpcClient(ev.IAMGrpcAddr)
	}),

	fx.Provide(func(ev *env.Env, logger logging.Logger) (*nats.Client, error) {
		return nats.NewClient(ev.NatsURL, nats.ClientOpts{
			Name:   "console",
			Logger: logger,
		})
	}),

	fx.Provide(func(c *nats.Client) (*nats.JetstreamClient, error) {
		return nats.NewJetstreamClient(c)
	}),

	fx.Provide(
		func(ev *env.Env, jc *nats.JetstreamClient) (app.SessionStore, error) {
			cxt := context.TODO()
			return kv.NewNatsKVRepo[*common.AuthSession](cxt, ev.SessionKVBucket, jc)
		},
	),

	fx.Provide(func(logger logging.Logger, ev *env.Env) httpServer.Server {
		return httpServer.NewServer(httpServer.ServerArgs{
			Logger:           logger,
			CorsAllowOrigins: &ev.HttpCorsOrigins,
			IsDev:            ev.IsDev,
		})
	}),

	fx.Provide(
		func(ev *env.Env, jc *nats.JetstreamClient) (kv.Repo[*common.AuthSession], error) {
			cxt := context.TODO()
			return kv.NewNatsKVRepo[*common.AuthSession](cxt, ev.SessionKVBucket, jc)
		},
	),

	app.Module,

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
