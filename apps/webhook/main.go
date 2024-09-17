package main

import (
	"context"
	"flag"
	"log/slog"
	"os"
	"time"

	"github.com/kloudlite/api/apps/webhook/internal/env"
	"github.com/kloudlite/api/apps/webhook/internal/framework"
	"github.com/kloudlite/api/common"
	"github.com/kloudlite/api/pkg/config"
	fn "github.com/kloudlite/api/pkg/functions"
	"github.com/kloudlite/api/pkg/logging"
	"go.uber.org/fx"
)

func main() {
	var isDev bool
	flag.BoolVar(&isDev, "dev", false, "--dev")
	flag.Parse()

	start := time.Now()
	common.PrintBuildInfo()

	logger := logging.NewSlogLogger(logging.SlogOptions{ShowCaller: true, ShowDebugLogs: isDev, SetAsDefaultLogger: true})

	app := fx.New(
		fx.NopLogger,
		fx.Provide(
			func() (logging.Logger, error) {
				return logging.New(&logging.Options{Name: "webhooks", Dev: isDev})
			},
		),

		fx.Provide(func() *slog.Logger {
			return logging.NewSlogLogger(logging.SlogOptions{
				ShowCaller:         true,
				ShowDebugLogs:      isDev,
				SetAsDefaultLogger: true,
			})
		}),

		fn.FxErrorHandler(),
		config.EnvFx[env.Env](),
		framework.Module,
	)

	ctx, cancel := func() (context.Context, context.CancelFunc) {
		if isDev {
			return context.WithTimeout(context.TODO(), 5*time.Second)
		}
		return context.WithTimeout(context.Background(), 2*time.Second)
	}()
	defer cancel()

	if err := app.Start(ctx); err != nil {
		logger.Error("failed to start webhooks-api, got", slog.String("err", err.Error()))
		os.Exit(1)
	}

	common.PrintReadyBanner2(time.Since(start))
	<-app.Done()
}
