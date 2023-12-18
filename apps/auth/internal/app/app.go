package app

import (
	"context"
	"github.com/kloudlite/api/pkg/nats"
	"go.uber.org/fx"
	"google.golang.org/grpc"

	"github.com/kloudlite/api/apps/auth/internal/app/graph"
	"github.com/kloudlite/api/apps/auth/internal/app/graph/generated"
	"github.com/kloudlite/api/apps/auth/internal/domain"
	"github.com/kloudlite/api/apps/auth/internal/env"
	"github.com/kloudlite/api/common"
	"github.com/kloudlite/api/constants"
	"github.com/kloudlite/api/grpc-interfaces/kloudlite.io/rpc/auth"
	"github.com/kloudlite/api/grpc-interfaces/kloudlite.io/rpc/comms"
	"github.com/kloudlite/api/pkg/cache"
	httpServer "github.com/kloudlite/api/pkg/http-server"
	"github.com/kloudlite/api/pkg/repos"
)

type CommsClientConnection *grpc.ClientConn

var Module = fx.Module(
	"app",
	repos.NewFxMongoRepo[*domain.User]("users", "usr", domain.UserIndexes),
	repos.NewFxMongoRepo[*domain.AccessToken]("access_tokens", "tkn", domain.AccessTokenIndexes),
	repos.NewFxMongoRepo[*domain.RemoteLogin]("remote_logins", "rlgn", domain.RemoteTokenIndexes),
	fx.Provide(
		func(ev *env.Env, jc *nats.JetstreamClient) (cache.Repo[*domain.VerifyToken], error) {
			cxt := context.TODO()
			return cache.NewNatsKVRepo[*domain.VerifyToken](cxt, ev.VerifyTokenKVBucket, jc)
		},
	),
	fx.Provide(
		func(ev *env.Env, jc *nats.JetstreamClient) (cache.Repo[*domain.ResetPasswordToken], error) {
			cxt := context.TODO()
			return cache.NewNatsKVRepo[*domain.ResetPasswordToken](cxt, ev.ResetPasswordTokenKVBucket, jc)
		},
	),

	fx.Provide(
		func(conn CommsClientConnection) comms.CommsClient {
			return comms.NewCommsClient((*grpc.ClientConn)(conn))
		},
	),

	fx.Provide(fxGithub),
	fx.Provide(fxGitlab),
	fx.Provide(fxGoogle),

	fx.Provide(fxRPCServer),
	fx.Invoke(
		func(server *grpc.Server, authServer auth.AuthServer) {
			auth.RegisterAuthServer(server, authServer)
		},
	),

	fx.Invoke(
		func(
			server httpServer.Server,
			d domain.Domain,
			ev *env.Env,
			repo cache.Repo[*common.AuthSession],
		) {
			schema := generated.NewExecutableSchema(
				generated.Config{Resolvers: graph.NewResolver(d, ev)},
			)

			server.SetupGraphqlServer(
				schema,
				httpServer.NewSessionMiddleware(
					repo,
					constants.CookieName,
					ev.CookieDomain,
					constants.CacheSessionPrefix,
				),
			)
		},
	),

	domain.Module,
)
