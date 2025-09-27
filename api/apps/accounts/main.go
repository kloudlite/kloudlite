package main

import (
	"context"
	"flag"
	"os"
	"time"

	fx_app "github.com/kloudlite/api/apps/accounts/fx-app"
	"github.com/kloudlite/api/common"
	"github.com/kloudlite/api/pkg/logging"
	"go.uber.org/fx"
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
		fx.Supply(logger),
		fx_app.NewAccountsModule(),
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
