package app

import (
	"context"
	"encoding/json"
	"github.com/kloudlite/api/apps/worker-audit-logging/internal/domain"
	"github.com/kloudlite/api/pkg/logging"
	"github.com/kloudlite/api/pkg/redpanda"
	"github.com/kloudlite/api/pkg/repos"
	"go.uber.org/fx"
	"time"
)

type EventLogConsumer redpanda.Consumer

var Module = fx.Module("app",
	repos.NewFxMongoRepo[*domain.EventLog]("events", "ev", domain.EventLogIndices),
	fx.Invoke(func(consumer redpanda.Consumer, logr logging.Logger, d domain.Domain) {
		consumer.StartConsuming(func(msg []byte, _ time.Time, offset int64) error {
			logger := logr.WithName("audit-events").WithKV("offset", offset)
			logger.Infof("started processing")
			defer func() {
				logger.Infof("finished processing")
			}()

			var el domain.EventLog
			if err := json.Unmarshal(msg, &el); err != nil {
				return err
			}

			event, err := d.PushEvent(context.TODO(), &el)
			if err != nil {
				return err
			}

			logger.WithKV("event-id", event.Id).Infof("pushed event to mongo")
			return nil
		})
	}),
	domain.Module,
	//fx.Invoke(func(lf fx.Lifecycle, producer redpanda.Producer, logger logging.Logger) {
	//	lf.Append(fx.Hook{
	//		OnStart: func(ctx context.Context) error {
	//			go func() {
	//				time.AfterFunc(5*time.Second, func() {
	//					bkon := beacon.NewBeacon(producer, "kl-events")
	//					if err := bkon.TriggerEvent(ctx, beacon.NewAuditLogEvent("sample", "user-asdf", "created", "creating a project asdjfkadsklf")); err != nil {
	//						logger.Errorf(err, "error triggering event")
	//					}
	//					logger.Infof("produced message")
	//				})
	//			}()
	//			return nil
	//		},
	//	})
	//}),
)
