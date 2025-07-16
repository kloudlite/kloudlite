package app

import (
	"context"
	"github.com/kloudlite/api/apps/auth/internal/app/grpc"
	"github.com/kloudlite/api/apps/auth/internal/domain"
	"github.com/kloudlite/api/apps/auth/internal/entities"
	"github.com/kloudlite/api/apps/auth/internal/env"
	auth_rpc "github.com/kloudlite/api/grpc-interfaces/kloudlite.io/rpc/auth"
	"github.com/kloudlite/api/pkg/kv"
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
		repos.NewFxMongoRepo[*entities.AccessToken]("access_tokens", "tkn", entities.AccessTokenIndexes),
		repos.NewFxMongoRepo[*entities.RemoteLogin]("remote_logins", "rlgn", entities.RemoteTokenIndexes),
		repos.NewFxMongoRepo[*entities.InviteCode]("invite_codes", "invcode", entities.InviteCodeIndexes),
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
		"grpc-servers",
		fx.Provide(grpc.NewInternalServer),
		fx.Provide(grpc.NewServer),
		fx.Invoke(func(server *googleGrpc.Server, internalAuthServerImpl auth_rpc.AuthInternalServer, authServer auth_rpc.AuthServer) {
			auth_rpc.RegisterAuthInternalServer(server, internalAuthServerImpl)
			auth_rpc.RegisterAuthServer(server, authServer)
		}),
	),

	domain.Module,
)
