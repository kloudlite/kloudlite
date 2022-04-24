package framework

import (
	"go.uber.org/fx"
	"kloudlite.io/apps/infra/internal/application"
	"kloudlite.io/pkg/config"
	rpc "kloudlite.io/pkg/grpc"
	"kloudlite.io/pkg/logger"
	"kloudlite.io/pkg/messaging"
)

type Env struct {
	Port         uint16 `env:"GRPC_PORT" required:"true"`
	KafkaBrokers string `env:"KAFKA_BOOTSTRAP_SERVERS", required:"true"`
}

func (env *Env) GetGRPCPort() uint16 {
	return env.Port
}

var Module = fx.Module("framework",
	config.EnvFx[Env](),
	fx.Provide(logger.NewLogger),
	rpc.NewGrpcServerFx[*Env](),
	fx.Provide(func(env *Env) messaging.KafkaClient {
		return messaging.NewKafkaClient(env.KafkaBrokers)
	}),
	application.Module,
)
