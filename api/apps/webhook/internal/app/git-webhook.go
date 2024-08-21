package app

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/kloudlite/api/apps/webhook/internal/domain"
	"github.com/kloudlite/api/common"
	"github.com/kloudlite/api/constants"
	httpServer "github.com/kloudlite/api/pkg/http-server"
	"github.com/kloudlite/api/pkg/messaging"
	types2 "github.com/kloudlite/api/pkg/messaging/types"
	"net/http"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/kloudlite/api/apps/webhook/internal/env"
	"github.com/kloudlite/api/pkg/errors"
	"github.com/kloudlite/api/pkg/logging"

	"github.com/kloudlite/api/pkg/types"
	"github.com/xanzy/go-gitlab"
	"go.uber.org/fx"
)

func validateGithubHook(ctx *fiber.Ctx, envVars *env.Env) (bool, error) {
	headers := ctx.GetReqHeaders()
	if v, ok := headers["X-Kloudlite-Trigger"]; ok {
		if len(v) != len(envVars.KlHookTriggerAuthzSecret) || v[0] != envVars.KlHookTriggerAuthzSecret {
			return false, errors.Newf("signature (%s) is invalid, sorry would need to drop the message", v)
		}
		return true, nil
	}

	hash := hmac.New(sha256.New, []byte(envVars.GithubAuthzSecret))
	hash.Write(ctx.Body())
	cHash := "sha256=" + hex.EncodeToString(hash.Sum(nil))

	ghSignature := headers["X-Hub-Signature-256"]
	if len(cHash) != len(ghSignature[0]) || cHash != ghSignature[0] {
		return false, errors.Newf("signature (%s) is invalid, sorry would need to drop the message", ghSignature)
	}
	return true, nil
}

func validateGitlabHook(ctx *fiber.Ctx, envVars *env.Env) (bool, error) {
	headers := ctx.GetReqHeaders()
	if v, ok := headers["X-Kloudlite-Trigger"]; ok {
		if len(v) != len(envVars.KlHookTriggerAuthzSecret) || v[0] != envVars.KlHookTriggerAuthzSecret {
			return false, errors.Newf("signature (%s) is invalid, sorry would need to drop the message", v)
		}
		return true, nil
	}

	gToken := headers["X-Gitlab-Token"]
	if len(envVars.GitlabAuthzSecret) != len(gToken) || envVars.GitlabAuthzSecret != gToken[0] {
		return false, errors.Newf("signature (%s) is invalid, sorry would need to drop the message", gToken)
	}
	return true, nil
}

func gitRepoUrl(provider string, hookBody []byte) (string, error) {
	switch provider {
	case "github":
		{
			// TODO: (immediate deletion, after github app webhook succeeds)
			// var evt struct {
			// 	Repo *github.Repository `json:"repository,omitempty"`
			// }
			// if err := json.Unmarshal(hookBody, &evt); err != nil {
			// 	return "", err
			// }
			// return *evt.Repo.HTMLURL, nil

			var evt struct {
				Repo struct {
					HtmlUrl string `json:"html_url,omitempty"`
				} `json:"repository"`
				// Repo *github.Repository `json:"repository,omitempty"`
			}
			if err := json.Unmarshal(hookBody, &evt); err != nil {
				return "", errors.NewE(err)
			}
			return evt.Repo.HtmlUrl, nil
		}

	case "gitlab":
		{
			var ev struct {
				Repo gitlab.Repository `json:"repository"`
			}
			if err := json.Unmarshal(hookBody, &ev); err != nil {
				return "", errors.NewE(err)
			}

			return ev.Repo.GitHTTPURL, nil
		}
	}
	return "", errors.Newf("unknown git provider")
}

func LoadGitWebhook() fx.Option {
	return fx.Invoke(
		func(server httpServer.Server, envVars *env.Env, producer messaging.Producer, logr logging.Logger, d domain.Domain) error {
			app := server.Raw()

			app.Get("/healthy", func(c *fiber.Ctx) error {
				return c.SendString("OK")
			})

			app.Post("/contact-us", func(ctx *fiber.Ctx) error {
				var data *domain.ContactUsData

				if err := ctx.BodyParser(&data); err != nil {
					return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request payload"})
				}

				err := d.SendContactUsEmail(ctx.Context(), data)
				if err != nil {
					return errors.NewE(err)
				}

				discordWebhookURL := envVars.DiscordWebhookUrl
				if discordWebhookURL != "" {
					discordMessage := fmt.Sprintf(
						""+
							"🚨 **NEW CONTACT US SUBMISSION** 🚨\n**Name:** %s\n**Email:** %s\n**Mobile:** %s\n**Company:** %s\n**Country:** %s\n**Message:** %s\n",
						data.Name, data.Email, data.MobileNumber, data.CompanyName, data.Country, data.Message,
					)

					payload := map[string]string{
						"content": discordMessage,
					}

					payloadBytes, err := json.Marshal(payload)
					if err != nil {
						return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to marshal Discord payload"})
					}

					resp, err := http.Post(discordWebhookURL, "application/json", bytes.NewBuffer(payloadBytes))
					if err != nil {
						return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to send message to Discord"})
					}
					defer resp.Body.Close()

					if resp.StatusCode != http.StatusNoContent {
						return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Discord API returned an error"})
					}
				}

				return ctx.Status(fiber.StatusOK).JSON(fiber.Map{"success": true})
			})

			app.Post(
				"/git/:provider", func(ctx *fiber.Ctx) error {
					logger := logr.WithName("git-webhook")

					gitProvider := ctx.Params("provider")

					_, err := func() (bool, error) {
						if gitProvider == constants.ProviderGithub {
							return validateGithubHook(ctx, envVars)
						}

						if gitProvider == constants.ProviderGitlab {
							return validateGitlabHook(ctx, envVars)
						}

						return false, errors.Newf("unknown git provider")
					}()

					if err != nil {
						logger.Errorf(err, "dropping webhook request")
						return ctx.Status(http.StatusUnauthorized).JSON(map[string]string{"error": err.Error()})
					}

					repoUrl, err := gitRepoUrl(gitProvider, ctx.Body())
					if err != nil {
						return errors.NewE(err)
					}
					logger = logger.WithKV("provider", gitProvider, "repo", repoUrl, "user-agent", ctx.GetReqHeaders()["User-Agent"])
					logger.Infof("received webhook")

					gitHook := types.GitHttpHook{
						HttpHook: types.HttpHook{
							Body:        ctx.Body(),
							Headers:     ctx.GetReqHeaders(),
							Url:         ctx.Request().URI().String(),
							QueryParams: ctx.Request().URI().QueryString(),
						},
						GitProvider: gitProvider,
					}
					b, err := json.Marshal(gitHook)
					if err != nil {
						return errors.NewE(err)
					}

					err = producer.Produce(ctx.Context(), types2.ProduceMsg{
						Subject: string(common.GitWebhookTopicName),
						Payload: b,
					})
					if err != nil {
						return errors.NewE(err)
					}

					if err != nil {
						errMsg := fmt.Sprintf("could not produce message to topic %s", gitProvider)
						logger.Errorf(err, errMsg)
						return ctx.Status(http.StatusInternalServerError).JSON(errMsg)
					}

					logger.WithKV(
						"produced.subject", string(common.GitWebhookTopicName),
						"produced.timestamp", time.Now(),
					).Infof("queued webhook")
					return ctx.Status(http.StatusAccepted).JSON(map[string]string{"status": "ok"})
				},
			)
			return nil
		},
	)
}
