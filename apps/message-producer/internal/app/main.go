package app

import (
	"fmt"
	"go.uber.org/fx"
	"kloudlite.io/pkg/messaging"

	"github.com/gofiber/fiber/v2"
	"kloudlite.io/pkg/errors"
)

var Module = fx.Module(
	"app",
	// Setup Router
	fx.Invoke(func(fiberServer *fiber.App, mc messaging.Producer) {
		router := fiberServer.Group("/")
		router.Get("/test", func(c *fiber.Ctx) error {
			body := messaging.Json{
				"action":       "create",
				"resourceType": "project",
				"metadata": map[string]string{
					"projectId": "proj-oe40wrvzed6ea86xkedrdk9w3ppkow10-kl",
				},
			}
			e := mc.SendMessage("hotspot-new-testing", "test-key", body)
			if e != nil {
				return c.Status(500).SendString(e.Error())
			}
			c.JSON(body)
			return nil
		})
		router.Get("/test-config", func(c *fiber.Ctx) error {
			body := messaging.Json{
				"action":       "create",
				"resourceType": "config",
				"projectId":    "proj-oe40wrvzed6ea86xkedrdk9w3ppkow10-kl",
				"metadata": map[string]string{
					"name":     "my-real-config-1",
					"configId": "cfg-cxi2ebhhnfpazewik06pkbwi0wh4w7iw-kl",
				},
			}
			e := mc.SendMessage("hotspot-new-testing", "test-key", body)
			if e != nil {
				return c.Status(500).SendString(e.Error())
			}
			c.JSON(body)
			return nil
		})
		router.Post("/", func(c *fiber.Ctx) (e error) {
			defer errors.HandleErr(&e)
			fmt.Println("POST received")
			payload := struct {
				Topic   string                 `json:"topic"`
				Key     string                 `json:"key"`
				Message map[string]interface{} `json:"message"`
			}{}

			e = c.BodyParser(&payload)
			fmt.Println("PAYLOAD received: ", payload)
			errors.AssertNoError(e, fmt.Errorf("could not parse POST body"))

			e = mc.SendMessage(payload.Topic, payload.Key, payload.Message)
			errors.AssertNoError(e, fmt.Errorf("failed to push message to kafka queue"))
			fmt.Println("")
			e = c.JSON(payload)
			errors.AssertNoError(e, fmt.Errorf("failed to send response"))
			return e
		})
	}),
)
