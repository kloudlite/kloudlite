package main

import (
	"flag"
	"go.uber.org/fx"
	"kloudlite.io/apps/ci/internal/framework"
	"kloudlite.io/pkg/logging"
)

func main() {
	isDev := flag.Bool("dev", false, "--dev")
	flag.Parse()

	fx.New(
		framework.Module,
		fx.Provide(
			func() (logging.Logger, error) {
				return logging.NewLogger(
					logging.Options{Name: "ci", Dev: *isDev},
				)
			},
		),
	).Run()
}
