package framework

import (
	"context"
	"fmt"
	"time"

	"github.com/kloudlite/api/apps/comms/internal/app"
	"github.com/kloudlite/api/apps/comms/internal/env"
	"github.com/kloudlite/api/common"
	"github.com/kloudlite/api/pkg/errors"
	"github.com/kloudlite/api/pkg/grpc"
	rpc "github.com/kloudlite/api/pkg/grpc"
	httpServer "github.com/kloudlite/api/pkg/http-server"
	"github.com/kloudlite/api/pkg/kv"
	"github.com/kloudlite/api/pkg/logging"
	"github.com/kloudlite/api/pkg/mail"
	"github.com/kloudlite/api/pkg/nats"
	mongoDb "github.com/kloudlite/api/pkg/repos"
	"go.uber.org/fx"
)

type fm struct {
	ev *env.Env
}

func (fm *fm) GetMongoConfig() (url string, dbName string) {
	return fm.ev.CommsDBUri, fm.ev.CommsDBName
}

var Module = fx.Module(
	"framework",

	fx.Provide(func(ev *env.Env) *fm {
		return &fm{ev}
	}),

	fx.Provide(func(ev *env.Env, logger logging.Logger) (*nats.Client, error) {
		return nats.NewClient(ev.NatsURL, nats.ClientOpts{
			Name:   "comms",
			Logger: logger,
		})
	}),

	fx.Provide(func(c *nats.Client) (*nats.JetstreamClient, error) {
		return nats.NewJetstreamClient(c)
	}),

	fx.Provide(func(ev *env.Env) mail.Mailer {
		return mail.NewSendgridMailer(ev.SendgridApiKey)
	}),

	fx.Provide(func(logger logging.Logger) (app.CommsGrpcServer, error) {
		return grpc.NewGrpcServer(grpc.ServerOpts{Logger: logger})
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

	fx.Provide(
		func(ev *env.Env, jc *nats.JetstreamClient) (kv.Repo[*common.AuthSession], error) {
			cxt := context.TODO()
			return kv.NewNatsKVRepo[*common.AuthSession](cxt, ev.SessionKVBucket, jc)
		},
	),

	fx.Provide(func(ev *env.Env) (app.IAMGrpcClient, error) {
		return rpc.NewGrpcClient(ev.IAMGrpcAddr)
	}),

	mongoDb.NewMongoClientFx[*fm](),

	fx.Invoke(func(lf fx.Lifecycle, server app.CommsGrpcServer, ev *env.Env, logger logging.Logger) {
		lf.Append(fx.Hook{
			OnStart: func(ctx context.Context) error {
				errCh := make(chan error, 1)

				tctx, cf := context.WithTimeout(ctx, 2*time.Second)
				defer cf()

				go func() {
					err := server.Listen(fmt.Sprintf(":%d", ev.GrpcPort))
					if err != nil {
						errCh <- err
						logger.Errorf(err, "failed to start grpc server")
					}
				}()

				select {
				case <-tctx.Done():
				case err := <-errCh:
					return errors.NewE(err)
				}

				return nil
			},

			OnStop: func(context.Context) error {
				server.Stop()
				return nil
			},
		})
	}),
)
