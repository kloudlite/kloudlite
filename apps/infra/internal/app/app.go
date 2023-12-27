package app

import (
	"context"

	"github.com/kloudlite/api/apps/infra/internal/entities"
	"github.com/kloudlite/api/pkg/errors"

	"github.com/99designs/gqlgen/graphql"
	"github.com/gofiber/fiber/v2"
	"github.com/kloudlite/api/apps/infra/internal/app/graph"
	"github.com/kloudlite/api/apps/infra/internal/app/graph/generated"
	"github.com/kloudlite/api/apps/infra/internal/domain"
	"github.com/kloudlite/api/apps/infra/internal/env"
	"github.com/kloudlite/api/common"
	"github.com/kloudlite/api/constants"
	"github.com/kloudlite/api/grpc-interfaces/kloudlite.io/rpc/accounts"
	"github.com/kloudlite/api/grpc-interfaces/kloudlite.io/rpc/iam"
	"github.com/kloudlite/api/grpc-interfaces/kloudlite.io/rpc/infra"
	message_office_internal "github.com/kloudlite/api/grpc-interfaces/kloudlite.io/rpc/message-office-internal"
	"github.com/kloudlite/api/pkg/cache"
	"github.com/kloudlite/api/pkg/grpc"
	httpServer "github.com/kloudlite/api/pkg/http-server"
	"github.com/kloudlite/api/pkg/logging"
	msg_nats "github.com/kloudlite/api/pkg/messaging/nats"
	"github.com/kloudlite/api/pkg/nats"
	"github.com/kloudlite/api/pkg/repos"
	"go.uber.org/fx"
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
	repos.NewFxMongoRepo[*entities.ClusterManagedService]("cmsvcs", "cmsvcs", entities.ClusterManagedServiceIndices),
	repos.NewFxMongoRepo[*entities.DomainEntry]("domain_entries", "de", entities.DomainEntryIndices),
	repos.NewFxMongoRepo[*entities.NodePool]("node_pools", "npool", entities.NodePoolIndices),
	repos.NewFxMongoRepo[*entities.Node]("node", "node", entities.NodePoolIndices),
	repos.NewFxMongoRepo[*entities.CloudProviderSecret]("cloud_provider_secrets", "cps", entities.CloudProviderSecretIndices),
	repos.NewFxMongoRepo[*entities.VPNDevice]("vpn_devices", "vpnd", entities.VPNDeviceIndexes),
	repos.NewFxMongoRepo[*entities.PersistentVolumeClaim]("pvcs", "pvc", entities.PersistentVolumeClaimIndices),

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

	fx.Provide(func(jsc *nats.JetstreamClient, logger logging.Logger) SendTargetClusterMessagesProducer {
		return msg_nats.NewJetstreamProducer(jsc)
	}),

	fx.Provide(func(p SendTargetClusterMessagesProducer) domain.ResourceDispatcher {
		return NewResourceDispatcher(p)
	}),

	fx.Invoke(func(lf fx.Lifecycle, producer SendTargetClusterMessagesProducer) {
		lf.Append(fx.Hook{
			OnStop: func(ctx context.Context) error {
				return producer.Stop(ctx)
			},
		})
	}),

	fx.Provide(func(cli *nats.Client, logger logging.Logger) domain.ResourceEventPublisher {
		return NewResourceEventPublisher(cli, logger)
	}),

	domain.Module,

	fx.Provide(func(d domain.Domain) infra.InfraServer {
		return newGrpcServer(d)
	}),

	fx.Invoke(func(gserver InfraGrpcServer, srv infra.InfraServer) {
		infra.RegisterInfraServer(gserver, srv)
	}),

	fx.Provide(func(jsc *nats.JetstreamClient, ev *env.Env) (ReceiveResourceUpdatesConsumer, error) {
		topic := common.GetPlatformClusterMessagingTopic("*", "*", common.InfraReceiver, common.EventResourceUpdate)

		consumerName := "infra:resource-updates"
		return msg_nats.NewJetstreamConsumer(context.TODO(), jsc, msg_nats.JetstreamConsumerArgs{
			Stream: ev.NatsStream,
			ConsumerConfig: msg_nats.ConsumerConfig{
				Name:           consumerName,
				Durable:        consumerName,
				Description:    "this consumer receives infra resource updates, processes them, and keeps our Database updated about things happening in the cluster",
				FilterSubjects: []string{topic},
			},
		})
	}),

	fx.Invoke(func(lf fx.Lifecycle, consumer ReceiveResourceUpdatesConsumer, d domain.Domain, logger logging.Logger) {
		lf.Append(fx.Hook{
			OnStart: func(context.Context) error {
				go processResourceUpdates(consumer, d, logger)
				return nil
			},
			OnStop: func(ctx context.Context) error {
				return consumer.Stop(ctx)
			},
		})
	}),

	fx.Provide(func(jsc *nats.JetstreamClient, ev *env.Env) (ErrorOnApplyConsumer, error) {
		topic := common.GetPlatformClusterMessagingTopic("*", "*", common.ConsoleReceiver, common.EventErrorOnApply)

		consumerName := "infra:error-on-apply"

		return msg_nats.NewJetstreamConsumer(context.TODO(), jsc, msg_nats.JetstreamConsumerArgs{
			Stream: ev.NatsStream,
			ConsumerConfig: msg_nats.ConsumerConfig{
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
		func(server httpServer.Server, d domain.Domain, sessionRepo cache.Repo[*common.AuthSession], env *env.Env) {
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
					return nil, errors.Newf("no cookie named '%s' present in request", env.AccountCookieName)
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
				httpServer.NewSessionMiddleware(
					sessionRepo,
					"hotspot-session",
					env.CookieDomain,
					constants.CacheSessionPrefix,
				),
			)
		},
	),
)
