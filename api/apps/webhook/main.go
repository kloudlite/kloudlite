package main

import (
	"flag"
	"log/slog"

	"github.com/kloudlite/api/apps/webhook/internal/env"
	"github.com/kloudlite/api/apps/webhook/internal/framework"
	"github.com/kloudlite/api/pkg/config"
	fn "github.com/kloudlite/api/pkg/functions"
	"github.com/kloudlite/api/pkg/logging"
	"go.uber.org/fx"
)

func main() {
	var isDev bool
	flag.BoolVar(&isDev, "dev", false, "--dev")
	flag.Parse()

	fx.New(
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
	).Run()
}
