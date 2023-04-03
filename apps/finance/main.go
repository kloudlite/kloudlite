package main

import (
	"flag"

	"github.com/kloudlite/operator/pkg/kubectl"
	"go.uber.org/fx"
	"k8s.io/client-go/rest"
	"kloudlite.io/apps/finance/internal/env"
	"kloudlite.io/apps/finance/internal/framework"
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
				return logging.New(&logging.Options{Name: "finance", Dev: isDev})
			},
		),

		fx.Provide(func() (*env.Env, error) {
			return env.LoadEnvOrDie()
		}),

		fx.Provide(func() (*rest.Config, error) {
			if isDev {
				return &rest.Config{Host: "localhost:8080"}, nil
			}
			return rest.InClusterConfig()
		}),
		fx.Provide(
			func(restCfg *rest.Config) (*k8s.YAMLClient, error) {
				return k8s.NewYAMLClient(restCfg)
			},
		),
		fx.Provide(func(restCfg *rest.Config) (*kubectl.YAMLClient, error) {
			return kubectl.NewYAMLClient(restCfg)
		}),
		framework.Module,
	).Run()
}
