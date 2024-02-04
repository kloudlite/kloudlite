package framework

import (
	"context"
	"fmt"
	"github.com/kloudlite/api/apps/webhook/internal/app"
	"github.com/kloudlite/api/apps/webhook/internal/env"
	"github.com/kloudlite/api/pkg/errors"
	httpServer "github.com/kloudlite/api/pkg/http-server"
	"github.com/kloudlite/api/pkg/logging"
	"github.com/kloudlite/api/pkg/nats"
	"go.uber.org/fx"
)

type fm struct {
	*env.Env
}

func (f fm) GetHttpPort() uint16 {
	return f.HttpPort
}

func (f fm) GetHttpCors() string {
	return ""
}

var Module = fx.Module(
	"framework",
	fx.Provide(
		func(vars *env.Env) *fm {
			return &fm{Env: vars}
		},
	),

	fx.Provide(func(ev *env.Env, logger logging.Logger) (*nats.JetstreamClient, error) {
		name := "webhook:jetstream-client"
		nc, err := nats.NewClient(ev.NatsURL, nats.ClientOpts{
			Name:   name,
			Logger: logger,
		})
		if err != nil {
			return nil, errors.NewE(err)
		}

		return nats.NewJetstreamClient(nc)
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
	app.Module,
)
