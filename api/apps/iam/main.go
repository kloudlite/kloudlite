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
	var isDev bool
	flag.BoolVar(&isDev, "dev", false, "--dev")
	flag.Parse()

	logger, err := logging.New(&logging.Options{Name: "iam", Dev: isDev})
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

	ctx, cancelFunc := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancelFunc()
	if err := app.Start(ctx); err != nil {
		logger.Errorf(err, "IAM api startup errors")
		logger.Infof("EXITING as errors encountered during startup")
		os.Exit(1)
	}

	common.PrintReadyBanner()
	<-app.Done()

}
