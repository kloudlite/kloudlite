package framework

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/kloudlite/api/apps/console/internal/domain"
	"github.com/kloudlite/api/common"
	"github.com/kloudlite/api/pkg/errors"

	app "github.com/kloudlite/api/apps/console/internal/app"
	"github.com/kloudlite/api/apps/console/internal/env"
	httpServer "github.com/kloudlite/api/pkg/http-server"
	"github.com/kloudlite/api/pkg/k8s"
	"github.com/kloudlite/api/pkg/kv"
	"github.com/kloudlite/api/pkg/logging"
	"github.com/kloudlite/api/pkg/nats"
	mongoDb "github.com/kloudlite/api/pkg/repos"

	"github.com/kloudlite/api/pkg/grpc"
	"go.uber.org/fx"
	"k8s.io/client-go/rest"

	"github.com/miekg/dns"
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
		func(ev *env.Env, jc *nats.JetstreamClient) (domain.ConsoleCacheStore, error) {
			return kv.NewNatsKVBinaryRepo(context.TODO(), ev.ConsoleCacheKVBucket, jc)
		},
	),

	fx.Provide(func(restCfg *rest.Config) (k8s.Client, error) {
		return k8s.NewClient(restCfg, nil)
	}),

	// grpc clients
	fx.Provide(func(ev *env.Env) (app.IAMGrpcClient, error) {
		return grpc.NewGrpcClient(ev.IAMGrpcAddr)
	}),

	fx.Provide(func(ev *env.Env) (app.InfraClient, error) {
		return grpc.NewGrpcClient(ev.InfraGrpcAddr)
	}),

	fx.Provide(func(ev *env.Env) (app.AccountsClient, error) {
		return grpc.NewGrpcClient(ev.AccountGRPCAddr)
	}),

	fx.Provide(func(ev *env.Env) (app.MessageOfficeInternalClient, error) {
		return grpc.NewGrpcClient(ev.MessageOfficeInternalGRPCAddr)
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

	fx.Provide(func(logger *slog.Logger) (app.ConsoleGrpcServer, error) {
		return grpc.NewGrpcServer(grpc.ServerOpts{Logger: logger})
	}),

	fx.Invoke(func(ev *env.Env, server app.ConsoleGrpcServer, lf fx.Lifecycle, logger logging.Logger) {
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
				return server.Listen(fmt.Sprintf(":%d", ev.Port))
			},
			OnStop: func(context.Context) error {
				return server.Close()
			},
		})
	}),

	fx.Provide(func(ev *env.Env) *app.DNSServer {
		return &app.DNSServer{
			Server: &dns.Server{
				Addr:      ev.DNSAddr,
				Net:       "udp",
				UDPSize:   0xffff,
				ReusePort: true,
			},
		}
	}),
)
