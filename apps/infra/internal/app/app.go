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
	"kloudlite.io/pkg/kafka"
	"kloudlite.io/pkg/logging"
	"kloudlite.io/pkg/repos"
)

type AuthCacheClient cache.Client

type (
	IAMGrpcClient                   grpc.Client
	AccountGrpcClient               grpc.Client
	MessageOfficeInternalGrpcClient grpc.Client
)

type (
	InfraGrpcServer grpc.Server
)

var Module = fx.Module(
	"app",
	repos.NewFxMongoRepo[*entities.Cluster]("clusters", "clus", entities.ClusterIndices),
	repos.NewFxMongoRepo[*entities.BYOCCluster]("byoc_clusters", "byoc", entities.BYOCClusterIndices),
	repos.NewFxMongoRepo[*entities.DomainEntry]("domain_entries", "de", entities.DomainEntryIndices),
	repos.NewFxMongoRepo[*entities.NodePool]("node_pools", "npool", entities.NodePoolIndices),
	repos.NewFxMongoRepo[*entities.Node]("node", "node", entities.NodePoolIndices),
	repos.NewFxMongoRepo[*entities.CloudProviderSecret]("cloud_provider_secrets", "cps", entities.CloudProviderSecretIndices),
	repos.NewFxMongoRepo[*entities.VPNDevice]("vpn_devices", "vpnd", entities.VPNDeviceIndexes),
	repos.NewFxMongoRepo[*entities.PersistentVolumeClaim]("pvcs", "pvc", entities.PersistentVolumeClaimIndices),
	repos.NewFxMongoRepo[*entities.BuildRun]("build_runs", "build_run", entities.BuildRunIndices),

	fx.Provide(
		func(conn IAMGrpcClient) iam.IAMClient {
			return iam.NewIAMClient(conn)
		},
	),

	fx.Provide(func(conn AccountGrpcClient) accounts.AccountsClient {
		return accounts.NewAccountsClient(conn)
	}),

	fx.Provide(func(client MessageOfficeInternalGrpcClient) message_office_internal.MessageOfficeInternalClient {
		return message_office_internal.NewMessageOfficeInternalClient(client)
	}),

	fx.Provide(func(conn kafka.Conn, logger logging.Logger) (domain.SendTargetClusterMessagesProducer, error) {
		return kafka.NewProducer(conn, kafka.ProducerOpts{
			Logger: logger.WithName("target-cluster-messages-producer"),
		})
	}),
	fx.Invoke(func(lf fx.Lifecycle, producer domain.SendTargetClusterMessagesProducer) {
		lf.Append(fx.Hook{
			OnStart: func(ctx context.Context) error {
				return producer.LifecycleOnStart(ctx)
			},
			OnStop: func(ctx context.Context) error {
				return producer.LifecycleOnStop(ctx)
			},
		})
	}),

	domain.Module,

	fx.Provide(func(d domain.Domain) infra.InfraServer {
		return newGrpcServer(d)
	}),

	fx.Invoke(func(gserver InfraGrpcServer, srv infra.InfraServer) {
		infra.RegisterInfraServer(gserver, srv)
	}),

	fx.Provide(func(conn kafka.Conn, ev *env.Env, logger logging.Logger) (ReceiveInfraUpdatesConsumer, error) {
		return kafka.NewConsumer(conn, ev.KafkaConsumerGroupId, []string{ev.KafkaTopicInfraUpdates}, kafka.ConsumerOpts{
			Logger: logger.WithName("infra-updates"),
		})
	}),
	fx.Invoke(func(lf fx.Lifecycle, consumer ReceiveInfraUpdatesConsumer, d domain.Domain) {
		lf.Append(fx.Hook{
			OnStart: func(ctx context.Context) error {
				go processInfraUpdates(consumer, d)
				return consumer.LifecycleOnStart(ctx)
			},
			OnStop: func(ctx context.Context) error {
				return consumer.LifecycleOnStop(ctx)
			},
		})
	}),

	fx.Invoke(
		func(server httpServer.Server, d domain.Domain, cacheClient AuthCacheClient, env *env.Env) {
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
			server.SetupGraphqlServer(schema,
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
