package main

import (
	"context"
	"flag"
	"os"
	"time"

	"github.com/kloudlite/api/apps/infra/internal/env"
	"github.com/kloudlite/api/apps/infra/internal/framework"
	"github.com/kloudlite/api/common"
	"k8s.io/client-go/rest"

	"github.com/kloudlite/api/pkg/k8s"
	"github.com/kloudlite/api/pkg/logging"
	crdsv1 "github.com/kloudlite/operator/apis/crds/v1"
	"go.uber.org/fx"
	"k8s.io/apimachinery/pkg/runtime"
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
			if e, err := env.LoadEnv(); err != nil {
				return nil, err
			} else {
				e.IsDev = isDev
				return e, nil
			}
		}),

		fx.Provide(func(e *env.Env) (*rest.Config, error) {
			if e.KubernetesApiProxy != "" {
				return &rest.Config{
					Host: e.KubernetesApiProxy,
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
