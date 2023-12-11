package main

import (
	"context"
	"flag"
	"os"
	"time"

	"k8s.io/client-go/rest"
	"kloudlite.io/apps/infra/internal/env"
	"kloudlite.io/apps/infra/internal/framework"
	"kloudlite.io/common"

	crdsv1 "github.com/kloudlite/operator/apis/crds/v1"
	"go.uber.org/fx"
	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	k8sScheme "k8s.io/client-go/kubernetes/scheme"
	"kloudlite.io/pkg/k8s"
	"kloudlite.io/pkg/logging"
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
			return env.LoadEnv()
		}),

		fx.Provide(func() (*rest.Config, error) {
			if isDev {
				return &rest.Config{
					Host: "localhost:8080",
				}, nil
			}
			return k8s.RestInclusterConfig()
		}),

		fx.Provide(func(restCfg *rest.Config) (k8s.Client, error) {
			scheme := runtime.NewScheme()
			utilruntime.Must(k8sScheme.AddToScheme(scheme))
			utilruntime.Must(crdsv1.AddToScheme(scheme))

			return k8s.NewClient(restCfg, scheme)
		}),

		framework.Module,
	)

	ctx, cancel := func() (context.Context, context.CancelFunc) {
		if isDev {
			return context.WithCancel(context.TODO())
		}
		return context.WithTimeout(context.Background(), 2*time.Second)
	}()
	defer cancel()

	if err := app.Start(ctx); err != nil {
		logger.Errorf(err, "failed to start app")
		os.Exit(1)
	}

	common.PrintReadyBanner()
	<-app.Done()
}
