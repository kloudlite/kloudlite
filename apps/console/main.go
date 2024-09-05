package main

import (
	"context"
	"flag"
	"log/slog"
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
	var isDev bool
	flag.BoolVar(&isDev, "dev", false, "--dev")

	var debug bool
	flag.BoolVar(&debug, "debug", false, "--debug")

	flag.Parse()

	if isDev {
		debug = true
	}

	common.PrintBuildInfo()

	logger := logging.NewSlogLogger(logging.SlogOptions{ShowCaller: true, ShowDebugLogs: debug, SetAsDefaultLogger: true})

	start := time.Now()

	app := fx.New(
		fx.StartTimeout(5*time.Second),
		fx.NopLogger,
		fx.Provide(func() (logging.Logger, error) {
			return logging.New(&logging.Options{Name: "console", ShowDebugLog: debug})
		}),

		fx.Provide(func() *slog.Logger {
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

		fx.Provide(func(e *env.Env) (*rest.Config, error) {
			if e.KubernetesApiProxy != "" {
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
			return context.WithTimeout(context.TODO(), 5*time.Second)
		}
		return context.WithTimeout(context.TODO(), 10*time.Second)
	}()
	defer cancelFunc()

	if err := app.Start(ctx); err != nil {
		logger.Error("while starting console, got", "err", err)
		os.Exit(1)
	}

	common.PrintReadyBanner2(time.Since(start))
	<-app.Done()
}
