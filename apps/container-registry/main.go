package main

import (
	"context"
	"flag"
	"log/slog"
	"runtime/trace"

	"github.com/kloudlite/api/apps/container-registry/internal/env"
	"github.com/kloudlite/api/pkg/errors"

	"github.com/kloudlite/api/apps/container-registry/internal/framework"
	"github.com/kloudlite/api/common"
	fn "github.com/kloudlite/api/pkg/functions"
	"github.com/kloudlite/api/pkg/logging"
	"go.uber.org/fx"
)

func main() {
	var isDev bool
	flag.BoolVar(&isDev, "dev", false, "--dev")
	flag.Parse()

	app := fx.New(
		fx.Provide(func() (*env.Env, error) {
			if e, err := env.LoadEnv(); err != nil {
				return nil, errors.NewE(err)
			} else {
				e.IsDev = isDev
				return e, nil
			}
		}),
		fx.NopLogger,
		fx.Provide(
			func() (logging.Logger, error) {
				return logging.New(&logging.Options{Name: "container-registry", Dev: isDev})
			},
		),

		fx.Provide(func() *slog.Logger {
			return logging.NewSlogLogger(logging.SlogOptions{
				ShowCaller:         true,
				ShowDebugLogs:      isDev,
				SetAsDefaultLogger: true,
			})
		}),

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
