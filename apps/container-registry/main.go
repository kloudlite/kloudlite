package main

import (
	"context"
	"flag"
	"fmt"

	"github.com/kloudlite/operator/pkg/kubectl"
	"go.uber.org/fx"
	"k8s.io/client-go/rest"
	"kloudlite.io/apps/container-registry/internal/env"
	"kloudlite.io/apps/container-registry/internal/framework"
	fn "kloudlite.io/pkg/functions"
	"kloudlite.io/pkg/k8s"
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

		fx.Provide(func(restCfg *rest.Config) (kubectl.YAMLClient, error) {
			return kubectl.NewYAMLClient(restCfg)
		}),

		fx.Provide(func(restCfg *rest.Config) (k8s.ExtendedK8sClient, error) {
			return k8s.NewExtendedK8sClient(restCfg)
		}),

		fn.FxErrorHandler(),
		framework.Module,
	)

	// ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	// defer cancel()
	if err := app.Start(context.TODO()); err != nil {
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
