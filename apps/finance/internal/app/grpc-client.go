package app

import (
	"go.uber.org/fx"
	"google.golang.org/grpc"
	"kloudlite.io/grpc-interfaces/kloudlite.io/rpc/auth"
	"kloudlite.io/grpc-interfaces/kloudlite.io/rpc/ci"
	"kloudlite.io/grpc-interfaces/kloudlite.io/rpc/comms"
	"kloudlite.io/grpc-interfaces/kloudlite.io/rpc/console"
	"kloudlite.io/grpc-interfaces/kloudlite.io/rpc/iam"
)

type CIGrpcClientConn *grpc.ClientConn
type AuthGrpcClientConn *grpc.ClientConn
type ConsoleClientConnection *grpc.ClientConn
type CommsClientConnection *grpc.ClientConn
type IAMClientConnection *grpc.ClientConn

var ConsoleClientFx = fx.Provide(func(conn ConsoleClientConnection) console.ConsoleClient {
	return console.NewConsoleClient((*grpc.ClientConn)(conn))
})

var IAMClientFx = fx.Provide(func(conn IAMClientConnection) iam.IAMClient {
	return iam.NewIAMClient((*grpc.ClientConn)(conn))
})

var CommsClientFx = fx.Provide(func(conn CommsClientConnection) comms.CommsClient {
	return comms.NewCommsClient((*grpc.ClientConn)(conn))
})

var CiClientFx = fx.Provide(
	func(conn CIGrpcClientConn) ci.CIClient {
		return ci.NewCIClient((*grpc.ClientConn)(conn))
	},
)

var AuthClientFx = fx.Provide(
	func(conn AuthGrpcClientConn) auth.AuthClient {
		return auth.NewAuthClient((*grpc.ClientConn)(conn))
	},
)
