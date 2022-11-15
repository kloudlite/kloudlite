package main

import (
	"context"
	"flag"
	"fmt"
	"time"

	"go.uber.org/fx"
	"kloudlite.io/apps/slack-notifier/internal/env"
	"kloudlite.io/apps/slack-notifier/internal/framework"
	"kloudlite.io/pkg/config"
	fn "kloudlite.io/pkg/functions"
	"kloudlite.io/pkg/logging"
)

func main() {
	var isDev bool
	flag.BoolVar(&isDev, "dev", false, "--dev")
	flag.Parse()

	app := fx.New(
		fx.NopLogger,
		fx.Provide(
			func() (logging.Logger, error) {
				return logging.New(&logging.Options{Name: "slack-notifier", Dev: isDev})
			},
		),
		config.EnvFx[env.Env](),
		fx.Provide(
			func() env.DevMode {
				return env.DevMode(isDev)
			},
		),
		framework.Module,
		fn.FxErrorHandler(),
	)

	ctx, cancelFn := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancelFn()
	if err := app.Start(ctx); err != nil {
		panic(err)
	}

	fmt.Println(
		`
██████  ███████  █████  ██████  ██    ██ 
██   ██ ██      ██   ██ ██   ██  ██  ██  
██████  █████   ███████ ██   ██   ████   
██   ██ ██      ██   ██ ██   ██    ██    
██   ██ ███████ ██   ██ ██████     ██    
	`,
	)

	<-app.Done()
}
