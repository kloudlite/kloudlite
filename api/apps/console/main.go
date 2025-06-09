package main

import (
	"context"
	"flag"
	"os"
	"time"

	"github.com/kloudlite/api/pkg/errors"

	"go.uber.org/fx"
	"k8s.io/client-go/rest"

	"github.com/kloudlite/api/apps/console/internal/env"
	"github.com/kloudlite/api/apps/console/internal/framework"
	"github.com/kloudlite/api/common"
	"github.com/kloudlite/api/pkg/k8s"
	"github.com/kloudlite/api/pkg/logging"
)

func main() {
	start := time.Now()
	common.PrintBuildInfo()

	var isDev bool
	flag.BoolVar(&isDev, "dev", false, "--dev")

	var debug bool
	flag.BoolVar(&debug, "debug", false, "--debug")

	var withWorker bool
	flag.BoolVar(&withWorker, "with-worker", false, "--with-worker")

	flag.Parse()

	if isDev {
		debug = true
	}

	logger := logging.NewSlogLogger(logging.SlogOptions{ShowCaller: true, ShowDebugLogs: debug, SetAsDefaultLogger: true})

	app := fx.New(
		fx.NopLogger,
		fx.Provide(func() (logging.Logger, error) {
			return logging.New(&logging.Options{Name: "console", ShowDebugLog: debug})
		}),

		fx.Supply(logger),

		fx.Provide(func() (*env.Env, error) {
			if e, err := env.LoadEnv(); err != nil {
				return nil, errors.NewE(err)
			} else {
				e.IsDev = isDev
				return e, nil
			}
		}),

		fx.Provide(func(e *env.Env) (*rest.Config, error) {
			if isDev {
				return &rest.Config{
					Host: e.KubernetesApiProxy,
				}, nil
			}
			return k8s.RestInclusterConfig()
		}),
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
		logger.Error("while starting console, got", "err", err)
		os.Exit(1)
	}

	common.PrintReadyBanner2(time.Since(start))
	<-app.Done()
}
