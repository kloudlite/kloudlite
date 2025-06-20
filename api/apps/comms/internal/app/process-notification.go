package app

import (
	"context"
	"log/slog"

	"github.com/kloudlite/api/apps/comms/internal/domain"
	"github.com/kloudlite/api/pkg/messaging/types"

	ntypes "github.com/kloudlite/api/apps/comms/types"
)

func processNotification(ctx context.Context, d domain.Domain, consumer NotificationConsumer, logr *slog.Logger) error {
	return consumer.Consume(func(msg *types.ConsumeMsg) error {
		logr.Info("received notification %s (%s)", msg.Subject)

		notif := ntypes.Notification{}
		if err := notif.ParseBytes(msg.Payload); err != nil {
			return err
		}

		return d.Notify(context.Background(), &notif)
	},
		types.ConsumeOpts{
			OnError: func(err error) error {
				logr.Error(err.Error(), "could not consume notification")
				return nil
			},
		})
}
