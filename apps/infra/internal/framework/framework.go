package framework

import (
	"go.uber.org/fx"
	"kloudlite.io/apps/infra/internal/app"
	"kloudlite.io/apps/infra/internal/env"
	"kloudlite.io/pkg/cache"
	rpc "kloudlite.io/pkg/grpc"
	httpServer "kloudlite.io/pkg/http-server"
	"kloudlite.io/pkg/redpanda"
	mongoRepo "kloudlite.io/pkg/repos"
)

type framework struct {
	*env.Env
}

func (f *framework) GetBrokers() (brokers string) {
	return f.Env.KafkaBrokers
}

func (f *framework) GetKafkaSASLAuth() *redpanda.KafkaSASLAuth {
	return &redpanda.KafkaSASLAuth{
		SASLMechanism: redpanda.ScramSHA256,
		User:          f.Env.KafkaUsername,
		Password:      f.Env.KafkaPassword,
	}
}

func (f *framework) GetGRPCServerURL() string {
	return f.FinanceGrpcAddr
}

func (f *framework) GetHttpCors() string {
	return "https://studio.apollographql.com"
}

func (f *framework) GetHttpPort() uint16 {
	return f.HttpPort
}

func (f *framework) GetMongoConfig() (url string, dbName string) {
	return f.InfraDbUri, f.InfraDbName
}

var Module = fx.Module("framework",
	fx.Provide(func(ev *env.Env) *framework {
		return &framework{Env: ev}
	}),

	mongoRepo.NewMongoClientFx[*framework](),
	httpServer.NewHttpServerFx[*framework](),

	redpanda.NewClientFx[*framework](),

	fx.Provide(
		func(f *framework) app.AuthCacheClient {
			return cache.NewRedisClient(f.AuthRedisHosts, f.AuthRedisUserName, f.AuthRedisPassword, f.AuthRedisPrefix)
		},
	),
	rpc.NewGrpcClientFx[*framework, app.FinanceClientConnection](),
	cache.FxLifeCycle[app.AuthCacheClient](),
	app.Module,
)
