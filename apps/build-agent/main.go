package main

import (
	"context"
	"flag"
	"fmt"
	"runtime/trace"

	"github.com/kloudlite/api/apps/build-agent/internal/env"
	"github.com/kloudlite/api/apps/build-agent/internal/framework"
	fn "github.com/kloudlite/api/pkg/functions"
	"github.com/kloudlite/api/pkg/k8s"
	"github.com/kloudlite/api/pkg/logging"
	"go.uber.org/fx"
	"k8s.io/client-go/rest"
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
				return logging.New(&logging.Options{Name: "console", Dev: isDev})
			},
		),
		fx.Provide(func() (*rest.Config, error) {
			if isDev {
				return &rest.Config{
					Host: "localhost:8080",
				}, nil
			}
			return k8s.RestInclusterConfig()
		}),

		fn.FxErrorHandler(),
		framework.Module,
	)

	// ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	// defer cancel()
	if err := app.Start(context.TODO()); err != nil {
		trace.Log(context.TODO(), "app.Start", err.Error())
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
