package framework

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/kloudlite/api/apps/observability/internal/app"
	"github.com/kloudlite/api/apps/observability/internal/env"
	"github.com/kloudlite/api/common"
	"github.com/kloudlite/api/pkg/grpc"
	"github.com/kloudlite/api/pkg/kv"
	"github.com/kloudlite/api/pkg/nats"
	"go.uber.org/fx"
)

var Module = fx.Module("framework",
	fx.Provide(func(ev *env.Env) (app.IAMGrpcClient, error) {
		return grpc.NewGrpcClient(ev.IAMGrpcAddr)
	}),

	fx.Provide(func(ev *env.Env) (app.InfraClient, error) {
		return grpc.NewGrpcClient(ev.InfraGrpcAddr)
	}),

	fx.Provide(func(ev *env.Env, logger *slog.Logger) (*nats.Client, error) {
		return nats.NewClient(ev.NatsURL, nats.ClientOpts{
			Name:   "observability-api",
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

	fx.Provide(
		func(ev *env.Env, jc *nats.JetstreamClient) (kv.Repo[*common.AuthSession], error) {
			cxt := context.TODO()
			return kv.NewNatsKVRepo[*common.AuthSession](cxt, ev.SessionKVBucket, jc)
		},
	),

	app.Module,

	fx.Provide(func() *http.ServeMux {
		return http.NewServeMux()
	}),

	fx.Invoke(func(lf fx.Lifecycle, ev *env.Env, mux *http.ServeMux, logger *slog.Logger) {
		mux.HandleFunc("/_healthy", func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		})

		logger.Info("starting observability api HTTP server on", "port", ev.HttpPort)

		server := &http.Server{Addr: fmt.Sprintf(":%d", ev.HttpPort), Handler: mux}
		lf.Append(fx.Hook{
			OnStart: func(context.Context) error {
				go server.ListenAndServe()
				return nil
			},
			OnStop: func(ctx context.Context) error {
				return server.Shutdown(ctx)
			},
		})
	}),
)
