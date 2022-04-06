package fiber_app

import (
	"github.com/gofiber/fiber/v2"
)

func NewFiberApp() (server *fiber.App, e error) {
	defer errors.HandleErr(&e)
	server = fiber.New()
	server.Get("/healthy", func(c *fiber.Ctx) error {
		return c.Status(200).Send([]byte("OK"))
	})
	return server, e
}
