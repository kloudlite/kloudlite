package app

import (
	"go.uber.org/fx"
	"kloudlite.io/grpc-interfaces/kloudlite.io/rpc/auth"
	"kloudlite.io/grpc-interfaces/kloudlite.io/rpc/comms"
	"kloudlite.io/grpc-interfaces/kloudlite.io/rpc/console"
	"kloudlite.io/grpc-interfaces/kloudlite.io/rpc/container_registry"
	"kloudlite.io/grpc-interfaces/kloudlite.io/rpc/iam"
	"kloudlite.io/pkg/grpc"
)

// type AuthClient *grpc.ClientConn
type AuthClient grpc.Client

// type ConsoleClientConn *grpc.ClientConn
type ConsoleClient grpc.Client

// type ContainerRegistryClientConn *grpc.ClientConn
type ContainerRegistryClient grpc.Client

// type CommsClientConn *grpc.ClientConn
type CommsClient grpc.Client

// type IAMClient *grpc.ClientConn
type IAMClient grpc.Client

var ConsoleClientFx = fx.Provide(func(conn ConsoleClient) console.ConsoleClient {
	return console.NewConsoleClient(conn)
})

var ContainerRegistryFx = fx.Provide(func(conn ContainerRegistryClient) container_registry.ContainerRegistryClient {
	return container_registry.NewContainerRegistryClient(conn)
})

var IAMClientFx = fx.Provide(func(conn IAMClient) iam.IAMClient {
	return iam.NewIAMClient(conn)
})

var CommsClientFx = fx.Provide(func(conn CommsClient) comms.CommsClient {
	return comms.NewCommsClient(conn)
})

var AuthClientFx = fx.Provide(
	func(conn AuthClient) auth.AuthClient {
		return auth.NewAuthClient(conn)
	},
)
