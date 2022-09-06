package main

import (
	"net"
	"sync"

	"github.com/gofiber/fiber/v2"
)

type Service struct {
	Name     string `json:"name"`
	Target   int    `json:"servicePort"`
	Port     int    `json:"proxyPort"`
	Listener net.Listener
	Closed   bool
}

const ()

func reloadConfig(conf []byte) error {
	return nil
}

func startApi() {
	app := fiber.New()
	app.Post("/post", func(c *fiber.Ctx) error {
		err := reloadConfig(c.Body())
		if err != nil {
			return err
		}
		c.Send([]byte("done"))
		return nil
	})
	app.Listen(":2998")
}
func main() {
	go startApi()
	err := reloadConfig(nil)
	if err != nil {
		panic(err)
	}
	var wg sync.WaitGroup
	wg.Add(1)
	wg.Wait()
}
