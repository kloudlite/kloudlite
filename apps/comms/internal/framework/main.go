package framework

import (
	"go.uber.org/fx"
	"kloudlite.io/apps/comms/internal/app"
	"kloudlite.io/pkg/config"
	rpc "kloudlite.io/pkg/grpc"
	"kloudlite.io/pkg/mail"
)

type Env struct {
	SendGridKey string `env:"SENDGRID_API_KEY" required:"true"`
	GrpcPort    uint16 `env:"GRPC_PORT" required:"true"`
}

func (e *Env) GetSendGridApiKey() string {
	return e.SendGridKey
}

func (e *Env) GetGRPCPort() uint16 {
	return e.GrpcPort
}

var Module = fx.Module(
	"framework",
	config.EnvFx[Env](),
	rpc.NewGrpcServerFx[*Env](),
	mail.NewSendGridMailerFx[*Env](),
	app.Module,
)
