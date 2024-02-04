package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/kloudlite/api/apps/worker-audit-logging/internal/env"
	"github.com/kloudlite/api/apps/worker-audit-logging/internal/framework"
	"github.com/kloudlite/api/pkg/config"
	fn "github.com/kloudlite/api/pkg/functions"
	"github.com/kloudlite/api/pkg/logging"
	"go.uber.org/fx"
	"time"
)

func main() {
	var isDev bool
	flag.BoolVar(&isDev, "dev", false, "--dev")
	flag.Parse()

	app := fx.New(
		func() fx.Option {
			if isDev {
				return fx.Options()
			}
			return fx.NopLogger
		}(),
		fx.Provide(
			func() (logging.Logger, error) {
				return logging.New(&logging.Options{Name: "audit-logging-worker", Dev: isDev})
			},
		),
		fn.FxErrorHandler(),
		config.EnvFx[env.Env](),
		framework.Module,
	)

	ctx, cancelFn := context.WithTimeout(context.Background(), 5*time.Second)
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
