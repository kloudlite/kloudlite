package app

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gofiber/fiber/v2"
	"github.com/google/go-github/v43/github"
	"github.com/xanzy/go-gitlab"
	"go.uber.org/fx"
	"kloudlite.io/apps/webhooks/internal/env"
	"kloudlite.io/pkg/errors"
	"kloudlite.io/pkg/logging"
	"kloudlite.io/pkg/redpanda"
)

func gitRepoUrl(provider string, hookBody []byte) (string, error) {
	switch provider {
	case "github":
		{
			var evt struct {
				Repo *github.Repository `json:"repository,omitempty"`
			}
			if err := json.Unmarshal(hookBody, &evt); err != nil {
				return "", err
			}
			return *evt.Repo.HTMLURL, nil
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

type GitWebhookPayload struct {
	Provider   string            `json:"provider"`
	Body       []byte            `json:"body"`
	ReqHeaders map[string]string `json:"reqHeaders"`
}

var Module = fx.Module(
	"app",
	fx.Invoke(
		func(app *fiber.App, envVars *env.Env, producer redpanda.Producer, logr logging.Logger) error {
			app.Post(
				"/git/:provider", func(ctx *fiber.Ctx) error {
					gitProvider := ctx.Params("provider")
					repoUrl, err := gitRepoUrl(gitProvider, ctx.Body())
					if err != nil {
						return err
					}
					logger := logr.WithName("git-webhook").WithKV("provider", gitProvider, "repo", repoUrl)
					logger.Infof("received webhook")
					p, err := json.Marshal(
						GitWebhookPayload{
							Provider:   gitProvider,
							Body:       ctx.Body(),
							ReqHeaders: ctx.GetReqHeaders(),
						},
					)
					if err != nil {
						return err
					}

					msg, err := producer.Produce(ctx.Context(), envVars.GitWebhooksTopic, repoUrl, p)
					if err != nil {
						errMsg := fmt.Sprintf("could not produce message to topic %s", gitProvider)
						logger.Errorf(err, errMsg)
						return ctx.Status(http.StatusInternalServerError).JSON(errMsg)
					}
					logger = logger.WithKV(
						"offset:", msg.Offset,
						"topic", msg.Topic,
						"timestamp", msg.Timestamp,
					)
					logger.Infof("queued webhook")
					return ctx.Status(http.StatusAccepted).JSON(msg)
				},
			)
			return nil
		},
	),
)
