package main

import (
	"embed"
	"flag"

	"go.uber.org/fx"
	"kloudlite.io/apps/comms/internal/app"
	"kloudlite.io/apps/comms/internal/framework"
	fn "kloudlite.io/pkg/functions"
	"kloudlite.io/pkg/logging"
)

var (
	//go:embed email-templates
	EmailTemplatesDir embed.FS
)

func main() {
	var isDev bool
	flag.BoolVar(&isDev, "dev", false, "--dev")
	flag.Parse()

	fx.New(
		fx.NopLogger,
		fx.Provide(
			func() (logging.Logger, error) {
				return logging.New(&logging.Options{Name: "comms", Dev: isDev})
			},
		),
		fn.FxErrorHandler(),
		fx.Provide(func() app.EmailTemplatesDir {
			return app.EmailTemplatesDir{
				FS: EmailTemplatesDir,
			}
		}),
		framework.Module,
	).Run()
}
