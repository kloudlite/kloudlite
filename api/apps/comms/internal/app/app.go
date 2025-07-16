package app

import (
	"context"
	"github.com/kloudlite/api/apps/comms/internal/app/grpc"
	"github.com/kloudlite/api/apps/comms/internal/domain"
	"github.com/kloudlite/api/grpc-interfaces/kloudlite.io/rpc/comms"
	"github.com/kloudlite/api/pkg/mail"
	googleGrpc "google.golang.org/grpc"

	"github.com/kloudlite/api/apps/comms/internal/domain/entities"
	"github.com/kloudlite/api/apps/comms/internal/env"
	"github.com/kloudlite/api/apps/comms/types"
	"github.com/kloudlite/api/common"
	"github.com/kloudlite/api/pkg/messaging"
	klnats "github.com/kloudlite/api/pkg/messaging/nats"
	"github.com/kloudlite/api/pkg/nats"
	"github.com/kloudlite/api/pkg/repos"
	"log/slog"

	"go.uber.org/fx"
)

type NotificationConsumer messaging.Consumer

var Module = fx.Module("app",
	fx.Module(
		"mongo-repos",
		repos.NewFxMongoRepo[*entities.NotificationConf]("notification-configs", "note-conf", entities.NotificationConfIndexes),
		repos.NewFxMongoRepo[*entities.Subscription]("subscriptions", "sub", entities.SubscriptionIndexes),
		repos.NewFxMongoRepo[*types.Notification]("notifications", "note", types.NotificationIndexes),
	),

	fx.Module(
		"sendgrid-mailer",
		fx.Provide(func(ev *env.CommsEnv) mail.Mailer {
			return mail.NewSendgridMailer(ev.SendgridApiKey)
		}),
	),

	fx.Module(
		"grpc-servers",
		fx.Provide(grpc.NewServer),
		fx.Invoke(func(server *googleGrpc.Server, serverImpl comms.CommsServer) {
			comms.RegisterCommsServer(server, serverImpl)
		}),
	),

	fx.Module(
		"nats-publisher",
		fx.Provide(func(cli *nats.Client, logger *slog.Logger) domain.ResourceEventPublisher {
			return NewResourceEventPublisher(cli, logger)
		}),
	),

	fx.Module(
		"nats-consumers",
		fx.Provide(func(jc *nats.JetstreamClient, ev *env.CommsEnv) (NotificationConsumer, error) {
			topic := string(common.NotificationTopicName)
			consumerName := "ntfy:message"
			return klnats.NewJetstreamConsumer(context.TODO(), jc, klnats.JetstreamConsumerArgs{
				Stream: ev.NotificationNatsStream,
				ConsumerConfig: klnats.ConsumerConfig{
					Name:        consumerName,
					Durable:     consumerName,
					Description: "this consumer reads message from a subject dedicated to errors, that occurred when the resource was applied at the agent",
					FilterSubjects: []string{
						topic,
					},
				},
			})
		}),
		fx.Invoke(func(lf fx.Lifecycle, consumer NotificationConsumer, d domain.Domain, logr *slog.Logger) {
			lf.Append(fx.Hook{
				OnStart: func(ctx context.Context) error {
					go func() {
						err := processNotification(ctx, d, consumer, logr)
						if err != nil {
							logr.Error(err.Error(), "could not process notifications")
						}
					}()
					return nil
				},
				OnStop: func(ctx context.Context) error {
					return nil
				},
			})
		}),
	),
	domain.Module,
)
