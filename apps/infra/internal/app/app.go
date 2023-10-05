package app

import (
	"context"
	"fmt"

	"kloudlite.io/apps/infra/internal/entities"

	"github.com/99designs/gqlgen/graphql"
	"github.com/gofiber/fiber/v2"
	"go.uber.org/fx"
	"kloudlite.io/apps/infra/internal/app/graph"
	"kloudlite.io/apps/infra/internal/app/graph/generated"
	"kloudlite.io/apps/infra/internal/domain"
	"kloudlite.io/apps/infra/internal/env"
	"kloudlite.io/common"
	"kloudlite.io/constants"
	"kloudlite.io/grpc-interfaces/kloudlite.io/rpc/accounts"
	"kloudlite.io/grpc-interfaces/kloudlite.io/rpc/iam"
	"kloudlite.io/grpc-interfaces/kloudlite.io/rpc/infra"
	message_office_internal "kloudlite.io/grpc-interfaces/kloudlite.io/rpc/message-office-internal"
	"kloudlite.io/pkg/cache"
	"kloudlite.io/pkg/grpc"
	httpServer "kloudlite.io/pkg/http-server"
	"kloudlite.io/pkg/logging"
	"kloudlite.io/pkg/redpanda"
	"kloudlite.io/pkg/repos"

	fWebsocket "github.com/gofiber/websocket/v2"
)

type AuthCacheClient cache.Client

type IAMGrpcClient grpc.Client
type AccountGrpcClient grpc.Client
type MessageOfficeInternalGrpcClient grpc.Client

type InfraGrpcServer grpc.Server

var Module = fx.Module(
	"app",
	repos.NewFxMongoRepo[*entities.Cluster]("clusters", "clus", entities.ClusterIndices),
	repos.NewFxMongoRepo[*entities.BYOCCluster]("byoc_clusters", "byoc", entities.BYOCClusterIndices),
	repos.NewFxMongoRepo[*entities.DomainEntry]("domain_entries", "de", entities.DomainEntryIndices),
	repos.NewFxMongoRepo[*entities.NodePool]("node_pools", "npool", entities.NodePoolIndices),
	repos.NewFxMongoRepo[*entities.Node]("node", "node", entities.NodePoolIndices),
	repos.NewFxMongoRepo[*entities.CloudProviderSecret]("secrets", "scrt", entities.SecretIndices),

	fx.Provide(
		func(conn IAMGrpcClient) iam.IAMClient {
			return iam.NewIAMClient(conn)
		},
	),

	fx.Provide(func(conn AccountGrpcClient) accounts.AccountsClient {
		return accounts.NewAccountsClient(conn)
	}),

	fx.Provide(func(gclient MessageOfficeInternalGrpcClient) message_office_internal.MessageOfficeInternalClient {
		return message_office_internal.NewMessageOfficeInternalClient(gclient)
	}),

	redpanda.NewProducerFx[redpanda.Client](),

	domain.Module,

	// fx.Provide(func(cli redpanda.Client, ev *env.Env, logger logging.Logger) (ByocClientUpdatesConsumer, error) {
	// 	return redpanda.NewConsumer(cli.GetBrokerHosts(), ev.KafkaConsumerGroupId, redpanda.ConsumerOpts{
	// 		SASLAuth: cli.GetKafkaSASLAuth(),
	// 		Logger:   logger.WithName("byoc-client-updates"),
	// 	}, []string{ev.KafkaTopicByocClientUpdates})
	// }),
	//
	// fx.Invoke(processByocClientUpdates),

	fx.Provide(func(d domain.Domain) infra.InfraServer {
		return newGrpcServer(d)
	}),

  fx.Invoke(func(gserver InfraGrpcServer, srv infra.InfraServer) {
    infra.RegisterInfraServer(gserver, srv)
  }),

	fx.Provide(func(cli redpanda.Client, ev *env.Env, logger logging.Logger) (InfraUpdatesConsumer, error) {
		return redpanda.NewConsumer(cli.GetBrokerHosts(), ev.KafkaConsumerGroupId, redpanda.ConsumerOpts{
			SASLAuth: cli.GetKafkaSASLAuth(),
			Logger:   logger.WithName("infra-updates"),
			// }, []string{ev.KafkaTopicByocClientUpdates})
		}, []string{ev.KafkaTopicInfraUpdates})
	}),

	fx.Invoke(processInfraUpdates),

	fx.Invoke(func(server *fiber.App, logger logging.Logger) {
		server.Use("/ws", func(ctx *fiber.Ctx) error {
			if fWebsocket.IsWebSocketUpgrade(ctx) {
				return ctx.Next()
			}
			return fiber.ErrUpgradeRequired
		})

		server.Get("/ws/status-updates", fWebsocket.New(func(conn *fWebsocket.Conn) {
			logger.Infof("new socket request received ...")
			defer conn.Close()
			conn.WriteJSON(map[string]any{"hello": "world"})
		}))
	}),

	fx.Invoke(
		func(
			server *fiber.App,
			d domain.Domain,
			cacheClient AuthCacheClient,
			env *env.Env,
		) {
			config := generated.Config{Resolvers: &graph.Resolver{Domain: d}}

			config.Directives.IsLoggedIn = func(ctx context.Context, _ interface{}, next graphql.Resolver) (res interface{}, err error) {
				sess := httpServer.GetSession[*common.AuthSession](ctx)
				if sess == nil {
					return nil, fiber.ErrUnauthorized
				}
				return next(ctx)
			}

			config.Directives.IsLoggedInAndVerified = func(ctx context.Context, _ interface{}, next graphql.Resolver) (res interface{}, err error) {
				sess := httpServer.GetSession[*common.AuthSession](ctx)
				if sess == nil {
					return nil, fiber.ErrUnauthorized
				}

				if sess.UserVerified {
					return next(ctx)
				}

				return nil, &fiber.Error{
					Code:    fiber.ErrUnauthorized.Code,
					Message: "user's email is not verified, yet",
				}
			}

			config.Directives.HasAccount = func(ctx context.Context, _ interface{}, next graphql.Resolver) (res interface{}, err error) {
				sess := httpServer.GetSession[*common.AuthSession](ctx)
				if sess == nil {
					return nil, fiber.ErrUnauthorized
				}

				m := httpServer.GetHttpCookies(ctx)
				klAccount := m[env.AccountCookieName]
				if klAccount == "" {
					return nil, fmt.Errorf("no cookie named '%s' present in request", env.AccountCookieName)
				}
				cc := domain.InfraContext{
					Context:     ctx,
					AccountName: klAccount,
					UserId:      sess.UserId,
					UserName:    sess.UserName,
					UserEmail:   sess.UserEmail,
				}
				return next(context.WithValue(ctx, "infra-ctx", cc))
			}

			schema := generated.NewExecutableSchema(config)
			httpServer.SetupGQLServer(
				server,
				schema,
				httpServer.NewSessionMiddleware[*common.AuthSession](
					cacheClient,
					"hotspot-session",
					env.CookieDomain,
					constants.CacheSessionPrefix,
				),
			)
		},
	),
)
