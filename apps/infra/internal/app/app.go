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
	"kloudlite.io/pkg/messaging/nats"
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

	fx.Provide(func(conn AccountGrpcClient) (domain.AccountsSvc, error) {
		ac := accounts.NewAccountsClient(conn)
		return NewAccountsSvc(ac), nil
	}),

	fx.Provide(func(client MessageOfficeInternalGrpcClient) message_office_internal.MessageOfficeInternalClient {
		return message_office_internal.NewMessageOfficeInternalClient(client)
	}),

	fx.Provide(func(jsc *nats.JetstreamClient, logger logging.Logger) domain.SendTargetClusterMessagesProducer {
		return jsc.CreateProducer()
	}),

	fx.Invoke(func(lf fx.Lifecycle, producer domain.SendTargetClusterMessagesProducer) {
		lf.Append(fx.Hook{
			OnStop: func(ctx context.Context) error {
				return producer.Stop(ctx)
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

	fx.Provide(func(jsc *nats.JetstreamClient, ev *env.Env) (ReceiveInfraUpdatesConsumer, error) {
		topic := common.GetPlatformClusterMessagingTopic("*", "*", common.KloudliteInfra, common.EventErrorOnApply)

		consumerName := "infra:resource-updates"
		return jsc.CreateConsumer(context.TODO(), nats.JetstreamConsumerArgs{
			Stream: ev.NatsStream,
			ConsumerConfig: nats.ConsumerConfig{
				Name:           consumerName,
				Durable:        consumerName,
				Description:    "this consumer receives infra resource updates, processes them, and keeps our Database updated about things happening in the cluster",
				FilterSubjects: []string{topic},
			},
		})
	}),

	fx.Invoke(func(lf fx.Lifecycle, consumer ReceiveInfraUpdatesConsumer, d domain.Domain, logger logging.Logger) {
		lf.Append(fx.Hook{
			OnStart: func(context.Context) error {
				go processInfraUpdates(consumer, d, logger)
				return nil
			},
			OnStop: func(ctx context.Context) error {
				return consumer.Stop(ctx)
			},
		})
	}),

	fx.Provide(func(jsc *nats.JetstreamClient, ev *env.Env) (ErrorOnApplyConsumer, error) {
		topic := common.GetPlatformClusterMessagingTopic("*", "*", common.KloudliteInfra, common.EventErrorOnApply)

		consumerName := "infra:error-on-apply"
		return jsc.CreateConsumer(context.TODO(), nats.JetstreamConsumerArgs{
			Stream: ev.NatsStream,
			ConsumerConfig: nats.ConsumerConfig{
				Name:           consumerName,
				Durable:        consumerName,
				Description:    "this consumer receives infra resource apply error on agent, processes them, and keeps our Database updated about why the resource apply failed at agent",
				FilterSubjects: []string{topic},
			},
		})
	}),

	fx.Invoke(func(lf fx.Lifecycle, consumer ErrorOnApplyConsumer, d domain.Domain, logger logging.Logger) {
		lf.Append(fx.Hook{
			OnStart: func(context.Context) error {
				go ProcessErrorOnApply(consumer, logger, d)
				return nil
			},
			OnStop: func(ctx context.Context) error {
				return consumer.Stop(ctx)
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
