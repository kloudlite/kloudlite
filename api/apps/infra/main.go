package main

import (
	"context"
	"flag"
	"os"
	"time"

	"github.com/kloudlite/api/pkg/errors"
	"github.com/kloudlite/api/pkg/helm"
	clustersv1 "github.com/kloudlite/operator/apis/clusters/v1"

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
	start := time.Now()
	common.PrintBuildInfo()

	var isDev bool
	flag.BoolVar(&isDev, "dev", false, "--dev")

	var debug bool
	flag.BoolVar(&debug, "debug", false, "--debug")

	flag.Parse()

	if isDev {
		debug = true
	}

	logger := logging.NewSlogLogger(logging.SlogOptions{ShowCaller: true, ShowDebugLogs: debug, SetAsDefaultLogger: true})

	app := fx.New(
		fx.NopLogger,

		fx.Provide(func() (logging.Logger, error) {
			return logging.New(&logging.Options{Name: "infra", Dev: isDev})
		}),

		fx.Supply(logger),

		fx.Provide(func() (*env.Env, error) {
			e, err := env.LoadEnv()
			if err != nil {
				return nil, errors.NewE(err)
			}

			e.IsDev = isDev
			return e, nil
		}),

		fx.Provide(func(e *env.Env) (*rest.Config, error) {
			if isDev && e.KubernetesApiProxy != "" {
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
			utilruntime.Must(clustersv1.AddToScheme(scheme))

			return k8s.NewClient(restCfg, scheme)
		}),

		fx.Provide(func(restCfg *rest.Config) (helm.Client, error) {
			client, err := helm.NewHelmClient(restCfg, helm.ClientOptions{
				RepositoryCacheDir: "/tmp",
			})
			if err != nil {
				return nil, err
			}

			if err := client.AddOrUpdateChartRepo(context.TODO(), helm.RepoEntry{
				Name: "kloudlite",
				URL:  "https://kloudlite.github.io/helm-charts",
			}); err != nil {
				return nil, err
			}

			return client, nil
		}),

		framework.Module,
	)

	ctx, cancel := func() (context.Context, context.CancelFunc) {
		if isDev {
			return context.WithTimeout(context.TODO(), 10*time.Second)
		}
		return context.WithTimeout(context.Background(), 2*time.Second)
	}()
	defer cancel()

	if err := app.Start(ctx); err != nil {
		logger.Error("failed to start infra api, got", "err", err)
		os.Exit(1)
	}

	common.PrintReadyBanner2(time.Since(start))
	<-app.Done()
}
