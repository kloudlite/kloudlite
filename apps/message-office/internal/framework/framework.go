package framework

import (
	"go.uber.org/fx"
	"kloudlite.io/apps/message-office/internal/app"
	"kloudlite.io/apps/message-office/internal/env"
	rpc "kloudlite.io/pkg/grpc"
	httpServer "kloudlite.io/pkg/http-server"
	"kloudlite.io/pkg/redpanda"
	mongoDb "kloudlite.io/pkg/repos"
)

type fm struct {
	*env.Env
}

func (f *fm) GetBrokerHosts() string {
	return f.KafkaBrokers
}

func (f *fm) GetMongoConfig() (url string, dbName string) {
	return f.DbUri, f.DbName
}

func (f *fm) GetKafkaSASLAuth() *redpanda.KafkaSASLAuth {
	return nil
	// return &redpanda.KafkaSASLAuth{
	// 	SASLMechanism: redpanda.ScramSHA256,
	// 	User:          f.KafkaSaslUsername,
	// 	Password:      f.KafkaSaslPassword,
	// }
}

func (f *fm) GetHttpPort() uint16 {
	return f.HttpPort
}

func (f *fm) GetHttpCors() string {
	return ""
}

func (e *fm) GetGRPCPort() uint16 {
	return e.GrpcPort
}

var Module = fx.Module("framework",
	fx.Provide(func(ev *env.Env) *fm {
		return &fm{Env: ev}
	}),
	redpanda.NewProducerFx[*fm](),
	mongoDb.NewMongoClientFx[*fm](),
	app.Module,
	rpc.NewGrpcServerFx[*fm](),
	httpServer.NewHttpServerFx[*fm](),
)
