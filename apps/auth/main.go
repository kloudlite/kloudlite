package main

import (
	"flag"

	"go.uber.org/fx"
	"kloudlite.io/apps/auth/internal/framework"
	"kloudlite.io/pkg/logging"
)

func main() {
	isDev := flag.Bool("dev", false, "--dev")
	flag.Parse()
	fx.New(
		framework.Module,
		fx.Provide(
			func() (logging.Logger, error) {
				return logging.New(&logging.Options{Name: "auth", Dev: *isDev})
			},
		),
	).Run()
}
