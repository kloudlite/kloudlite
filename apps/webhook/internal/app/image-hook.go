package app

import (
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/gofiber/fiber/v2"
	"github.com/kloudlite/api/apps/webhook/internal/domain"
	"github.com/kloudlite/api/apps/webhook/internal/env"
	"github.com/kloudlite/api/common"
	httpServer "github.com/kloudlite/api/pkg/http-server"
	"github.com/kloudlite/api/pkg/logging"
	"github.com/kloudlite/api/pkg/messaging"
	types2 "github.com/kloudlite/api/pkg/messaging/types"
	"github.com/pkg/errors"
	"go.uber.org/fx"
	"net/http"
	"strings"
	"time"
)

func validateAndDecodeAccessToken(accessToken string, tokenSecret string) (accountName string, err error) {
	b, err := base64.StdEncoding.DecodeString(accessToken)
	if err != nil {
		return "", errors.Wrap(err, "invalid access token, incorrect format")
	}

	info := string(b)

	sp := strings.SplitN(info, ";sha256sum=", 2)

	if len(sp) != 2 {
		return "", errors.New("invalid access token, incorrect format")
	}
	data := sp[0]
	sum := sp[1]

	h := sha256.New()
	h.Write([]byte(data + tokenSecret))
	calculatedSum := fmt.Sprintf("%x", h.Sum(nil))

	if sum != calculatedSum {
		return "", errors.New("invalid access token, checksum mismatch")
	}

	s := strings.SplitN(data, ";", 2)

	if len(s) != 1 {
		return "", errors.New("invalid access token, incorrect data format")
	}
	for _, v := range strings.Split(s[0], ";") {
		sp := strings.SplitN(v, "=", 2)
		if len(sp) != 2 {
			return "", errors.New("invalid access token, incorrect data format")
		}
		if sp[0] == "account" {
			accountName = sp[1]
		}
	}
	return accountName, nil
}

func LoadImageHook() fx.Option {
	return fx.Invoke(
		func(server httpServer.Server, envVars *env.Env, producer messaging.Producer, logr logging.Logger, d domain.Domain) error {

			app := server.Raw()

			app.Post("/image", func(ctx *fiber.Ctx) error {
				logger := logr.WithName("image-hook")

				headers := ctx.GetReqHeaders()
				v, ok := headers["Authorization"]
				if !ok {
					return ctx.Status(fiber.StatusUnauthorized).SendString("no authorization header passed")
				}

				accountName, err := validateAndDecodeAccessToken(v[0], envVars.WebhookTokenHashingSecret)
				if err != nil {
					return ctx.Status(fiber.StatusUnauthorized).SendString(err.Error())
				}

				data := struct {
					Image       string         `json:"image"`
					AccountName string         `json:"accountName"`
					Meta        map[string]any `json:"meta"`
				}{}

				body := ctx.Body()
				if err := json.Unmarshal(body, &data); err != nil {
					return ctx.Status(fiber.StatusBadRequest).SendString(err.Error())
				}
				data.AccountName = accountName

				logger = logger.WithKV("account", data.AccountName, "image", data.Image, "meta", data.Meta)
				logger.Infof("received image-hook")

				jsonPayload, err := json.Marshal(data)
				if err != nil {
					return err
				}
				err = producer.Produce(ctx.Context(), types2.ProduceMsg{
					Subject: string(common.RegistryHookTopicName),
					Payload: jsonPayload,
				})
				if err != nil {
					return err
				}

				if err != nil {
					errMsg := fmt.Sprintf("failed to produce message: %s", "webhook-provider")
					logger.Errorf(err, errMsg)
					return ctx.Status(http.StatusInternalServerError).JSON(errMsg)
				}
				logger.WithKV(
					"produced.subject", string(common.RegistryHookTopicName),
					"produced.timestamp", time.Now(),
				).Infof("queued webhook")
				return ctx.Status(http.StatusAccepted).JSON(map[string]string{"status": "ok"})
			})
			return nil
		})
}
