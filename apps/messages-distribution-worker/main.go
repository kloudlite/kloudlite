package main

import (
	"context"
	"flag"
	"os"
	"time"

	"go.uber.org/fx"
	"kloudlite.io/apps/messages-distribution-worker/internal/env"
	"kloudlite.io/apps/messages-distribution-worker/internal/framework"
	"kloudlite.io/common"
	"kloudlite.io/pkg/logging"
)

func main() {
	var isDev bool
	flag.BoolVar(&isDev, "dev", false, "--dev")
	flag.Parse()

	logger, err := logging.New(&logging.Options{Name: "message-distributor", Dev: isDev})
	if err != nil {
		panic(err)
	}

	app := fx.New(
		fx.NopLogger,

		fx.Provide(func() logging.Logger {
			return logger
		}),

		fx.Provide(func() (*env.Env, error) {
			return env.LoadEnv()
		}),

		framework.Module,
	)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := app.Start(ctx); err != nil {
		logger.Errorf(err, "failed to start app")
		os.Exit(1)
	}

	common.PrintReadyBanner()
	<-app.Done()
}
