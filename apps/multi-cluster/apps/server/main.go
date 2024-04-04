package server

import (
	"github.com/gofiber/fiber/v2"
	"github.com/kloudlite/operator/apps/multi-cluster/apps/server/env"
	"github.com/kloudlite/operator/apps/multi-cluster/mpkg/wg"
	"github.com/kloudlite/operator/pkg/logging"
)

const (
	TempConfigPath = "./bin/server-config.json"
)

func Run() error {
	env := env.GetEnvOrDie()

	app := fiber.New(fiber.Config{
		DisableStartupMessage: true,
		AppName:               "multi-cluster",
	})

	l, err := logging.New(&logging.Options{})
	if err != nil {
		return err
	}

	c, err := wg.NewClient()
	if err != nil {
		return err
	}

	mserver := server{
		client: c,
		logger: l,
		app:    app,
		env:    env,
	}

	if err := mserver.Start(); err != nil {
		return err
	}

	l.WithName("gatew").Infof("listening on addr %s", env.Addr)
	if err := app.Listen(env.Addr); err != nil {
		return err
	}

	return nil
}
