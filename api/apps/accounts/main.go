package main

import (
	"context"
	"flag"
	"os"
	"time"

	"github.com/kloudlite/api/pkg/errors"

	"github.com/kloudlite/api/common"
	"github.com/kloudlite/api/pkg/k8s"
	"github.com/kloudlite/api/pkg/logging"
	"go.uber.org/fx"
	"k8s.io/client-go/rest"

	"github.com/kloudlite/api/apps/accounts/internal/env"
	"github.com/kloudlite/api/apps/accounts/internal/framework"
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

	logger := logging.NewSlogLogger(logging.SlogOptions{
		ShowCaller:         true,
		ShowDebugLogs:      debug,
		SetAsDefaultLogger: true,
	})

	app := fx.New(
		fx.NopLogger,

		fx.Provide(func() (logging.Logger, error) {
			return logging.New(&logging.Options{Name: "accounts-api", Dev: isDev, ShowDebugLog: debug})
		}),

		fx.Supply(logger),

		fx.Provide(func() (*env.Env, error) {
			e, err := env.LoadEnv()
			if err != nil {
				return nil, errors.NewE(err)
			}
			e.IsDev = isDev
			return e, nil
		}),

		fx.Provide(func(e *env.Env) (*rest.Config, error) {
			if e.KubernetesApiProxy != "" {
				return &rest.Config{
					Host: e.KubernetesApiProxy,
				}, nil
			}

			return k8s.RestInclusterConfig()
		}),

		fx.Provide(func(cfg *rest.Config) (k8s.Client, error) {
			return k8s.NewClient(cfg, nil)
		}),

		framework.Module,
	)

	ctx, cancelFunc := func() (context.Context, context.CancelFunc) {
		if isDev {
			return context.WithTimeout(context.TODO(), 20*time.Second)
		}
		return context.WithTimeout(context.TODO(), 10*time.Second)
	}()
	defer cancelFunc()

	if err := app.Start(ctx); err != nil {
		logger.Error("failed to start accounts api, got", "err", err)
		os.Exit(1)
	}

	common.PrintReadyBanner2(time.Since(start))
	<-app.Done()
}
