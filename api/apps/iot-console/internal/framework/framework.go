package framework

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/kloudlite/api/apps/iot-console/internal/app"
	"github.com/kloudlite/api/apps/iot-console/internal/domain"
	"github.com/kloudlite/api/apps/iot-console/internal/env"
	"github.com/kloudlite/api/common"
	httpServer "github.com/kloudlite/api/pkg/http-server"
	"github.com/kloudlite/api/pkg/k8s"
	"github.com/kloudlite/api/pkg/kv"
	"github.com/kloudlite/api/pkg/logging"
	"github.com/kloudlite/api/pkg/nats"
	mongoDb "github.com/kloudlite/api/pkg/repos"
	"go.uber.org/fx"
	"k8s.io/client-go/rest"
)

type fm struct {
	ev *env.Env
}

func (fm *fm) GetMongoConfig() (url string, dbName string) {
	return fm.ev.IotDeviceDBUri, fm.ev.IotDeviceDBName
}

var Module = fx.Module("framework",
	fx.Provide(func(ev *env.Env) *fm {
		return &fm{ev}
	}),

	mongoDb.NewMongoClientFx[*fm](),

	fx.Provide(func(ev *env.Env, logger *slog.Logger) (*nats.Client, error) {
		return nats.NewClient(ev.NatsURL, nats.ClientOpts{
			Name:   "console",
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

	fx.Provide(
		func(ev *env.Env, jc *nats.JetstreamClient) (domain.IOTConsoleCacheStore, error) {
			return kv.NewNatsKVBinaryRepo(context.TODO(), ev.IOTConsoleCacheKVBucket, jc)
		},
	),

	fx.Provide(func(restCfg *rest.Config) (k8s.Client, error) {
		return k8s.NewClient(restCfg, nil)
	}),

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
)
