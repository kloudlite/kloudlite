package framework

import (
	"context"
	"fmt"
	"github.com/kloudlite/operator/pkg/kubectl"
	"go.uber.org/fx"
	"k8s.io/client-go/rest"
	app "kloudlite.io/apps/console/internal/app"
	"kloudlite.io/apps/console/internal/env"
	"kloudlite.io/pkg/cache"
	rpc "kloudlite.io/pkg/grpc"
	httpServer "kloudlite.io/pkg/http-server"
	"kloudlite.io/pkg/k8s"
	"kloudlite.io/pkg/kafka"
	"kloudlite.io/pkg/logging"
	loki_client "kloudlite.io/pkg/loki-client"
	mongoDb "kloudlite.io/pkg/repos"
)

type fm struct {
	ev *env.Env
}

func (fm *fm) GetMongoConfig() (url string, dbName string) {
	return fm.ev.ConsoleDBUri, fm.ev.ConsoleDBName
}

func (fm *fm) RedisOptions() (hosts, username, password, basePrefix string) {
	return fm.ev.AuthRedisHosts, fm.ev.AuthRedisUserName, fm.ev.AuthRedisPassword, fm.ev.AuthRedisPrefix
}

func (fm *fm) GetHttpPort() uint16 {
	return fm.ev.Port
}

func (fm *fm) GetHttpCors() string {
	return "*"
}

var Module = fx.Module("framework",
	fx.Provide(func(ev *env.Env) *fm {
		return &fm{ev}
	}),

	mongoDb.NewMongoClientFx[*fm](),

	fx.Provide(func(ev *env.Env) app.AuthCacheClient {
		return cache.NewRedisClient(ev.AuthRedisHosts, ev.AuthRedisUserName, ev.AuthRedisPassword, ev.AuthRedisPrefix)
	}),

	cache.FxLifeCycle[app.AuthCacheClient](),

	fx.Provide(func(restCfg *rest.Config) (kubectl.YAMLClient, error) {
		return kubectl.NewYAMLClient(restCfg)
	}),

	fx.Provide(func(restCfg *rest.Config) (k8s.ExtendedK8sClient, error) {
		return k8s.NewExtendedK8sClient(restCfg)
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
					return err
				}
				if err := c2.Close(); err != nil {
					return err
				}
				return nil
			},
		})
	}),

	fx.Provide(func(ev *env.Env) (kafka.Conn, error) {
		return kafka.Connect(ev.KafkaBrokers, kafka.ConnectOpts{})
	}),

	app.Module,

	fx.Provide(func(logger logging.Logger) httpServer.Server {
		corsOrigins := "https://studio.apollographql.com"
		return httpServer.NewServer(httpServer.ServerArgs{Logger: logger, CorsAllowOrigins: &corsOrigins})
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
					return err
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
