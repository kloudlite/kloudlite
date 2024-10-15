package main

import (
	"context"
	"embed"
	"flag"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/kloudlite/api/apps/comms/internal/domain"
	"github.com/kloudlite/api/apps/comms/internal/env"
	"github.com/kloudlite/api/apps/comms/internal/framework"
	"github.com/kloudlite/api/common"
	"github.com/kloudlite/api/pkg/logging"
	"go.uber.org/fx"
)

//go:embed email-templates
var EmailTemplatesDir embed.FS

func main() {
	start := time.Now()
	common.PrintBuildInfo()

	var isDev bool
	flag.BoolVar(&isDev, "dev", false, "--dev")
	flag.Parse()

	logger := logging.NewSlogLogger(logging.SlogOptions{ShowCaller: true, ShowDebugLogs: isDev, SetAsDefaultLogger: true})

	app := fx.New(
		fx.NopLogger,
		fx.Provide(func() (logging.Logger, error) {
			return logging.New(&logging.Options{Name: "comms", Dev: isDev})
		}),

		fx.Supply(logger),

		fx.Provide(func() (*env.Env, error) {
			return env.LoadEnv()
		}),
		fx.Provide(func() domain.EmailTemplatesDir {
			return domain.EmailTemplatesDir{
				FS: EmailTemplatesDir,
			}
		}),
		framework.Module,
	)

	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-ch
		logger.Info("shutting down...")
		app.Stop(context.Background())
	}()

	ctx, cf := context.WithTimeout(context.Background(), 5*time.Second)
	defer cf()

	if err := app.Start(ctx); err != nil {
		logger.Error("failed to start comms api, got", "err", err)
		os.Exit(1)
	}

	common.PrintReadyBanner2(time.Since(start))
	<-app.Done()
}
