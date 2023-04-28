package app

import (
	"encoding/json"
	"net/http"

	"github.com/gofiber/fiber/v2"
	"go.uber.org/fx"
	"kloudlite.io/apps/webhooks/internal/env"
	"kloudlite.io/pkg/harbor"
	"kloudlite.io/pkg/logging"
	"kloudlite.io/pkg/redpanda"
	"kloudlite.io/pkg/types"
)

func getMsgKey(body []byte) string {
	var whBody harbor.WebhookBody
	if err := json.Unmarshal(body, &whBody); err != nil {
		return ""
	}
	return whBody.EventData.Repository.RepoFullName
}

func LoadHarborWebhook() fx.Option {
	return fx.Invoke(
		func(app *fiber.App, envVars *env.Env, producer redpanda.Producer, logr logging.Logger, p *publisher) error {
			app.Post(
				"/harbor", func(ctx *fiber.Ctx) error {
					logger := logr.WithName("harbor-webhook")

					logger.Infof("received webhook")

					headers := ctx.GetReqHeaders()
					if authz, ok := headers["Authorization"]; !ok || authz != envVars.HarborAuthzSecret {
						logger.Infof("bad authorization code, dropping request...")
						return ctx.Status(http.StatusUnauthorized).JSON("bad authorization token")
					}

					httpHook := types.HttpHook{
						Body:        ctx.Body(),
						Headers:     headers,
						Url:         ctx.Request().URI().String(),
						QueryParams: ctx.Request().URI().QueryString(),
					}

					b, err := json.Marshal(httpHook)
					if err != nil {
						return ctx.Status(http.StatusBadRequest).JSON(err.Error())
					}

					p.publishMessage(PublishMessage{
						Message: b,
						Topic:   envVars.HarborWebhookTopic,
						Key:     getMsgKey(httpHook.Body),
					})

					return ctx.SendStatus(http.StatusAccepted)

					// // tctx, cancelFunc := context.WithTimeout(ctx.Context(), 5*time.Second)
					// tctx, cancelFunc := context.WithTimeout(context.TODO(), 5*time.Second)
					// defer cancelFunc()
					// msg, err := producer.Produce(tctx, envVars.HarborWebhookTopic, getMsgKey(httpHook.Body), b)
					// if err != nil {
					// 	wErr := errors.NewEf(err, "could not produce message to topic %s", envVars.HarborWebhookTopic)
					// 	logger.Infof(wErr.Error())
					// 	return ctx.Status(http.StatusInternalServerError).JSON(wErr.Error())
					// }
					// logger = logger.WithKV(
					// 	"produced.offset", msg.Offset,
					// 	"produced.topic", msg.Topic,
					// 	"produced.timestamp", msg.Timestamp,
					// )
					// logger.Infof("queued webhook")
					// return ctx.Status(http.StatusAccepted).JSON(msg)
				},
			)
			return nil
		},
	)
}
