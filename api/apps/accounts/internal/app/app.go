package app

import (
	"github.com/kloudlite/api/apps/accounts/internal/domain"
	"github.com/kloudlite/api/apps/accounts/internal/entities"
	"github.com/kloudlite/api/grpc-interfaces/kloudlite.io/rpc/accounts"
	"github.com/kloudlite/api/grpc-interfaces/kloudlite.io/rpc/auth"
	"github.com/kloudlite/api/grpc-interfaces/kloudlite.io/rpc/comms"
	"github.com/kloudlite/api/grpc-interfaces/kloudlite.io/rpc/console"
	"github.com/kloudlite/api/grpc-interfaces/kloudlite.io/rpc/iam"
	"github.com/kloudlite/api/pkg/grpc"
	"github.com/kloudlite/api/pkg/kv"
	"github.com/kloudlite/api/pkg/repos"
	"go.uber.org/fx"
)

type AuthCacheClient kv.Client

type AuthClient grpc.Client

type ConsoleClient grpc.Client

type (
	ContainerRegistryClient grpc.Client
	CommsClient             grpc.Client
	IAMClient               grpc.Client
)

var Module = fx.Module("app",
	repos.NewFxMongoRepo[*entities.Account]("accounts", "acc", entities.AccountIndices),
	repos.NewFxMongoRepo[*entities.Invitation]("invitations", "invite", entities.InvitationIndices),

	fx.Provide(func(client AuthCacheClient) kv.Repo[*entities.Invitation] {
		return kv.NewRepo[*entities.Invitation](client)
	}),

	// grpc clients
	fx.Provide(func(conn ConsoleClient) console.ConsoleClient {
		return console.NewConsoleClient(conn)
	}),

	fx.Provide(func(conn IAMClient) iam.IAMClient {
		return iam.NewIAMClient(conn)
	}),

	fx.Provide(func(conn CommsClient) comms.CommsClient {
		return comms.NewCommsClient(conn)
	}),

	fx.Provide(func(conn AuthClient) auth.AuthClient {
		return auth.NewAuthClient(conn)
	}),

	fx.Provide(func(d domain.Domain) accounts.AccountsServer {
		return &accountsGrpcServer{d: d}
	}),

	fx.Invoke(func(d domain.Domain, gserver AccountsGrpcServer) {
		registerAccountsGRPCServer(gserver, d)
	}),

	domain.Module,
)
