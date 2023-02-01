package main

import (
	"context"
	"flag"
	"fmt"
	"kloudlite.io/pkg/config"
	"time"

	"go.uber.org/fx"
	"kloudlite.io/apps/infra/internal/env"
	"kloudlite.io/apps/infra/internal/framework"
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
				return logging.New(&logging.Options{Name: "infra", Dev: isDev})
			},
		),

		fx.Provide(func() (*env.Env, error) {
			ev, err := config.LoadEnv[env.Env]()()
			if err != nil {
				return nil, err
			}
			return ev, nil
		}),
		fn.FxErrorHandler(),
		framework.Module,
	)

	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
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
