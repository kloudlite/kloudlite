package app

import (
	"go.uber.org/fx"
	"google.golang.org/grpc"
	"kloudlite.io/grpc-interfaces/kloudlite.io/rpc/auth"
	"kloudlite.io/grpc-interfaces/kloudlite.io/rpc/comms"
	"kloudlite.io/grpc-interfaces/kloudlite.io/rpc/console"
	"kloudlite.io/grpc-interfaces/kloudlite.io/rpc/container_registry"
	"kloudlite.io/grpc-interfaces/kloudlite.io/rpc/iam"
)

type AuthGrpcClientConn *grpc.ClientConn
type ConsoleClientConnection *grpc.ClientConn
type ContainerRegistryClientConnection *grpc.ClientConn
type CommsClientConnection *grpc.ClientConn
type IAMClientConnection *grpc.ClientConn

var ConsoleClientFx = fx.Provide(func(conn ConsoleClientConnection) console.ConsoleClient {
	return console.NewConsoleClient((*grpc.ClientConn)(conn))
})

var ContainerRegistryFx = fx.Provide(func(conn ConsoleClientConnection) container_registry.ContainerRegistryClient {
	return container_registry.NewContainerRegistryClient((*grpc.ClientConn)(conn))
})

var IAMClientFx = fx.Provide(func(conn IAMClientConnection) iam.IAMClient {
	return iam.NewIAMClient((*grpc.ClientConn)(conn))
})

var CommsClientFx = fx.Provide(func(conn CommsClientConnection) comms.CommsClient {
	return comms.NewCommsClient((*grpc.ClientConn)(conn))
})

var AuthClientFx = fx.Provide(
	func(conn AuthGrpcClientConn) auth.AuthClient {
		return auth.NewAuthClient((*grpc.ClientConn)(conn))
	},
)
