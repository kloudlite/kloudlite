package main

import (
	"flag"

	"go.uber.org/fx"
	"kloudlite.io/apps/nodecontroller/internal/framework"
	"kloudlite.io/pkg/logging"
)

func main() {
	var isDev bool
	flag.BoolVar(&isDev, "dev", false, "--dev")
	flag.Parse()
	fx.New(
		framework.Module,
		fx.Provide(
			func() (logging.Logger, error) {
				return logging.New(&logging.Options{Name: "auth", Dev: isDev})
			},
		),
		// fx.NopLogger,
	).Run()
}
