package app

import (
	"github.com/kloudlite/api/apps/accounts/internal/app/grpc"
	"github.com/kloudlite/api/apps/accounts/internal/domain"
	"github.com/kloudlite/api/apps/accounts/internal/entities"
	"github.com/kloudlite/api/grpc-interfaces/kloudlite.io/rpc/accounts"
	"github.com/kloudlite/api/pkg/repos"
	"go.uber.org/fx"
	googleGrpc "google.golang.org/grpc"
)

var Module = fx.Module("app",
	repos.NewFxMongoRepo[*entities.Account]("accounts", "acc", entities.AccountIndices),
	repos.NewFxMongoRepo[*entities.Invitation]("invitations", "invite", entities.InvitationIndices),

	fx.Module(
		"grpc-servers",
		fx.Provide(grpc.NewAccountsInternalServer),
		fx.Provide(grpc.NewServer),
		fx.Invoke(
			func(lifecycle fx.Lifecycle, server *googleGrpc.Server, internalAccountsServerImpl accounts.AccountsInternalServer, accountsServer accounts.AccountsServer) {
				accounts.RegisterAccountsServer(server, accountsServer)
				accounts.RegisterAccountsInternalServer(server, internalAccountsServerImpl)
			},
		),
	),

	domain.Module,
)
