package framework

import (
	"context"
	"fmt"

	"github.com/kloudlite/operator/pkg/kubectl"
	"go.uber.org/fx"
	"k8s.io/client-go/rest"

	"kloudlite.io/apps/message-office/internal/app"
	"kloudlite.io/apps/message-office/internal/env"
	"kloudlite.io/pkg/grpc"
	httpServer "kloudlite.io/pkg/http-server"
	"kloudlite.io/pkg/logging"
	"kloudlite.io/pkg/redpanda"
	mongoDb "kloudlite.io/pkg/repos"
)

type fm struct {
	*env.Env
}

func (f *fm) GetBrokers() string {
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
	return e.ExternalGrpcPort
}

var Module = fx.Module("framework",
	fx.Provide(func(ev *env.Env) *fm {
		return &fm{Env: ev}
	}),
	redpanda.NewClientFx[*fm](),
	mongoDb.NewMongoClientFx[*fm](),

	fx.Provide(func(restCfg *rest.Config) (kubectl.YAMLClient, error) {
		return kubectl.NewYAMLClient(restCfg)
	}),

	fx.Provide(func(f *fm) (app.RealVectorGrpcClient, error) {
		return grpc.NewGrpcClient(f.VectorGrpcAddr)
	}),

	fx.Provide(func(f *fm) (app.InfraGrpcClient, error) {
		return grpc.NewGrpcClient(f.InfraGrpcAddr)
	}),

	app.Module,

	fx.Provide(func(logr logging.Logger) (app.InternalGrpcServer, error) {
		return grpc.NewGrpcServer(grpc.ServerOpts{
			Logger: logr.WithName("internal-grpc-server"),
		})
	}),

	fx.Invoke(func(lf fx.Lifecycle, server app.InternalGrpcServer, ev *env.Env) {
		lf.Append(fx.Hook{
			OnStart: func(context.Context) error {
				go server.Listen(fmt.Sprintf(":%d", ev.InternalGrpcPort))
				return nil
			},
			OnStop: func(context.Context) error {
				server.Stop()
				return nil
			},
		})
	}),

	fx.Provide(func(logr logging.Logger) (app.ExternalGrpcServer, error) {
		return grpc.NewGrpcServer(grpc.ServerOpts{
			Logger: logr.WithName("external-grpc-server"),
		})
	}),

	fx.Invoke(func(lf fx.Lifecycle, server app.ExternalGrpcServer, ev *env.Env) {
		lf.Append(fx.Hook{
			OnStart: func(context.Context) error {
				go server.Listen(fmt.Sprintf(":%d", ev.ExternalGrpcPort))
				return nil
			},
			OnStop: func(context.Context) error {
				server.Stop()
				return nil
			},
		})
	}),

	httpServer.NewHttpServerFx[*fm](),
)
