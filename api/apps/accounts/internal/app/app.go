package app

import (
	"github.com/kloudlite/api/apps/accounts/internal/domain"
	"github.com/kloudlite/api/apps/accounts/internal/entities"
	"github.com/kloudlite/api/grpc-interfaces/kloudlite.io/rpc/accounts"
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

	fx.Provide(func(d domain.Domain) accounts.AccountsServer {
		return &accountsGrpcServer{d: d}
	}),

	domain.Module,
)
