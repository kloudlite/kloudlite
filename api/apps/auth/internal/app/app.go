package app

import (
	"context"

	recaptchaenterprise "cloud.google.com/go/recaptchaenterprise/v2/apiv1"

	"github.com/kloudlite/api/apps/auth/internal/app/grpcv2"
	"github.com/kloudlite/api/apps/auth/internal/domain"
	"github.com/kloudlite/api/apps/auth/internal/entities"
	"github.com/kloudlite/api/apps/auth/internal/env"
	"github.com/kloudlite/api/grpc-interfaces/kloudlite.io/rpc/auth"
	authV2 "github.com/kloudlite/api/grpc-interfaces/kloudlite.io/rpc/auth/v2"
	"github.com/kloudlite/api/grpc-interfaces/kloudlite.io/rpc/comms"
	"github.com/kloudlite/api/pkg/kv"
	"github.com/kloudlite/api/pkg/nats"
	"github.com/kloudlite/api/pkg/repos"
	"go.uber.org/fx"
	"google.golang.org/grpc"
)

type CommsClientConnection *grpc.ClientConn

type AuthGrpcServer *grpc.Server
type AuthGrpcServerV2 *grpc.Server

var Module = fx.Module(
	"app",
	repos.NewFxMongoRepo[*entities.User]("users", "usr", entities.UserIndexes),
	repos.NewFxMongoRepo[*entities.AccessToken]("access_tokens", "tkn", entities.AccessTokenIndexes),
	repos.NewFxMongoRepo[*entities.RemoteLogin]("remote_logins", "rlgn", entities.RemoteTokenIndexes),
	repos.NewFxMongoRepo[*entities.InviteCode]("invite_codes", "invcode", entities.InviteCodeIndexes),
	fx.Provide(
		func(ev *env.Env, jc *nats.JetstreamClient) (kv.Repo[*entities.VerifyToken], error) {
			cxt := context.TODO()
			return kv.NewNatsKVRepo[*entities.VerifyToken](cxt, ev.VerifyTokenKVBucket, jc)
		},
	),
	fx.Provide(
		func(ev *env.Env, jc *nats.JetstreamClient) (kv.Repo[*entities.ResetPasswordToken], error) {
			cxt := context.TODO()
			return kv.NewNatsKVRepo[*entities.ResetPasswordToken](cxt, ev.ResetPasswordTokenKVBucket, jc)
		},
	),

	fx.Provide(
		func(conn CommsClientConnection) comms.CommsClient {
			return comms.NewCommsClient((*grpc.ClientConn)(conn))
		},
	),

	fx.Provide(
		func(ev *env.Env) (*recaptchaenterprise.Client, error) {
			if ev.GoogleRecaptchaEnabled {
				client, err := recaptchaenterprise.NewClient(context.TODO())
				if err != nil {
					return nil, err
				}
				return client, nil
			}
			return nil, nil
		},
	),

	fx.Provide(fxGithub),
	fx.Provide(fxGitlab),
	fx.Provide(fxGoogle),

	fx.Provide(fxRPCServer),
	fx.Provide(grpcv2.NewServer),

	fx.Invoke(
		func(server AuthGrpcServer, authServer auth.AuthServer) {
			auth.RegisterAuthServer((*grpc.Server)(server), authServer)
		},
	),

	fx.Invoke(
		func(server AuthGrpcServerV2, authServer authV2.AuthV2Server) {
			authV2.RegisterAuthV2Server((*grpc.Server)(server), authServer)
		},
	),

	domain.Module,
)
