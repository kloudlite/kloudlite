package app

import (
	"context"
	"encoding/json"
	"github.com/kloudlite/api/apps/worker-audit-logging/internal/domain"
	"github.com/kloudlite/api/apps/worker-audit-logging/internal/env"
	"github.com/kloudlite/api/common"
	"github.com/kloudlite/api/pkg/logging"
	"github.com/kloudlite/api/pkg/messaging"
	msg_nats "github.com/kloudlite/api/pkg/messaging/nats"
	"github.com/kloudlite/api/pkg/messaging/types"
	"github.com/kloudlite/api/pkg/nats"

	"github.com/kloudlite/api/pkg/repos"
	"go.uber.org/fx"
)

type EventLogConsumer messaging.Consumer

var Module = fx.Module("app",
	repos.NewFxMongoRepo[*domain.EventLog]("events", "ev", domain.EventLogIndices),
	fx.Provide(func(jc *nats.JetstreamClient, ev *env.Env, logger logging.Logger) (EventLogConsumer, error) {
		topic := common.AuditEventLogTopicName
		consumerName := "worker-audit-logging:event-log"
		return msg_nats.NewJetstreamConsumer(context.TODO(), jc, msg_nats.JetstreamConsumerArgs{
			Stream: ev.EventLogNatsStream,
			ConsumerConfig: msg_nats.ConsumerConfig{
				Name:           consumerName,
				Durable:        consumerName,
				Description:    "this consumer reads message from a subject dedicated to errors, that occurred when the resource was applied at the agent",
				FilterSubjects: []string{string(topic)},
			},
		})
	}),

	fx.Invoke(func(consumer EventLogConsumer, logr logging.Logger, d domain.Domain) error{
		if err := consumer.Consume(func(msg *types.ConsumeMsg) error {
			logger := logr.WithName("audit-events")
			logger.Infof("started processing")
			defer func() {
				logger.Infof("finished processing")
			}()

			var el domain.EventLog
			if err := json.Unmarshal(msg.Payload, &el); err != nil {
				return err
			}

			event, err := d.PushEvent(context.TODO(), &el)
			if err != nil {
				return err
			}

			logger.WithKV("event-id", event.Id).Infof("pushed event to mongo")
			return nil
		}, types.ConsumeOpts{}); err != nil {
			logr.Errorf(err, "error consuming messages")
			return err
		}
		return nil
	}),
	domain.Module,
)
