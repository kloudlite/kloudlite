package main

import (
	"context"
	"flag"
	"fmt"
	"time"

	"go.uber.org/fx"
	"kloudlite.io/apps/console/internal/env"
	"kloudlite.io/apps/console/internal/framework"
	fn "kloudlite.io/pkg/functions"
	"kloudlite.io/pkg/logging"
)

func main() {
	var isDev bool
	flag.BoolVar(&isDev, "dev", false, "--dev")
	flag.Parse()

	app := fx.New(
		fx.Provide(env.LoadEnv),
		fx.NopLogger,
		fx.Provide(
			func() (logging.Logger, error) {
				return logging.New(&logging.Options{Name: "ci", Dev: isDev})
			},
		),
		fn.FxErrorHandler(),
		framework.Module,
	)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
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
