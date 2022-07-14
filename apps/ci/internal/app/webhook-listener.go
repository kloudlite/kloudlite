package app

import (
	"fmt"
	"github.com/gofiber/fiber/v2"
	"kloudlite.io/pkg/logging"
	"net/http"
	"strings"
)

func NewWebhook(app *fiber.App, basePath string, logger logging.Logger) {
	logger.Infof("Attaching new webhook handler at (root=%s)", basePath)
	makeRoute := func(ep string) string {
		if strings.HasSuffix(basePath, "/") {
			return fmt.Sprintf("%s%s", basePath[:len(basePath)-1], ep)
		}
		return fmt.Sprintf("%s%s", basePath, ep)
	}

	logger.Debugf("[webhook logger]: /healthy %s", makeRoute("/healthy"))
	app.Get(
		makeRoute("/healthy"), func(ctx *fiber.Ctx) error {
			return ctx.SendStatus(http.StatusOK)
		},
	)

	app.Post(
		makeRoute("/git/github"), func(ctx *fiber.Ctx) error {
			return ctx.SendStatus(http.StatusNotImplemented)
		},
	)

	app.Post(
		makeRoute("/git/gitlab"), func(ctx *fiber.Ctx) error {
			return ctx.SendStatus(http.StatusNotImplemented)
		},
	)
}
