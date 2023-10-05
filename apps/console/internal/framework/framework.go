package framework

import (
	"context"

	"github.com/kloudlite/operator/pkg/kubectl"
	"go.uber.org/fx"
	"k8s.io/client-go/rest"
	app "kloudlite.io/apps/console/internal/app"
	"kloudlite.io/apps/console/internal/env"
	"kloudlite.io/pkg/cache"
	rpc "kloudlite.io/pkg/grpc"
	httpServer "kloudlite.io/pkg/http-server"
	"kloudlite.io/pkg/k8s"
	"kloudlite.io/pkg/logging"
	loki_client "kloudlite.io/pkg/loki-client"
	"kloudlite.io/pkg/redpanda"
	mongoDb "kloudlite.io/pkg/repos"
)

type fm struct {
	ev *env.Env
}

func (fm *fm) GetBrokers() (brokers string) {
	return fm.ev.KafkaBrokers
}

func (fm *fm) GetKafkaSASLAuth() *redpanda.KafkaSASLAuth {
	return nil
	// return &redpanda.KafkaSASLAuth{
	// 	SASLMechanism: redpanda.ScramSHA256,
	// 	User:          fm.ev.KafkaUsername,
	// 	Password:      fm.ev.KafkaPassword,
	// }
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

	redpanda.NewClientFx[*fm](),

	app.Module,
	httpServer.NewHttpServerFx[*fm](),

	fx.Provide(func() app.LogsAndMetricsHttpServer {
		return httpServer.NewHttpServerV2[app.LogsAndMetricsHttpServer](httpServer.HttpServerV2Opts{})
	}),

	fx.Invoke(func(lf fx.Lifecycle, ev *env.Env, server app.LogsAndMetricsHttpServer, logger logging.Logger) {
		lf.Append(fx.Hook{
			OnStart: func(ctx context.Context) error {
				return httpServer.StartHttpServerV2(ctx, server, ev.LogsAndMetricsHttpPort, logger.WithKV("server-name", "logs-and-metrics"))
			},
			OnStop: func(context.Context) error {
				return httpServer.StopHttpServerV2(server)
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
