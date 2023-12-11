package main

import (
	"context"
	"flag"
	"os"
	"time"

	"go.uber.org/fx"
	"k8s.io/client-go/rest"

	"kloudlite.io/apps/message-office/internal/env"
	"kloudlite.io/apps/message-office/internal/framework"
	"kloudlite.io/common"
	"kloudlite.io/pkg/k8s"
	"kloudlite.io/pkg/logging"
)

func main() {
	var isDev bool
	flag.BoolVar(&isDev, "dev", false, "--dev")
	flag.Parse()

	logger, err := logging.New(&logging.Options{Name: "message-office", Dev: true})
	if err != nil {
		panic(err)
	}

	app := fx.New(
		fx.NopLogger,

		fx.Provide(func() *env.Env {
			return env.LoadEnvOrDie()
		}),

		fx.Provide(
			func() logging.Logger {
				return logger
			},
		),

		fx.Provide(func() (*rest.Config, error) {
			if isDev {
				return &rest.Config{
					Host: "localhost:8080",
				}, nil
			}
			return k8s.RestInclusterConfig()
		}),

		framework.Module,
	)

	ctx, cancelFn := func() (context.Context, context.CancelFunc) {
		if isDev {
			return context.WithCancel(context.TODO())
		}
		return context.WithTimeout(context.TODO(), 5*time.Second)
	}()

	defer cancelFn()
	if err := app.Start(ctx); err != nil {
		logger.Errorf(err, "message office startup errors")
		logger.Infof("EXITING as errors encountered during startup")
		os.Exit(1)
	}

	common.PrintReadyBanner()
	<-app.Done()
}
