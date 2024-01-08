package env

import (
	"github.com/codingconcepts/env"
)

type Env struct {
	MaxConcurrentReconciles int `env:"MAX_CONCURRENT_RECONCILES"`

	WorkspaceRouteSwitcherService string `env:"WORKSPACE_ROUTE_SWITCHER_SERVICE" required:"true"`
	WorkspaceRouteSwitcherPort    uint16 `env:"WORKSPACE_ROUTE_SWITCHER_PORT" required:"true"`

	DefaultIngressClass  string `env:"DEFAULT_INGRESS_CLASS" required:"true"`
	DefaultClusterIssuer string `env:"DEFAULT_CLUSTER_ISSUER" required:"true"`
}

func GetEnvOrDie() *Env {
	var ev Env
	if err := env.Set(&ev); err != nil {
		panic(err)
	}
	if ev.MaxConcurrentReconciles == 0 {
		ev.MaxConcurrentReconciles = 5
	}
	return &ev
}
