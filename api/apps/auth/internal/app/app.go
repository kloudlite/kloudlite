package app

import (
	"context"
	"log/slog"
	"github.com/kloudlite/api/apps/auth/internal/app/email"
	"github.com/kloudlite/api/apps/auth/internal/app/grpc"
	"github.com/kloudlite/api/apps/auth/internal/domain"
	"github.com/kloudlite/api/apps/auth/internal/entities"
	"github.com/kloudlite/api/apps/auth/internal/env"
	auth_rpc "github.com/kloudlite/api/grpc-interfaces/kloudlite.io/rpc/auth"
	"github.com/kloudlite/api/pkg/kv"
	"github.com/kloudlite/api/pkg/mail"
	"github.com/kloudlite/api/pkg/nats"
	"github.com/kloudlite/api/pkg/repos"
	"go.uber.org/fx"
	googleGrpc "google.golang.org/grpc"
)

var Module = fx.Module(
	"app",

	fx.Module(
		"mongo-repos",
		repos.NewFxMongoRepo[*entities.User]("users", "usr", entities.UserIndexes),
		repos.NewFxMongoRepo[*entities.DeviceFlow]("device_flows", "devflow", entities.DeviceFlowIndexes),
		repos.NewFxMongoRepo[*entities.PlatformUser]("platform-users", "pu", entities.PlatformUserIndices),
		repos.NewFxMongoRepo[*entities.Notification]("notifications", "notif", entities.NotificationIndices),
	),

	fx.Module(
		"kv-repos",
		fx.Provide(
			func(ev *env.AuthEnv, jc *nats.JetstreamClient) (kv.Repo[*entities.VerifyToken], error) {
				cxt := context.TODO()
				return kv.NewNatsKVRepo[*entities.VerifyToken](cxt, ev.VerifyTokenKVBucket, jc)
			},
		),
		fx.Provide(
			func(ev *env.AuthEnv, jc *nats.JetstreamClient) (kv.Repo[*entities.ResetPasswordToken], error) {
				cxt := context.TODO()
				return kv.NewNatsKVRepo[*entities.ResetPasswordToken](cxt, ev.ResetPasswordTokenKVBucket, jc)
			},
		),
	),

	fx.Module(
		"email-service",
		fx.Provide(func(ev *env.AuthEnv) (mail.Mailer, error) {
			return mail.NewMailtrapMailer(ev.MailtrapApiToken, ev.SupportEmail), nil
		}),
		fx.Provide(func(mailer mail.Mailer, ev *env.AuthEnv) (*email.EmailService, error) {
			return email.NewEmailService(mailer, ev.SupportEmail, ev.WebUrl)
		}),
	),

	fx.Module(
		"grpc-servers",
		fx.Provide(grpc.NewInternalServer),
		fx.Provide(grpc.NewServer),
		fx.Invoke(func(server *googleGrpc.Server, internalAuthServerImpl auth_rpc.AuthInternalServer, authServer auth_rpc.AuthServer) {
			auth_rpc.RegisterAuthInternalServer(server, internalAuthServerImpl)
			auth_rpc.RegisterAuthServer(server, authServer)
		}),
	),

	domain.Module,
	
	fx.Invoke(func(lifecycle fx.Lifecycle, d domain.Domain, ev *env.AuthEnv, logger *slog.Logger) {
		lifecycle.Append(fx.Hook{
			OnStart: func(ctx context.Context) error {
				// Initialize platform if owner email is provided
				if ev.PlatformOwnerEmail != "" {
					logger.Info("initializing platform", "ownerEmail", ev.PlatformOwnerEmail)
					if err := d.InitializePlatform(ctx, ev.PlatformOwnerEmail); err != nil {
						logger.Error("failed to initialize platform", "error", err)
						// Don't fail startup on initialization errors
					}
				}
				return nil
			},
		})
	}),
)
