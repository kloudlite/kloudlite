package main

import (
	"flag"

	"go.uber.org/fx"

	"kloudlite.io/apps/nodectrl/internal/env"
	"kloudlite.io/apps/nodectrl/internal/framework"
	"kloudlite.io/pkg/logging"
)

func main() {
	var isDev bool
	flag.BoolVar(&isDev, "dev", false, "--dev")
	flag.Parse()
	fx.New(
		// fn.FxErrorHandler(),
		// fx.NopLogger,
		fx.Provide(env.LoadEnv),
		fx.Provide(
			func() (logging.Logger, error) {
				return logging.New(&logging.Options{Name: "nodectrl", Dev: isDev})
			},
		),
		framework.Module,
	).Run()
}
