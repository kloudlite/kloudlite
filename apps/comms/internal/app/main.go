package app

import (
	"go.uber.org/fx"
	"google.golang.org/grpc"
	"kloudlite.io/grpc-interfaces/kloudlite.io/rpc/comms"
	"kloudlite.io/pkg/config"
)

type Env struct {
	SupportEmail string `env:"SUPPORT_EMAIL"`
}

var Module = fx.Module("app",
	config.EnvFx[Env](),
	fx.Provide(fxRPCServer),
	fx.Invoke(func(server *grpc.Server, commsServer comms.CommsServer) {
		comms.RegisterCommsServer(server, commsServer)
	}),
)
