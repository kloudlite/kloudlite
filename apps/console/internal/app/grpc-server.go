package app

import (
	"kloudlite.io/apps/console/internal/domain"
	"kloudlite.io/grpc-interfaces/kloudlite.io/rpc/console"
)

type consoleServerI struct {
	console.UnimplementedConsoleServer
	d domain.Domain
}

func fxConsoleGrpcServer(d domain.Domain) console.ConsoleServer {
	return &consoleServerI{
		d: d,
	}
}
