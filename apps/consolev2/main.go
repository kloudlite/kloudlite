package main

import (
	"flag"
	"go.uber.org/fx"
	"k8s.io/client-go/rest"
	"kloudlite.io/apps/consolev2/internal/env"
	"kloudlite.io/apps/consolev2/internal/framework"
	"kloudlite.io/pkg/config"
	"kloudlite.io/pkg/k8s"
	"kloudlite.io/pkg/logging"
)

func main() {
	var isDev bool
	flag.BoolVar(&isDev, "dev", false, "--dev")
	flag.Parse()

	fx.New(
		fx.Provide(
			func() (logging.Logger, error) {
				return logging.New(&logging.Options{Name: "console", Dev: isDev})
			},
		),
		config.EnvFx[env.Env](),
		fx.Provide(
			func() (*k8s.YAMLClient, error) {
				if isDev {
					return k8s.NewYAMLClient(&rest.Config{Host: "127.0.0.1:8080"})
				}
				inclusterCfg, err := rest.InClusterConfig()
				if err != nil {
					return nil, err
				}
				return k8s.NewYAMLClient(inclusterCfg)
			},
		),
		framework.Module,
	).Run()
}
