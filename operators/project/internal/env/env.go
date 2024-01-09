package env

import (
	"github.com/codingconcepts/env"
)

type Env struct {
	MaxConcurrentReconciles int `env:"MAX_CONCURRENT_RECONCILES"`

	SvcAccountName string `env:"SVC_ACCOUNT_NAME" required:"true"`

	WorkspaceRouteSwitcherName  string `env:"WORKSPACE_ROUTE_SWITCHER_NAME" required:"false"`
	WorkspaceRouteSwitcherImage string `env:"WORKSPACE_ROUTE_SWITCHER_IMAGE" required:"false"`

	DefaultIngressClass string `env:"DEFAULT_INGRESS_CLASS" required:"true"`
}

func GetEnvOrDie() *Env {
	var ev Env
	if err := env.Set(&ev); err != nil {
		panic(err)
	}
	if len(ev.WorkspaceRouteSwitcherName) == 0 {
		ev.WorkspaceRouteSwitcherName = "workspace-route-switcher"
	}

	if len(ev.WorkspaceRouteSwitcherImage) == 0 {
		ev.WorkspaceRouteSwitcherImage = "ghcr.io/kloudlite/operators/workspace-route-switcher:v1.0.5-nightly"
	}
	return &ev
}
