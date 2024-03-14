package app

import (
	"context"
	"github.com/kloudlite/api/apps/auth/internal/entities"

	"github.com/99designs/gqlgen/graphql"
	"github.com/gofiber/fiber/v2"
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
	httpServer "github.com/kloudlite/api/pkg/http-server"
	"github.com/kloudlite/api/pkg/kv"
	"github.com/kloudlite/api/pkg/repos"
)

type CommsClientConnection *grpc.ClientConn

var Module = fx.Module(
	"app",
	repos.NewFxMongoRepo[*entities.User]("users", "usr", entities.UserIndexes),
	repos.NewFxMongoRepo[*entities.AccessToken]("access_tokens", "tkn", entities.AccessTokenIndexes),
	repos.NewFxMongoRepo[*entities.RemoteLogin]("remote_logins", "rlgn", entities.RemoteTokenIndexes),
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
			repo kv.Repo[*common.AuthSession],
		) {
			gqlConfig := generated.Config{Resolvers: graph.NewResolver(d, ev)}
			gqlConfig.Directives.IsLoggedIn = func(ctx context.Context, obj any, next graphql.Resolver) (res interface{}, err error) {
				sess := httpServer.GetSession[*common.AuthSession](ctx)
				if sess == nil {
					return nil, fiber.ErrUnauthorized
				}

				return next(context.WithValue(ctx, "user-session", sess))
			}

			gqlConfig.Directives.IsLoggedInAndVerified = func(ctx context.Context, obj any, next graphql.Resolver) (res interface{}, err error) {
				sess := httpServer.GetSession[*common.AuthSession](ctx)
				if sess == nil {
					return nil, fiber.ErrUnauthorized
				}

				if !sess.UserVerified {
					return nil, &fiber.Error{
						Code:    fiber.StatusForbidden,
						Message: "user's email is not verified",
					}
				}

				return next(context.WithValue(ctx, "user-session", sess))
			}

			schema := generated.NewExecutableSchema(gqlConfig)

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
