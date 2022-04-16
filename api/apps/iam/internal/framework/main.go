package framework

import (
	"context"

	"go.uber.org/fx"
	"google.golang.org/grpc"
	"kloudlite.io/apps/iam/internal/application"
	"kloudlite.io/pkg/cache"
	"kloudlite.io/pkg/config"
	rpc "kloudlite.io/pkg/grpc"
	"kloudlite.io/pkg/repos"
)

type Env struct {
	Port        int    `env:"GRPC_PORT" required:"true"`
	MongoDbUri  string `env:"MONGO_DB_URI" required:"true"`
	MongoDbName string `env:"MONGO_DB_NAME" required:"true"`
	RedisHosts  string `env:"REDIS_HOSTS" required:"true"`
}

func (env *Env) RedisOptions() (hosts string, username string, password string) {
	return env.RedisHosts, "", ""
}

func (env *Env) GetMongoConfig() (url, dbName string) {
	return env.MongoDbUri, env.MongoDbName
}

var Module = fx.Module("framework",
	fx.Provide(config.LoadEnv[Env]()),
	fx.Provide(grpc.NewServer),
	repos.NewMongoClientFx[*Env](),

	application.Module,
	cache.NewRedisFx[*Env](),
	fx.Invoke(func(lf fx.Lifecycle, env *Env, server *grpc.Server) {
		lf.Append(fx.Hook{
			OnStart: func(ctx context.Context) error {
				return rpc.GRPCStartServer(ctx, server, env.Port)
			},
			OnStop: func(context.Context) error {
				server.Stop()
				return nil
			},
		})
	}),
)
