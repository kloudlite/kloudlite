package app

import (
	"encoding/json"
	"github.com/gofiber/fiber/v2"
	tekton "github.com/tektoncd/triggers/pkg/apis/triggers/v1alpha1"
	"go.uber.org/fx"
	"kloudlite.io/apps/tekton-interceptor/internal/domain"
	"net/http"
)

const (
	Github string = "github"
	Gitlab string = "gitlab"
)

var Module = fx.Module(
	"app",
	fx.Invoke(
		func(app *fiber.App, d domain.Domain) {
			app.Get(
				"/healthy", func(ctx *fiber.Ctx) error {
					return ctx.SendStatus(http.StatusOK)
				},
			)

			app.Post(
				"/:git-provider", func(ctx *fiber.Ctx) error {
					body := ctx.Request().Body()
					var req tekton.InterceptorRequest
					err := json.Unmarshal(body, &req)
					if err != nil {
						return err
					}

					gitProvider := ctx.Params("git-provider")

					switch gitProvider {
					case Github:
						{
						}
					case Gitlab:
						{
						}
					}

					return nil
				},
			)
		},
	),
)
