package framework

import (
	"context"
	"fmt"

	"go.uber.org/fx"

	"github.com/kloudlite/api/apps/auth/internal/app"
	"github.com/kloudlite/api/apps/auth/internal/env"
	"github.com/kloudlite/api/pkg/cache"
	rpc "github.com/kloudlite/api/pkg/grpc"
	httpServer "github.com/kloudlite/api/pkg/http-server"
	"github.com/kloudlite/api/pkg/logging"
	"github.com/kloudlite/api/pkg/repos"
)

type CommsGrpcEnv struct {
	*env.Env
}

func (c CommsGrpcEnv) GetGRPCServerURL() string {
	return c.CommsService
}

type fm struct {
	*env.Env
}

func (e *fm) GetHttpPort() uint16 {
	return e.Port
}

func (e *fm) GetHttpCors() string {
	return e.CorsOrigins
}

func (e *fm) RedisOptions() (hosts, username, password, basePrefix string) {
	return e.RedisHosts, e.RedisUserName, e.RedisPassword, e.RedisPrefix
}

func (e *fm) GetMongoConfig() (url string, dbName string) {
	return e.MongoUri, e.MongoDbName
}

func (e *fm) GetGRPCPort() uint16 {
	return e.GrpcPort
}

var Module fx.Option = fx.Module(
	"framework",
	fx.Provide(func(ev *env.Env) *fm {
		return &fm{ev}
	}),
	fx.Provide(func(ev *env.Env) *CommsGrpcEnv {
		return &CommsGrpcEnv{ev}
	}),
	repos.NewMongoClientFx[*fm](),
	cache.NewRedisFx[*fm](),

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

	rpc.NewGrpcServerFx[*fm](),
	rpc.NewGrpcClientFx[*CommsGrpcEnv, app.CommsClientConnection](),
	app.Module,
)
