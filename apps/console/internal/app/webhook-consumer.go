package app

import (
	"context"
	"encoding/json"
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

		webhook := &domain.ImageHookPayload{}
		if err := json.Unmarshal(msg.Payload, &webhook); err != nil {
			logger.Errorf(err, "could not unmarshal into *ImageHookPayload")
			return errors.NewE(err)
		}
		if webhook.Image == "" || webhook.AccountName == "" {
			return errors.Newf("invalid webhook payload")
		}
		hook := &domain.ImageHookPayload{
			Image:       webhook.Image,
			AccountName: webhook.AccountName,
			Meta:        webhook.Meta,
		}

		_, err := d.CreateRegistryImage(context.TODO(), hook.AccountName, hook.Image, hook.Meta)
		if err != nil {
			logger.Errorf(err, "could not process image hook")
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
