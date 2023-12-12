package app

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"kloudlite.io/constants"
	"net/http"

	"github.com/gofiber/fiber/v2"
	"github.com/xanzy/go-gitlab"
	"go.uber.org/fx"
	"kloudlite.io/apps/webhooks/internal/env"
	"kloudlite.io/pkg/errors"
	"kloudlite.io/pkg/logging"
	"kloudlite.io/pkg/redpanda"
	"kloudlite.io/pkg/types"
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
	if len(cHash) != len(ghSignature) || cHash != ghSignature[0] {
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
				return "", err
			}
			return evt.Repo.HtmlUrl, nil
		}

	case "gitlab":
		{
			var ev struct {
				Repo gitlab.Repository `json:"repository"`
			}
			if err := json.Unmarshal(hookBody, &ev); err != nil {
				return "", err
			}

			return ev.Repo.GitHTTPURL, nil
		}
	}
	return "", errors.Newf("unknown git provider")
}

func LoadGitWebhook() fx.Option {
	return fx.Invoke(
		func(app *fiber.App, envVars *env.Env, producer redpanda.Producer, logr logging.Logger) error {
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
						return err
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
						return err
					}

					msg, err := producer.Produce(ctx.Context(), envVars.GitWebhooksTopic, repoUrl, b)
					if err != nil {
						errMsg := fmt.Sprintf("could not produce message to topic %s", gitProvider)
						logger.Errorf(err, errMsg)
						return ctx.Status(http.StatusInternalServerError).JSON(errMsg)
					}

					logger.WithKV(
						"produced.offset", msg.Offset,
						"produced.topic", msg.Topic,
						"produced.timestamp", msg.Timestamp,
					).Infof("queued webhook")
					return ctx.Status(http.StatusAccepted).JSON(msg)
				},
			)
			return nil
		},
	)
}
