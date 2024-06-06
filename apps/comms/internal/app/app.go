package app

import (
	"context"

	"github.com/99designs/gqlgen/graphql"
	"github.com/gofiber/fiber/v2"
	"github.com/kloudlite/api/constants"
	httpServer "github.com/kloudlite/api/pkg/http-server"
	"github.com/kloudlite/api/pkg/kv"

	"github.com/kloudlite/api/apps/comms/internal/app/graph"
	"github.com/kloudlite/api/apps/comms/internal/app/graph/generated"
	"github.com/kloudlite/api/apps/comms/internal/domain"

	"github.com/kloudlite/api/apps/comms/internal/domain/entities"
	"github.com/kloudlite/api/apps/comms/internal/env"
	"github.com/kloudlite/api/apps/comms/types"
	"github.com/kloudlite/api/common"
	"github.com/kloudlite/api/pkg/errors"
	"github.com/kloudlite/api/pkg/logging"
	"github.com/kloudlite/api/pkg/messaging"
	msg_nats "github.com/kloudlite/api/pkg/messaging/nats"
	"github.com/kloudlite/api/pkg/nats"
	"github.com/kloudlite/api/pkg/repos"

	"github.com/kloudlite/api/pkg/grpc"

	"github.com/kloudlite/api/grpc-interfaces/kloudlite.io/rpc/comms"
	"github.com/kloudlite/api/grpc-interfaces/kloudlite.io/rpc/iam"
	"go.uber.org/fx"
)

type (
	IAMGrpcClient grpc.Client
)

type NotificationConsumer messaging.Consumer

type CommsGrpcServer grpc.Server

var Module = fx.Module("app",
	repos.NewFxMongoRepo[*entities.NotificationConf]("nconfs", "prj", entities.NotificationConfIndexes),
	repos.NewFxMongoRepo[*entities.Subscription]("subscriptions", "prj", entities.SubscriptionIndexes),
	repos.NewFxMongoRepo[*types.Notification]("notifications", "prj", entities.SubscriptionIndexes),

	domain.Module,

	fx.Provide(func(jc *nats.JetstreamClient, ev *env.Env, logger logging.Logger) (NotificationConsumer, error) {
		topic := string(common.NotificationTopicName)
		consumerName := "ntfy:message"
		return msg_nats.NewJetstreamConsumer(context.TODO(), jc, msg_nats.JetstreamConsumerArgs{
			Stream: ev.NotificationNatsStream,
			ConsumerConfig: msg_nats.ConsumerConfig{
				Name:        consumerName,
				Durable:     consumerName,
				Description: "this consumer reads message from a subject dedicated to errors, that occurred when the resource was applied at the agent",
				FilterSubjects: []string{
					topic,
				},
			},
		})
	}),

	fx.Provide(
		func(conn IAMGrpcClient) iam.IAMClient {
			return iam.NewIAMClient(conn)
		},
	),

	fx.Provide(func(et domain.EmailTemplatesDir) (*domain.EmailTemplates, error) {
		return domain.GetEmailTemplates(et)
	}),

	fx.Provide(newCommsSvc),

	fx.Invoke(func(server CommsGrpcServer, commsServer comms.CommsServer) {
		comms.RegisterCommsServer(server, commsServer)
	}),

	fx.Provide(func(cli *nats.Client, logger logging.Logger) domain.ResourceEventPublisher {
		return NewResourceEventPublisher(cli, logger)
	}),

	fx.Invoke(
		func(server httpServer.Server, d domain.Domain, sessionRepo kv.Repo[*common.AuthSession], ev *env.Env) {
			gqlConfig := generated.Config{Resolvers: &graph.Resolver{Domain: d, Env: ev}}

			gqlConfig.Directives.IsLoggedInAndVerified = func(ctx context.Context, _ interface{}, next graphql.Resolver) (res interface{}, err error) {
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

			gqlConfig.Directives.HasAccount = func(ctx context.Context, _ interface{}, next graphql.Resolver) (res interface{}, err error) {
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
			server.SetupGraphqlServer(schema, httpServer.NewReadSessionMiddleware(sessionRepo, constants.CookieName, constants.CacheSessionPrefix))
		},
	),

	fx.Invoke(func(lf fx.Lifecycle, consumer NotificationConsumer, d domain.Domain, logr logging.Logger) {
		lf.Append(fx.Hook{
			OnStart: func(ctx context.Context) error {
				go func() {
					err := processNotification(ctx, d, consumer, logr)
					if err != nil {
						logr.Errorf(err, "could not process notifications")
					}
				}()
				return nil
			},
			OnStop: func(ctx context.Context) error {
				return nil
			},
		})
	}),
)
