package main

import (
	"context"
	"flag"
	"os"
	"time"

	"github.com/kloudlite/api/pkg/errors"

	"go.uber.org/fx"

	"github.com/kloudlite/api/apps/auth/internal/env"
	"github.com/kloudlite/api/apps/auth/internal/framework"
	"github.com/kloudlite/api/common"
	fn "github.com/kloudlite/api/pkg/functions"
	"github.com/kloudlite/api/pkg/logging"
)

func main() {
	common.PrintBuildInfo()
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
		fx.NopLogger,
		fn.FxErrorHandler(),
		fx.Provide(func() (*env.Env, error) {
			if e, err := env.LoadEnv(); err != nil {
				return nil, errors.NewE(err)
			} else {
				e.IsDev = isDev
				return e, nil
			}
		}),
		fx.Provide(
			func() (logging.Logger, error) {
				return logging.New(&logging.Options{Name: "auth", Dev: isDev})
			},
		),

		fx.Supply(logger),

		framework.Module,
	)

	ctx, cancelFunc := func() (context.Context, context.CancelFunc) {
		if isDev {
			return context.WithTimeout(context.TODO(), 10*time.Second)
		}
		return context.WithTimeout(context.TODO(), 5*time.Second)
	}()
	defer cancelFunc()

	if err := app.Start(ctx); err != nil {
		logger.Error("failed to start auth-api, got", "err", err)
		os.Exit(1)
	}

	common.PrintReadyBanner2(time.Since(start))
	<-app.Done()
}
