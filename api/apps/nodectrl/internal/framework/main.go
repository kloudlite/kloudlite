package framework

import (
	"go.uber.org/fx"

	"kloudlite.io/apps/nodectrl/internal/app"
	"kloudlite.io/apps/nodectrl/internal/env"
)

type fm struct {
	env *env.Env
}

var Module = fx.Module(
	"framework",
	fx.Provide(func(env *env.Env) *fm {
		return &fm{env}
	}),
	app.Module,
)
