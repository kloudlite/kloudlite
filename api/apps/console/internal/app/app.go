package app

import (
	"context"

	"github.com/kloudlite/api/grpc-interfaces/kloudlite.io/rpc/console"
	"github.com/kloudlite/api/pkg/k8s"

	"github.com/kloudlite/api/pkg/errors"

	"github.com/99designs/gqlgen/graphql"
	"github.com/gofiber/fiber/v2"
	"go.uber.org/fx"

	"github.com/kloudlite/api/apps/console/internal/app/graph"
	"github.com/kloudlite/api/apps/console/internal/app/graph/generated"
	"github.com/kloudlite/api/apps/console/internal/domain"
	"github.com/kloudlite/api/apps/console/internal/entities"
	"github.com/kloudlite/api/apps/console/internal/env"
	"github.com/kloudlite/api/common"
	"github.com/kloudlite/api/constants"
	"github.com/kloudlite/api/grpc-interfaces/infra"
	"github.com/kloudlite/api/grpc-interfaces/kloudlite.io/rpc/iam"
	"github.com/kloudlite/api/pkg/grpc"
	httpServer "github.com/kloudlite/api/pkg/http-server"
	"github.com/kloudlite/api/pkg/kv"
	"github.com/kloudlite/api/pkg/logging"
	msg_nats "github.com/kloudlite/api/pkg/messaging/nats"
	"github.com/kloudlite/api/pkg/nats"
	"github.com/kloudlite/api/pkg/repos"

	"github.com/miekg/dns"
)

type (
	IAMGrpcClient grpc.Client
	InfraClient   grpc.Client
)

type (
	ConsoleGrpcServer grpc.Server
)

type DNSServer struct {
	*dns.Server
}

func toConsoleContext(requestCtx context.Context, accountCookieName string) (domain.ConsoleContext, error) {
	sess := httpServer.GetSession[*common.AuthSession](requestCtx)
	if sess == nil {
		return domain.ConsoleContext{}, fiber.ErrUnauthorized
	}
	m := httpServer.GetHttpCookies(requestCtx)
	klAccount := m[accountCookieName]
	if klAccount == "" {
		return domain.ConsoleContext{}, errors.Newf("no cookie named '%s' present in request", accountCookieName)
	}

	return domain.NewConsoleContext(requestCtx, sess.UserId, klAccount), nil
}

