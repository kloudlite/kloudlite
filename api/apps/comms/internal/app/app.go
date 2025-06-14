package app

import (
	"context"

	"github.com/kloudlite/api/apps/comms/internal/domain"

	"github.com/kloudlite/api/apps/comms/internal/domain/entities"
	"github.com/kloudlite/api/apps/comms/internal/env"
	"github.com/kloudlite/api/apps/comms/types"
	"github.com/kloudlite/api/common"
	"github.com/kloudlite/api/pkg/logging"
	"github.com/kloudlite/api/pkg/messaging"
	msg_nats "github.com/kloudlite/api/pkg/messaging/nats"
	"github.com/kloudlite/api/pkg/nats"
	"github.com/kloudlite/api/pkg/repos"

	"github.com/kloudlite/api/pkg/grpc"

	"github.com/kloudlite/api/grpc-interfaces/kloudlite.io/rpc/comms"
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

	// fx.Provide(
	// 	func(conn IAMGrpcClient) iam.IAMClient {
	// 		return iam.NewIAMClient(conn)
	// 	},
	// ),

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
