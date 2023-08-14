package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"time"

	crdsv1 "github.com/kloudlite/operator/apis/crds/v1"
	"github.com/kloudlite/operator/pkg/kubectl"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/rest"
	"kloudlite.io/pkg/config"
	"kloudlite.io/pkg/k8s"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"go.uber.org/fx"
	"kloudlite.io/apps/infra/internal/env"
	"kloudlite.io/apps/infra/internal/framework"
	"kloudlite.io/pkg/logging"

	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	k8sScheme "k8s.io/client-go/kubernetes/scheme"
)

func main() {
	var isDev bool
	flag.BoolVar(&isDev, "dev", false, "--dev")
	flag.Parse()

	logger, err := logging.New(&logging.Options{Name: "infra", Dev: isDev})
	if err != nil {
		panic(err)
	}

	app := fx.New(
		fx.NopLogger,

		fx.Provide(func() logging.Logger {
			return logger
		}),

		fx.Provide(func() (*env.Env, error) {
			ev, err := config.LoadEnv[env.Env]()()
			if err != nil {
				return nil, err
			}
			return ev, nil
		}),

		fx.Provide(func() (*rest.Config, error) {
			if isDev {
				return &rest.Config{
					Host: "localhost:8080",
				}, nil
			}
			return k8s.RestInclusterConfig()
		}),

		fx.Provide(func(restCfg *rest.Config) (client.Client, error) {
			scheme := runtime.NewScheme()
			utilruntime.Must(k8sScheme.AddToScheme(scheme))
			utilruntime.Must(crdsv1.AddToScheme(scheme))

			return client.New(restCfg, client.Options{
				Scheme: scheme,
				Mapper: nil,
			})
		}),

		fx.Provide(func(restCfg *rest.Config) (*kubectl.YAMLClient, error) {
			return kubectl.NewYAMLClient(restCfg)
		}),

		fx.Provide(func(restCfg *rest.Config) (k8s.ExtendedK8sClient, error) {
			return k8s.NewExtendedK8sClient(restCfg)
		}),

		framework.Module,
	)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)

	defer cancel()
	if err := app.Start(ctx); err != nil {
		logger.Errorf(err, "failed to start app")
		os.Exit(1)
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
