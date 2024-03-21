package main

import (
	"context"
	"flag"
	"os"
	"time"

	"github.com/kloudlite/api/apps/observability/internal/env"
	"github.com/kloudlite/api/apps/observability/internal/framework"
	"github.com/kloudlite/api/common"
	"github.com/kloudlite/api/pkg/errors"
	"github.com/kloudlite/api/pkg/logging"
	"go.uber.org/fx"
)

func main() {
	var isDev bool
	flag.BoolVar(&isDev, "dev", false, "--dev")
	flag.Parse()

	ev, err := env.LoadEnv()
	if err != nil {
		panic(err)
	}

	ev.IsDev = isDev

	logger, err := logging.New(&logging.Options{Name: "observability", Dev: isDev})
	if err != nil {
		panic(err)
	}

	app := fx.New(
		fx.NopLogger,

		fx.Provide(func() logging.Logger {
			return logger
		}),

		fx.Provide(func() (*env.Env, error) {
			if e, err := env.LoadEnv(); err != nil {
				return nil, errors.NewE(err)
			} else {
				e.IsDev = isDev
				return e, nil
			}
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
		logger.Errorf(err, "observability api startup errors")
		os.Exit(1)
	}

	common.PrintReadyBanner()
	<-app.Done()
}
