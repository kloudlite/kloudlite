package domain

import (
	"go.uber.org/fx"

	"kloudlite.io/apps/nodectrl/internal/env"
)

type domain struct {
	env *env.Env
}

func (d domain) GetEnv() *env.Env {
	return d.env
}

var Module = fx.Module("domain",
	fx.Provide(
		func(env *env.Env) Domain {
			return domain{
				env: env,
			}
		},
	),
	ProviderClientFx,
)
