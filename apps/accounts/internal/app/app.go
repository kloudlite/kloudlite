package app

import (
	"context"

	"github.com/99designs/gqlgen/graphql"
	"github.com/gofiber/fiber/v2"
	"go.uber.org/fx"
	"kloudlite.io/apps/accounts/internal/app/graph"
	"kloudlite.io/apps/accounts/internal/app/graph/generated"
	"kloudlite.io/apps/accounts/internal/domain"
	"kloudlite.io/apps/accounts/internal/entities"
	"kloudlite.io/apps/accounts/internal/env"
	"kloudlite.io/common"
	"kloudlite.io/constants"
	"kloudlite.io/pkg/cache"
	"kloudlite.io/pkg/repos"

	rpcAccounts "kloudlite.io/grpc-interfaces/kloudlite.io/rpc/accounts"
	httpServer "kloudlite.io/pkg/http-server"
)

type AuthCacheClient cache.Client

var Module = fx.Module("app",
	repos.NewFxMongoRepo[*entities.Account]("accounts", "acc", entities.AccountIndices),
	repos.NewFxMongoRepo[*entities.Invitation]("invitations", "invite", entities.InvitationIndices),
	cache.NewFxRepo[*entities.Invitation](),

	// grpc clients
	ConsoleClientFx,
	ContainerRegistryFx,
	IAMClientFx,
	CommsClientFx,
	AuthClientFx,

	fx.Invoke(
		func(server *fiber.App, d domain.Domain, env *env.Env, cacheClient AuthCacheClient) {
			gqlConfig := generated.Config{Resolvers: graph.NewResolver(d)}

			gqlConfig.Directives.IsLoggedIn = func(ctx context.Context, obj interface{}, next graphql.Resolver) (res interface{}, err error) {
				sess := httpServer.GetSession[*common.AuthSession](ctx)
				if sess == nil {
					return nil, fiber.ErrUnauthorized
				}

				cc := domain.AccountsContext{Context: ctx, UserId: sess.UserId}
				return next(context.WithValue(ctx, "kl-accounts-ctx", cc))
			}

			schema := generated.NewExecutableSchema(gqlConfig)
			httpServer.SetupGQLServer(
				server,
				schema,
				httpServer.NewSessionMiddleware[*common.AuthSession](
					cacheClient,
					constants.CookieName,
					env.CookieDomain,
					constants.CacheSessionPrefix,
				),
			)
		},
	),

	// func NewGrpcServerFx[T ServerOptions]() fx.Option {
	// 	return fx.Module(
	// 		"grpc-server",
	// 		fx.Provide(func(logger logging.Logger) *grpc.Server {
	// 			return grpc.NewServer(
	// 				grpc.StreamInterceptor(func(srv interface{}, stream grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
	// 					p, ok := peer.FromContext(stream.Context())
	// 					if ok {
	// 						logger.Debugf("[Stream] New connection from %s", p.Addr.String())
	// 					}
	// 					return handler(srv, stream)
	// 				}),
	// 			)
	// 		}),
	// 		fx.Invoke(
	// 			func(lf fx.Lifecycle, env T, server *grpc.Server, logger logging.Logger) {
	// 				lf.Append(
	// 					fx.Hook{
	// 						OnStart: func(ctx context.Context) error {
	// 							listen, err := net.Listen("tcp", fmt.Sprintf(":%d", env.GetGRPCPort()))
	// 							defer func() {
	// 								logger.Infof("[GRPC Server] started on port (%d)", env.GetGRPCPort())
	// 							}()
	// 							if err != nil {
	// 								return errors.NewEf(err, "could not listen to net/tcp server")
	// 							}
	// 							go func() error {
	// 								err := server.Serve(listen)
	// 								if err != nil {
	// 									return errors.NewEf(err, "could not start grpc server ")
	// 								}
	// 								return nil
	// 							}()
	// 							return nil
	// 						},
	// 						OnStop: func(context.Context) error {
	// 							server.Stop()
	// 							return nil
	// 						},
	// 					},
	// 				)
	// 			},
	// 		),
	// 	)
	// }

	fx.Provide(func(d domain.Domain) rpcAccounts.AccountsServer {
		return &accountsGrpcServer{d: d}
	}),

	domain.Module,
)
