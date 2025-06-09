package app

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/kloudlite/api/apps/console/internal/domain"
	"github.com/kloudlite/api/pkg/errors"
	"github.com/kloudlite/api/pkg/logging"
	msgTypes "github.com/kloudlite/api/pkg/messaging/types"
)

func processWebhooks(consumer WebhookConsumer, d domain.Domain, logger logging.Logger) error {
	err := consumer.Consume(func(msg *msgTypes.ConsumeMsg) error {
		logger := logger.WithName("webhook-consumer")
		logger.Infof("started processing message")

		defer func() {
			logger.Infof("finished processing message")
		}()

		hook := &domain.ImageHookPayload{}
		if err := json.Unmarshal(msg.Payload, &hook); err != nil {
			logger.Errorf(err, "could not unmarshal into *ImageHookPayload")
			return errors.NewE(err)
		}
		if hook.Image == "" || hook.AccountName == "" {
			return errors.Newf("invalid webhook payload")
		}

		_, err := d.UpsertRegistryImage(context.TODO(), hook.AccountName, hook.Image, hook.Meta)
		if err != nil {
			logger.Errorf(err, "could not process image hook")
			return errors.NewE(err)
		}

		// domain.NewConsoleContext(ctx, userId repos.ID, accountName string)
		dctx := domain.NewConsoleContext(context.TODO(), "sys-user:apply-on-error-worker", hook.AccountName)

		if err := d.RolloutAppsByImage(dctx, fmt.Sprintf("%s:%s", hook.Image, hook.Image)); err != nil {
			logger.Errorf(err, "could not rollout apps by image")
			return errors.NewE(err)
		}

		return nil
	}, msgTypes.ConsumeOpts{
		OnError: func(err error) error {
			logger.Errorf(err, "error while consuming message")
			return nil
		},
	})
	if err != nil {
		return errors.NewE(err)
	}

	return nil
}
