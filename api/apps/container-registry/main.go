package main

import (
	"context"
	"flag"
	"os"
	"os/signal"
	"time"

	"github.com/kloudlite/api/apps/container-registry/internal/env"
	"github.com/kloudlite/api/pkg/errors"

	"github.com/kloudlite/api/apps/container-registry/internal/framework"
	"github.com/kloudlite/api/common"
	fn "github.com/kloudlite/api/pkg/functions"
	"github.com/kloudlite/api/pkg/logging"
	"go.uber.org/fx"
)

func main() {
	start := time.Now()

	var isDev bool
	flag.BoolVar(&isDev, "dev", false, "--dev")

	var debug bool
	flag.BoolVar(&debug, "debug", false, "--debug")

	flag.Parse()

	if isDev {
		debug = true
	}

	logger := logging.NewSlogLogger(logging.SlogOptions{ShowCaller: true, ShowDebugLogs: debug, SetAsDefaultLogger: true})

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

		fx.Supply(logger),

		fn.FxErrorHandler(),
		framework.Module,
	)

	ctx, cf := signal.NotifyContext(context.TODO(), os.Interrupt)
	defer cf()

	if err := app.Start(ctx); err != nil {
		logger.Error("failed to start container registry api, got", "err", err)
		os.Exit(1)
	}

	common.PrintReadyBanner2(time.Since(start))
	<-app.Done()
}
