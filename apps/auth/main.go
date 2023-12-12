package main

import (
	"context"
	"flag"
	"time"

	"go.uber.org/fx"

	"github.com/kloudlite/api/apps/auth/internal/env"
	"github.com/kloudlite/api/apps/auth/internal/framework"
	"github.com/kloudlite/api/common"
	fn "github.com/kloudlite/api/pkg/functions"
	"github.com/kloudlite/api/pkg/logging"
)

func main() {
	var isDev bool
	flag.BoolVar(&isDev, "dev", false, "--dev")
	flag.Parse()
	app := fx.New(
		fx.NopLogger,
		fn.FxErrorHandler(),
		fx.Provide(func() (*env.Env, error) {
			return env.LoadEnv()
		}),
		fx.Provide(
			func() (logging.Logger, error) {
				return logging.New(&logging.Options{Name: "auth", Dev: isDev})
			},
		),
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
		panic(err)
	}

	common.PrintReadyBanner()
	<-app.Done()
}
