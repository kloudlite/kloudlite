package framework

import (
	"fmt"

	"github.com/gofiber/fiber/v2"
	appsvc "kloudlite.io/apps/message-producer/internal/app"
	"kloudlite.io/pkg/errors"
)

func MakeFW(c *Config) (fm FW, e error) {
	defer errors.HandleErr(&e)
	server := fiber.New()
	mc, e := makeKafkaMessagingClient(c.KafkaBrokers)
	app := appsvc.MakeApp(mc)
	app.Init(server)
	fm = func() {
		err := server.Listen(fmt.Sprintf(":%d", c.HttpPort))
		panic(err)
	}
	return
}
