package main

import (
	"context"
	"flag"
	"os"
	"time"

	"go.uber.org/fx"
	"k8s.io/client-go/rest"

	"github.com/kloudlite/api/apps/message-office/internal/env"
	"github.com/kloudlite/api/apps/message-office/internal/framework"
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

	start := time.Now()
	common.PrintBuildInfo()

	logger := logging.NewSlogLogger(logging.SlogOptions{ShowCaller: true, ShowDebugLogs: debug, SetAsDefaultLogger: true})

	app := fx.New(
		fx.NopLogger,

		fx.Provide(func() *env.Env {
			return env.LoadEnvOrDie()
		}),

		fx.Provide(
			func() (logging.Logger, error) {
				return logging.New(&logging.Options{Name: "message-office", ShowDebugLog: isDev, HideCallerTrace: false})
			},
		),

		fx.Supply(logger),

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
		logger.Error("failed to start message-office, got", "err", err)
		os.Exit(1)
	}

	common.PrintReadyBanner2(time.Since(start))
	<-app.Done()
}