var Module = fx.Module("app",
	repos.NewFxMongoRepo[*entities.Environment]("environments", "env", entities.EnvironmentIndexes),
	repos.NewFxMongoRepo[*entities.App]("apps", "app", entities.AppIndexes),
	repos.NewFxMongoRepo[*entities.ExternalApp]("ext_apps", "extapp", entities.ExternalAppIndexes),
	repos.NewFxMongoRepo[*entities.Config]("configs", "cfg", entities.ConfigIndexes),
	repos.NewFxMongoRepo[*entities.Secret]("secrets", "scrt", entities.SecretIndexes),
	repos.NewFxMongoRepo[*entities.ManagedResource]("managed_resources", "mres", entities.MresIndexes),
	repos.NewFxMongoRepo[*entities.ImportedManagedResource]("imported_managed_resources", "impmres", entities.ImportedManagedResourceIndexes),
	repos.NewFxMongoRepo[*entities.Router]("routers", "rt", entities.RouterIndexes),
	repos.NewFxMongoRepo[*entities.ImagePullSecret]("image_pull_secrets", "ips", entities.ImagePullSecretIndexes),
	repos.NewFxMongoRepo[*entities.ResourceMapping]("resource_mappings", "rmap", entities.ResourceMappingIndices),
	repos.NewFxMongoRepo[*entities.ConsoleVPNDevice]("vpn_devices", "devs", entities.VPNDeviceIndexes),
	repos.NewFxMongoRepo[*entities.ServiceBinding]("service_bindings", "svcb", entities.ServiceBindingIndexes),

	fx.Provide(
		func(conn IAMGrpcClient) iam.IAMClient {
			return iam.NewIAMClient(conn)
		},
	),

	fx.Invoke(
		func(server httpServer.Server, d domain.Domain, sessionRepo kv.Repo[*common.AuthSession], ev *env.Env) {
			gqlConfig := generated.Config{Resolvers: &graph.Resolver{Domain: d}}

			gqlConfig.Directives.IsLoggedIn = func(ctx context.Context, obj interface{}, next graphql.Resolver) (res interface{}, err error) {
				sess := httpServer.GetSession[*common.AuthSession](ctx)
				if sess == nil {
					return nil, fiber.ErrUnauthorized
				}

				return next(context.WithValue(ctx, "user-session", sess))
			}

			gqlConfig.Directives.IsLoggedInAndVerified = func(ctx context.Context, obj interface{}, next graphql.Resolver) (res interface{}, err error) {
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

			gqlConfig.Directives.HasAccount = func(ctx context.Context, obj interface{}, next graphql.Resolver) (res interface{}, err error) {
				sess := httpServer.GetSession[*common.AuthSession](ctx)
				if sess == nil {
					return nil, fiber.ErrUnauthorized
				}
				m := httpServer.GetHttpCookies(ctx)
				klAccount := m[ev.AccountCookieName]
				if klAccount == "" {
					return nil, errors.Newf("no cookie named %q present in request", ev.AccountCookieName)
				}

				nctx := context.WithValue(ctx, "user-session", sess)
				nctx = context.WithValue(nctx, "account-name", klAccount)
				return next(nctx)
			}

			schema := generated.NewExecutableSchema(gqlConfig)
			server.SetupGraphqlServer(schema,
				httpServer.NewReadSessionMiddleware(sessionRepo, constants.CookieName, constants.CacheSessionPrefix),
			)
		},
	),

	fx.Provide(func(jc *nats.JetstreamClient, logger logging.Logger) domain.MessageDispatcher {
		return msg_nats.NewJetstreamProducer(jc)
	}),

	fx.Invoke(func(lf fx.Lifecycle, producer domain.MessageDispatcher) {
		lf.Append(fx.Hook{
			OnStop: func(ctx context.Context) error {
				return producer.Stop(ctx)
			},
		})
	}),

	fx.Provide(
		func(conn InfraClient) infra.InfraClient {
			return infra.NewInfraClient(conn)
		},
	),

	fx.Provide(func(cli *nats.Client, logger logging.Logger) domain.ResourceEventPublisher {
		return NewResourceEventPublisher(cli, logger)
	}),

	domain.Module,

	fx.Provide(func(d domain.Domain, kcli k8s.Client) console.ConsoleServer {
		return newConsoleGrpcServer(d, kcli)
	}),

	fx.Invoke(func(gserver ConsoleGrpcServer, srv console.ConsoleServer) {
		console.RegisterConsoleServer(gserver, srv)
	}),

	fx.Provide(func(jc *nats.JetstreamClient, ev *env.Env, logger logging.Logger) (ErrorOnApplyConsumer, error) {
		topic := common.GetPlatformClusterMessagingTopic("*", "*", common.ConsoleReceiver, common.EventErrorOnApply)
		consumerName := "console:error-on-apply"
		return msg_nats.NewJetstreamConsumer(context.TODO(), jc, msg_nats.JetstreamConsumerArgs{
			Stream: ev.NatsResourceSyncStream,
			ConsumerConfig: msg_nats.ConsumerConfig{
				Name:           consumerName,
				Durable:        consumerName,
				Description:    "this consumer reads message from a subject dedicated to errors, that occurred when the resource was applied at the agent",
				FilterSubjects: []string{topic},
			},
		})
	}),

	fx.Invoke(func(lf fx.Lifecycle, consumer ErrorOnApplyConsumer, d domain.Domain, logger logging.Logger) {
		lf.Append(fx.Hook{
			OnStart: func(context.Context) error {
				go ProcessErrorOnApply(consumer, d, logger)
				return nil
			},
			OnStop: func(ctx context.Context) error {
				return consumer.Stop(ctx)
			},
		})
	}),

	fx.Provide(func(jc *nats.JetstreamClient, ev *env.Env, logger logging.Logger) (ResourceUpdateConsumer, error) {
		topic := common.GetPlatformClusterMessagingTopic("*", "*", common.ConsoleReceiver, common.EventResourceUpdate)

		consumerName := "console:resource-updates"
		return msg_nats.NewJetstreamConsumer(context.TODO(), jc, msg_nats.JetstreamConsumerArgs{
			Stream: ev.NatsResourceSyncStream,
			ConsumerConfig: msg_nats.ConsumerConfig{
				Name:           consumerName,
				Durable:        consumerName,
				Description:    "this consumer reads message from a subject dedicated to console resource updates from tenant clusters",
				FilterSubjects: []string{topic},
			},
		})
	}),

	fx.Invoke(func(lf fx.Lifecycle, consumer ResourceUpdateConsumer, d domain.Domain, logger logging.Logger) {
		lf.Append(fx.Hook{
			OnStart: func(context.Context) error {
				go ProcessResourceUpdates(consumer, d, logger)
				return nil
			},
			OnStop: func(ctx context.Context) error {
				return consumer.Stop(ctx)
			},
		})
	}),

	fx.Provide(func(svcBindingRepo repos.DbRepo[*entities.ServiceBinding]) domain.ServiceBindingDomain {
		return domain.NewSvcBindingDomain(svcBindingRepo)
	}),

	fx.Provide(func(logger logging.Logger, sbd domain.ServiceBindingDomain) *dnsHandler {
		return &dnsHandler{
			logger:               logger,
			serviceBindingDomain: sbd,
		}
	}),

	fx.Invoke(func(server *DNSServer, handler *dnsHandler, lf fx.Lifecycle, logger logging.Logger) {
		lf.Append(fx.Hook{
			OnStart: func(ctx context.Context) error {
				logger.Infof("starting dns server at %s", server.Addr)
				server.Handler = handler
				go func() {
					err := server.ListenAndServe()
					if err != nil {
						logger.Errorf(err, "failed to start dns server")
						panic(err)
					}
				}()
				return nil
			},
			OnStop: func(ctx context.Context) error {
				return server.ShutdownContext(ctx)
			},
		})
	}),
)
