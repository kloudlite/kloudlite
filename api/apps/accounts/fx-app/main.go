package fx_app

import (
	"github.com/kloudlite/api/apps/accounts/internal/app"
	"github.com/kloudlite/api/apps/accounts/internal/env"
	"github.com/kloudlite/api/pkg/errors"
	"github.com/kloudlite/api/pkg/k8s"
	"go.uber.org/fx"
	"k8s.io/client-go/rest"
)

func NewAccountsModule() fx.Option {
	accountsApp := fx.Module(
		"accounts:app",
		fx.Provide(func() (*env.AccountsEnv, error) {
			if e, err := env.LoadEnv(); err != nil {
				return nil, errors.NewE(err)
			} else {
				return e, nil
			}
		}),
		fx.Module(
			"kube-access",
			fx.Provide(func(e *env.AccountsEnv) (*rest.Config, error) {
				if e.KubernetesApiProxy != "" {
					return &rest.Config{
						Host: e.KubernetesApiProxy,
					}, nil
				}
				return k8s.RestInclusterConfig()
			}),
			fx.Provide(func(cfg *rest.Config) (k8s.Client, error) {
				return k8s.NewClient(cfg, nil)
			}),
		),

		app.Module,
	)
	return accountsApp
}
