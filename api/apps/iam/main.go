package main

import (
	"context"
	"flag"
	"os"
	"time"

	"github.com/kloudlite/api/apps/iam/internal/env"
	"github.com/kloudlite/api/apps/iam/internal/framework"
	"github.com/kloudlite/api/common"
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

	logger := logging.NewSlogLogger(logging.SlogOptions{ShowCaller: true, SetAsDefaultLogger: true, ShowDebugLogs: debug})

	app := fx.New(
		fx.NopLogger,
		fx.Supply(logger),
		fx.Provide(func() (logging.Logger, error) {
			return logging.New(&logging.Options{Name: "iam", Dev: isDev})
		}),

		fx.Provide(func() (*env.Env, error) {
			return env.LoadEnv()
		}),

		framework.Module,
	)

	ctx, cancelFunc := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancelFunc()
	if err := app.Start(ctx); err != nil {
		logger.Error("failed to start iam api, got", "err", err)
		os.Exit(1)
	}

	common.PrintReadyBanner2(time.Since(start))
	<-app.Done()
}
