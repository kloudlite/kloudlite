package main

import (
	"context"
	"flag"
	"os"
	"time"

	"go.uber.org/fx"
	"k8s.io/client-go/rest"

	"github.com/kloudlite/api/apps/console/internal/env"
	"github.com/kloudlite/api/apps/console/internal/framework"
	"github.com/kloudlite/api/common"
	"github.com/kloudlite/api/pkg/k8s"
	"github.com/kloudlite/api/pkg/logging"
)

func main() {
	var isDev bool
	flag.BoolVar(&isDev, "dev", false, "--dev")
	flag.Parse()

	logger, err := logging.New(&logging.Options{Name: "console", Dev: isDev})
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

		fx.Provide(func() (*rest.Config, error) {
			if isDev {
				return &rest.Config{Host: "localhost:8080"}, nil
			}

			return k8s.RestInclusterConfig()
		}),
		framework.Module,
	)

	ctx, cancelFunc := func() (context.Context, context.CancelFunc) {
		if isDev {
			return context.WithTimeout(context.TODO(), 20*time.Second)
		}
		return context.WithTimeout(context.TODO(), 5*time.Second)
	}()
	defer cancelFunc()

	if err := app.Start(ctx); err != nil {
		logger.Errorf(err, "console startup errors")
		logger.Infof("EXITING as errors encountered during startup")
		os.Exit(1)
	}

	common.PrintReadyBanner()
	<-app.Done()
}
