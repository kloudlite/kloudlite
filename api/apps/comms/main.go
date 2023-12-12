package main

import (
	"context"
	"embed"
	"flag"
	"os"
	"time"

	"github.com/kloudlite/api/apps/comms/internal/app"
	"github.com/kloudlite/api/apps/comms/internal/env"
	"github.com/kloudlite/api/apps/comms/internal/framework"
	"github.com/kloudlite/api/common"
	"github.com/kloudlite/api/pkg/logging"
	"go.uber.org/fx"
)

//go:embed email-templates
var EmailTemplatesDir embed.FS

func main() {
	var isDev bool
	flag.BoolVar(&isDev, "dev", false, "--dev")
	flag.Parse()

	logger, err := logging.New(&logging.Options{Name: "comms", Dev: isDev})
	if err != nil {
		panic(err)
	}

	app := fx.New(
		fx.NopLogger,
		fx.Provide(
			func() logging.Logger {
				return logger
			},
		),
		fx.Provide(func() (*env.Env, error) {
			return env.LoadEnv()
		}),
		fx.Provide(func() app.EmailTemplatesDir {
			return app.EmailTemplatesDir{
				FS: EmailTemplatesDir,
			}
		}),
		framework.Module,
	)

	ctx, cf := context.WithTimeout(context.Background(), 5*time.Second)
	defer cf()

	if err := app.Start(ctx); err != nil {
		logger.Errorf(err, "comms-api startup errors")
		logger.Infof("EXITING as errors encountered during startup")
		os.Exit(1)
	}

	common.PrintReadyBanner()
	<-app.Done()
}
