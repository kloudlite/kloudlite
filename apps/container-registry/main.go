package main

import (
	"context"
	"flag"
	"runtime/trace"

	"go.uber.org/fx"
	"kloudlite.io/apps/container-registry/internal/env"
	"kloudlite.io/apps/container-registry/internal/framework"
	"kloudlite.io/common"
	fn "kloudlite.io/pkg/functions"
	"kloudlite.io/pkg/logging"
)

func main() {
	var isDev bool
	flag.BoolVar(&isDev, "dev", false, "--dev")
	flag.Parse()

	app := fx.New(
		fx.Provide(env.LoadEnv),
		fx.NopLogger,
		fx.Provide(
			func() (logging.Logger, error) {
				return logging.New(&logging.Options{Name: "container-registry", Dev: isDev})
			},
		),

		fn.FxErrorHandler(),
		framework.Module,
	)

	if err := app.Start(context.TODO()); err != nil {
		trace.Log(context.TODO(), "app.Start", err.Error())
		panic(err)
	}

	common.PrintReadyBanner()
	<-app.Done()
}
