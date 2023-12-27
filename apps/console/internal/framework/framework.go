package framework

import (
	"context"
	"fmt"
	"github.com/kloudlite/api/common"
	"github.com/kloudlite/api/pkg/errors"

	app "github.com/kloudlite/api/apps/console/internal/app"
	"github.com/kloudlite/api/apps/console/internal/env"
	"github.com/kloudlite/api/pkg/cache"
	rpc "github.com/kloudlite/api/pkg/grpc"
	httpServer "github.com/kloudlite/api/pkg/http-server"
	"github.com/kloudlite/api/pkg/k8s"
	"github.com/kloudlite/api/pkg/logging"
	loki_client "github.com/kloudlite/api/pkg/loki-client"
	"github.com/kloudlite/api/pkg/nats"
	mongoDb "github.com/kloudlite/api/pkg/repos"
	"go.uber.org/fx"
	"k8s.io/client-go/rest"
)

type fm struct {
	ev *env.Env
}

func (fm *fm) GetMongoConfig() (url string, dbName string) {
	return fm.ev.ConsoleDBUri, fm.ev.ConsoleDBName
}

var Module = fx.Module("framework",
	fx.Provide(func(ev *env.Env) *fm {
		return &fm{ev}
	}),

	mongoDb.NewMongoClientFx[*fm](),

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
		func(ev *env.Env, jc *nats.JetstreamClient) (cache.Repo[*common.AuthSession], error) {
			cxt := context.TODO()
			return cache.NewNatsKVRepo[*common.AuthSession](cxt, ev.SessionKVBucket, jc)
		},
	),

	fx.Provide(func(restCfg *rest.Config) (k8s.Client, error) {
		return k8s.NewClient(restCfg, nil)
	}),

	fx.Provide(func(ev *env.Env) (app.IAMGrpcClient, error) {
		return rpc.NewGrpcClient(ev.IAMGrpcAddr)
	}),

	fx.Provide(func(ev *env.Env) (app.InfraClient, error) {
		return rpc.NewGrpcClient(ev.InfraGrpcAddr)
	}),

	fx.Invoke(func(lf fx.Lifecycle, c1 app.IAMGrpcClient, c2 app.InfraClient) {
		lf.Append(fx.Hook{
			OnStop: func(context.Context) error {
				if err := c1.Close(); err != nil {
					return errors.NewE(err)
				}
				if err := c2.Close(); err != nil {
					return errors.NewE(err)
				}
				return nil
			},
		})
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

	fx.Provide(func(logger logging.Logger) app.LogsAndMetricsHttpServer {
		return httpServer.NewServer(httpServer.ServerArgs{Logger: logger})
	}),

	fx.Invoke(func(lf fx.Lifecycle, ev *env.Env, server app.LogsAndMetricsHttpServer, logger logging.Logger) {
		lf.Append(fx.Hook{
			OnStart: func(ctx context.Context) error {
				return server.Listen(fmt.Sprintf(":%d", ev.LogsAndMetricsHttpPort))
			},
			OnStop: func(context.Context) error {
				return server.Close()
			},
		})
	}),

	fx.Provide(func(ev *env.Env, logger logging.Logger) (loki_client.LokiClient, error) {
		return loki_client.NewLokiClient(ev.LokiServerHttpAddr, loki_client.ClientOpts{Logger: logger.WithKV("component", "loki-client")})
	}),

	fx.Invoke(func(lf fx.Lifecycle, lc loki_client.LokiClient, logger logging.Logger, ev *env.Env) {
		lf.Append(fx.Hook{
			OnStart: func(ctx context.Context) error {
				if err := lc.Ping(ctx); err != nil {
					return errors.NewE(err)
				}
				logger.Infof("loki client successfully connected to loki server running @ %s", ev.LokiServerHttpAddr)
				return nil
			},
			OnStop: func(context.Context) error {
				lc.Close()
				return nil
			},
		})
	}),
)
