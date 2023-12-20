package main

import (
	"context"
	"flag"
	"os"
	"time"

	"github.com/kloudlite/api/common"
	"github.com/kloudlite/api/pkg/k8s"
	"github.com/kloudlite/api/pkg/logging"
	"go.uber.org/fx"
	"k8s.io/client-go/rest"

	"github.com/kloudlite/api/apps/accounts/internal/env"
	"github.com/kloudlite/api/apps/accounts/internal/framework"
)

func main() {
	var isDev bool
	flag.BoolVar(&isDev, "dev", false, "--dev")
	flag.Parse()

	logger, err := logging.New(&logging.Options{Name: "accounts", Dev: isDev})
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
				return nil, err
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

		fx.Provide(func(cfg *rest.Config) (k8s.Client, error) {
			return k8s.NewClient(cfg, nil)
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
		logger.Errorf(err, "error starting accounts app")
		logger.Infof("EXITING as errors encountered during startup")
		os.Exit(1)
	}

	common.PrintReadyBanner()
	<-app.Done()
}
