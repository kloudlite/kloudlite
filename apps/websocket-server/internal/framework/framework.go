package framework

import (
	"context"
	"fmt"

	"github.com/kloudlite/api/common"
	"github.com/kloudlite/api/pkg/errors"
	"github.com/kloudlite/api/pkg/nats"

	"github.com/kloudlite/api/pkg/kv"

	"github.com/kloudlite/api/pkg/logging"

	"github.com/kloudlite/api/apps/websocket-server/internal/app"
	"github.com/kloudlite/api/apps/websocket-server/internal/env"
	"github.com/kloudlite/api/pkg/grpc"
	httpServer "github.com/kloudlite/api/pkg/http-server"
	"go.uber.org/fx"
)

type fm struct {
	env *env.Env
}

var Module = fx.Module("framework",
	fx.Provide(func(ev *env.Env) *fm {
		return &fm{env: ev}
	}),

	fx.Provide(func(ev *env.Env, logger logging.Logger) (*nats.Client, error) {
		name := "RUP:nat-client"
		return nats.NewClient(ev.NatsURL, nats.ClientOpts{
			Name:   name,
			Logger: logger,
		})
	}),

	fx.Provide(func(ev *env.Env, logger logging.Logger, cli *nats.Client) (*nats.JetstreamClient, error) {
		return nats.NewJetstreamClient(cli)
	}),

	fx.Provide(
		func(ev *env.Env, jc *nats.JetstreamClient) (kv.Repo[*common.AuthSession], error) {
			cxt := context.TODO()
			return kv.NewNatsKVRepo[*common.AuthSession](cxt, ev.SessionKVBucket, jc)
		},
	),

	fx.Provide(func(ev *env.Env) (app.IAMClient, error) {
		return grpc.NewGrpcClient(ev.IamGrpcAddr)
	}),

	app.Module,

	fx.Invoke(func(c2 app.IAMClient, lf fx.Lifecycle) {
		lf.Append(fx.Hook{
			OnStop: func(context.Context) error {
				if err := c2.Close(); err != nil {
					return errors.NewE(err)
				}
				return nil
			},
		})
	}),

	fx.Provide(func(logger logging.Logger, e *env.Env) httpServer.Server {
		corsOrigins := "https://console.devc.kloudlite.io"
		return httpServer.NewServer(httpServer.ServerArgs{Logger: logger, CorsAllowOrigins: &corsOrigins, IsDev: e.IsDev})
	}),

	// have to create socket server here
	fx.Invoke(func(lf fx.Lifecycle, server httpServer.Server, ev *env.Env) {
		lf.Append(fx.Hook{
			OnStart: func(context.Context) error {
				return server.Listen(fmt.Sprintf(":%d", ev.SocketPort))
			},
			OnStop: func(context.Context) error {
				return server.Close()
			},
		})
	}),
)
