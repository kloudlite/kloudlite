package app

import (
	"context"
	"log/slog"
	
	"github.com/kloudlite/api/apps/accounts/internal/app/grpc"
	"github.com/kloudlite/api/apps/accounts/internal/app/jwt"
	"github.com/kloudlite/api/apps/accounts/internal/domain"
	"github.com/kloudlite/api/apps/accounts/internal/entities"
	"github.com/kloudlite/api/apps/accounts/internal/env"
	"github.com/kloudlite/api/grpc-interfaces/kloudlite.io/rpc/accounts"
	"github.com/kloudlite/api/pkg/repos"
	"go.uber.org/fx"
	googleGrpc "google.golang.org/grpc"
)

var Module = fx.Module("app",
	fx.Module(
		"mongo-repos",
		repos.NewFxMongoRepo[*entities.Account]("accounts", "acc", entities.AccountIndices),
		repos.NewFxMongoRepo[*entities.Invitation]("invitations", "invite", entities.InvitationIndices),
		repos.NewFxMongoRepo[*entities.Team]("teams", "team", entities.TeamIndices),
		repos.NewFxMongoRepo[*entities.TeamMembership]("team-memberships", "tm", entities.TeamMembershipIndices),
		repos.NewFxMongoRepo[*entities.TeamApprovalRequest]("team-approval-requests", "tar", entities.TeamApprovalRequestIndices),
		repos.NewFxMongoRepo[*entities.PlatformSettings]("platform-settings", "ps", entities.PlatformSettingsIndices),
		repos.NewFxMongoRepo[*entities.PlatformInvitation]("platform-invitations", "pi", entities.PlatformInvitationIndices),
	),

	fx.Module(
		"jwt",
		fx.Provide(func(ev *env.Env) *jwt.JWTInterceptor {
			return jwt.NewJWTInterceptor(ev.JWTSecret)
		}),
	),

	fx.Module(
		"grpc-servers",
		fx.Provide(grpc.NewAccountsInternalServer),
		fx.Provide(grpc.NewServer),
		fx.Invoke(
			func(server *googleGrpc.Server, internalAccountsServerImpl accounts.AccountsInternalServer, accountsServer accounts.AccountsServer) error {
				accounts.RegisterAccountsServer(server, accountsServer)
				accounts.RegisterAccountsInternalServer(server, internalAccountsServerImpl)
				return nil
			},
		),
	),

	domain.Module,
	
	fx.Invoke(func(lifecycle fx.Lifecycle, d domain.Domain, logger *slog.Logger) {
		lifecycle.Append(fx.Hook{
			OnStart: func(ctx context.Context) error {
				// Initialize platform settings
				logger.Info("initializing platform settings")
				if err := d.InitializePlatform(ctx); err != nil {
					logger.Error("failed to initialize platform settings", "error", err)
					// Don't fail startup on initialization errors
				}
				return nil
			},
		})
	}),
)
