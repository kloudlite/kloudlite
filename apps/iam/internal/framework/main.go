package framework

import (
	"go.uber.org/fx"
	"kloudlite.io/apps/iam/internal/application"
	"kloudlite.io/apps/iam/internal/env"
	"kloudlite.io/pkg/cache"
	rpc "kloudlite.io/pkg/grpc"
	"kloudlite.io/pkg/repos"
)

type fm struct {
	*env.Env
}

func (f *fm) RedisOptions() (hosts, username, password, basePrefix string) {
	return f.RedisHosts, f.RedisUsername, f.RedisPassword, f.RedisPrefix
}

func (f *fm) GetMongoConfig() (url, dbName string) {
	return f.MongoDbUri, f.MongoDbName
}

func (f *fm) GetGRPCPort() uint16 {
	return f.GrpcPort
}

var Module fx.Option = fx.Module(
	"framework",
	fx.Provide(func(ev *env.Env) *fm {
		return &fm{Env: ev}
	}),
	repos.NewMongoClientFx[*fm](),
	cache.NewRedisFx[*fm](),
	rpc.NewGrpcServerFx[*fm](),
	application.Module,
)
